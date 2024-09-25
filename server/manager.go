package main

import (
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/weifan01/websocket-example/common"

	"github.com/gorilla/websocket"
)

// Manager 用于处理客户端常规http请求，并将其升级为websocket连接，manager能够跟踪管理所有的客户端
// pic[example](https://programmingpercy.tech/_app/immutable/assets/img5-be7c5c24.webp)

var (
	// websocketUpgrader 用于将传入的HTTP请求升级为持久的websocket连接
	websocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

// Manager 用于管理所有客户端连接的注册、广播等
type Manager struct {
	clients ClientList
	// 更改 clients 的锁，也可以用 channel
	sync.RWMutex
	// handlers 处理事件的处理器，以消息类型为key，EventHandler 为value
	handlers map[string]EventHandler
}

// NewManager 初始化Manager
func NewManager() *Manager {
	m := &Manager{
		clients:  make(ClientList),
		handlers: make(map[string]EventHandler),
	}
	m.setupEventHandlers()
	return m
}

// serveWS ws连接入口方法
func (m *Manager) serveWS(w http.ResponseWriter, r *http.Request) {

	log.Println("New connection")
	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := NewClient(conn, m)
	m.addClient(client)

	go client.readMessages()
	go client.writeMessages()
}

// addClient 添加 clients 到 clientList
func (m *Manager) addClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	m.clients[client] = true
}

// removeClient 清理连接对象
func (m *Manager) removeClient(client *Client) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.clients[client]; ok {
		log.Println("remove connection", client.connection.RemoteAddr())
		client.connection.Close()
		delete(m.clients, client)
	}
}

var (
	ErrEventNotSupported = errors.New("this event type is not supported")
)

// setupEventHandlers 注册所有的 handlers
func (m *Manager) setupEventHandlers() {
	m.handlers[common.TEST] = SendMessageHandler
}

// routeEvent 合理的事件用对应的handler处理
func (m *Manager) routeEvent(event common.Event, c *Client) error {
	if handler, ok := m.handlers[event.Type]; ok {
		if err := handler(event, c); err != nil {
			return err
		}
		return nil
	} else {
		return ErrEventNotSupported
	}
}
