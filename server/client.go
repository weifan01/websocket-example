package main

import (
	"encoding/json"
	"log"

	"github.com/weifan01/websocket-example/common"

	"github.com/gorilla/websocket"
)

// ClientList 可以用来查找client的map，每个client都有一个manager的引用，用来管理client
type ClientList map[*Client]bool

// Client websocket 客户端对象，所有客户端相关逻辑在此基础上实现，能够发送和接收消息，且能够由manager管理起来
// 容易忽略的一点是websocket连接不能并发写，可以用无缓冲通道
type Client struct {
	connection *websocket.Conn
	// 管理客户端的 manager
	manager *Manager
	// egress 用通道来避免在websocket连接上并发写
	egress chan []byte
}

// NewClient 初始化新的客户端连接对象
func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		connection: conn,
		manager:    manager,
		egress:     make(chan []byte),
	}
}

// readMessages 读消息并处理，以 goroutine 运行
func (c *Client) readMessages() {
	defer func() {
		// 该方法执行完的话平滑关闭并移除连接
		c.manager.removeClient(c)
	}()

	// 消息最大大小，单位Bytes，超过最大值则连接自动关闭
	c.connection.SetReadLimit(819200)

	for {
		// 读取到客户端发送来的CloseMessage时关闭并清除连接对象
		_, event, err := c.connection.ReadMessage()
		if err != nil {
			// 如果从已关闭的连接中读取消息则err不为空
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
				log.Println("The connection was closed due to an exception, ", err)
				return
			}
			log.Println("connection has closed normally, read goroutine has exited, ", err)
			return
		}

		var request common.Event
		if err := json.Unmarshal(event, &request); err != nil {
			log.Printf("error marshalling message: %v", err)
			continue
		}

		if err := c.manager.routeEvent(request, c); err != nil {
			log.Println("Error handling Message: ", err)
		}
	}
}

// writeMessages 监听新事件并发送给 Client
func (c *Client) writeMessages() {
	defer func() {
		// Graceful close if this triggers a closing
		c.manager.removeClient(c)
	}()

	for {
		select {
		case message, ok := <-c.egress:
			// 如果通道关闭，则ok为false
			if !ok {
				// Manager 关闭了channel，则给客户端发送 CloseMessage
				if err := c.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Println("connection closed: ", err)
				}
				// 停止协程并清理该连接对象
				log.Println("write chan has closed normally. write goroutine has exited.")
				return
			}

			if err := c.connection.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Println(err)
				// todo 处理失败的消息要放回去重新发送
				continue
			}
			log.Println("server sent message")
		}
	}
}
