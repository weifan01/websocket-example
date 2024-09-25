package main

import (
	"log"
	"net/http"

	"github.com/weifan01/websocket-example/common"
)

func main() {
	log.Println("Current version: ", common.Version)
	setupAPI()
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func setupAPI() {
	// 新建manager管理ws连接
	manager := NewManager()

	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.HandleFunc("/ws", manager.serveWS)
}
