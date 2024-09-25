package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/weifan01/websocket-example/common"
)

// EventHandler 根据type路由到EventHandler函数
type EventHandler func(event common.Event, c *Client) error

// SendMessageHandler 将收到的消息发送给其他客户端
func SendMessageHandler(event common.Event, c *Client) error {
	var chatEvent common.TestMessageEvent
	if err := json.Unmarshal(event.Payload, &chatEvent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}
	fmt.Printf("server received from client, body: %s\n", chatEvent.Message)

	var broadMessage common.TestMessageEvent
	broadMessage.Sent = time.Now()
	broadMessage.From = chatEvent.From
	broadMessage.Message = "test message from server."

	data, err := json.Marshal(broadMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal broadcast message: %v", err)
	}

	// 填充 Event
	var outgoingEvent common.Event
	outgoingEvent.Payload = data
	outgoingEvent.Type = common.TEST
	d, _ := json.Marshal(outgoingEvent)
	// 广播
	for client := range c.manager.clients {
		client.egress <- d
	}

	return nil

}
