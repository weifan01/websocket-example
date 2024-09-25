package common

import (
	"encoding/json"
	"time"
)

const (
	TEST = "test"
)

// Event 最终所有要发送的事件对象都封装为 Event
type Event struct {
	// Type 消息类型
	Type string `json:"type"`
	// Payload 消息内容
	Payload json.RawMessage `json:"payload"`
}

// TestMessage 测试消息 payload
type TestMessage struct {
	Message string `json:"message"`
	From    string `json:"from"`
}
type TestMessageEvent struct {
	TestMessage
	Sent time.Time `json:"sent"`
}
