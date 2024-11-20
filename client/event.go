package main

import (
	"errors"
	"log"

	"github.com/weifan01/websocket-example/common"
)

// EventHandler 事件处理器定义
type EventHandler func(event common.Event, c *Client) error

// Handlers 事件类型为key，处理器为value的map，根据事件类型路由到EventHandler函数处理
type Handlers map[string]EventHandler

func NewHandlers() Handlers {
	return make(Handlers)
}

var (
	ErrEventNotSupported = errors.New("this event type is not supported")
)

// setupEventHandlers 注册所有的 Handlers
func (h Handlers) setupEventHandlers() {
	h[common.TEST] = func(e common.Event, c *Client) error {
		var d []byte
		_ = e.Payload.UnmarshalJSON(d)
		log.Println("received message from server, ", e.Type, string(d))
		return nil
	}
	// h[EventSendMessage] = SendMessageHandler
}

// routeEvent 合理的事件用对应的handler处理
func (h Handlers) routeEvent(event common.Event, c *Client) {
	if handler, ok := h[event.Type]; ok {
		if err := handler(event, c); err != nil {
			log.Println("error handling event:", err)
		}
	} else {
		log.Println(ErrEventNotSupported)
	}
}
