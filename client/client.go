package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/weifan01/websocket-example/common"

	"github.com/gorilla/websocket"
)

/*
Client 连接对象，用来收发消息
*/

type Client struct {
	connection *websocket.Conn
	// 连接标识，以便server定向发送消息
	id string
	// 发消息无缓冲通道，避免并发写
	egress       chan []byte
	normalClosed bool
}

// NewClient 初始化 Client 对象
func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		connection: conn,
		id:         os.Getenv("environment"),
		egress:     make(chan []byte),
	}
}

func (c *Client) run() {
	go c.readMessages()
	go c.writeMessages()

	// 启动发送测试消息
	go c.produceMessage()
}

var (
	// pongWait 响应pong包之间的超时时间
	pongWait = 15 * time.Second
	// 必须要小于pongWait时间，否则响应包未收到就已经超时了
	pingInterval = (pongWait * 8) / 10
	readMaxSize  = int64(102400)
)

// pongHandler 处理来自客户端的 PongMessages，重置deadline
func (c *Client) pongHandler(pongMsg string) error {
	// log.Println("received pong")
	return c.connection.SetReadDeadline(time.Now().Add(pongWait))
}

func (c *Client) writeMessages() {
	// ticker 同方法发送ping，以避免对ws连接的并发写，比单独定义事件类型实现更简单
	ticker := time.NewTicker(pingInterval)
	defer func() {
		c.connection.Close()
		ticker.Stop()
	}()
	for {
		select {
		case message, ok := <-c.egress:
			if !ok { // 主动关闭，通道关闭则ok为false
				log.Println("write chan has closed normally. write goroutine will be exited.")
				c.normalClosed = true

				time.Sleep(time.Second * 10) // 10s后关闭socket连接，用于读协程处理尚未收到的数据，此值应小于ping_interval，以免连接底层检测到连接断开后提前退出read协程
				// 给服务端发送 CloseMessage
				if err := c.connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
					log.Println("connection closed: ", err)
				}
				log.Println("write goroutine has exited.")
				return // 退出协程
			}

			if err := c.connection.WriteMessage(websocket.TextMessage, message); err != nil {
				// todo 消息写失败，存入FIFO队列等待重新发送消息
				log.Println("send message failed, ", err)
				continue
			}
			log.Println("event has sent by client.")
		case <-ticker.C:
			if err := c.connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println("send ping err: ", err)
			}
		}
	}
}

func (c *Client) readMessages() {
	defer func() {
		c.connection.Close()
	}()

	// 处理pong消息，未收到pong消息表示连接状态已被破坏，读取消息会返回错误，此时客户端需要重连
	if err := c.connection.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Println(err)
		return
	}
	c.connection.SetPongHandler(c.pongHandler)
	c.connection.SetReadLimit(readMaxSize)

	for {
		_, payload, err := c.connection.ReadMessage()
		// 如果从已关闭的连接中读取消息则err不为空
		if err != nil {
			if c.normalClosed {
				log.Println("connection has closed normally, read goroutine has exited")
				return
			}
			// if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			// 	log.Printf("The connection was closed due to an exception: %v", err)
			// }
			log.Println("The connection was closed due to an exception, ", err)
			c.reconnect()
			continue
		}

		// 序列化消息内容为 Event 对象
		var request common.Event
		if err := json.Unmarshal(payload, &request); err != nil {
			log.Printf("error marshalling message: %v", err)
			continue
		}

		go router.routeEvent(request, c)

	}
}

func (c *Client) produceMessage() {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	count := 1
	// 用来结束生产者
	done := make(chan bool)
	go func() {
		time.Sleep(time.Second * 15)
		done <- true
	}()

	for {
		select {
		case <-done:
			fmt.Println("produce test done")
			close(c.egress) // test if or not goroutine is exited normally
			return
		case t := <-ticker.C:
			log.Printf("produce a message, current time: %d, count: %d\n", t.Second(), count)
			testMessage := common.TestMessageEvent{
				TestMessage: common.TestMessage{
					Message: fmt.Sprintf("This is a test message, count: %d", count),
					From:    "dev",
				},
				Sent: t,
			}
			x, _ := json.Marshal(testMessage)
			event := common.Event{
				Type:    common.TEST,
				Payload: x,
			}
			d, _ := json.Marshal(event)
			c.egress <- d
			count++
		}
	}
}

// reconnect 重新连接服务端
func (c *Client) reconnect() {
	log.Println("reconnect to server")
	client := connect()
	c.connection = client
	// c.run()
}

func connect() *websocket.Conn {
	// 加一个标识，server端用来做简单验证
	var header = make(http.Header)
	header.Add("identification", os.Getenv("identification"))

	count := 0
	interval := 3
	attempt := 7 * 24 * 60 * 60 / interval // 重试最多一周
	for {
		conn, resp, err := websocket.DefaultDialer.Dial(os.Getenv("serverAddress"), header)
		if resp != nil {
			log.Println("response code: ", resp.StatusCode)
		}
		if err == nil {
			return conn
		}
		log.Println("error connect to server, ", err.Error())

		count++
		if count > attempt {
			// 关闭程序
			log.Println("exceeded retry limit")
			os.Exit(1)
		}
		log.Println("retry... ", count)

		time.Sleep(time.Second * time.Duration(interval))
	}
}
