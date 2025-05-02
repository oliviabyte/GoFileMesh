// main.go is the entry point of the GoFileMesh system

package main

import (
	"fmt"
	"net"
	"time"

	"github.com/heyxtanya/GoFileMesh/p2p"
)

func main() {
	// 创建了一个独立的 goroutine
	// 主线程继续执行，不会被 select {} 阻塞
	go startServer(":3000")
	time.Sleep(1 * time.Second) // 等待 server 启动
	sendPing("localhost:3000")
	select {} // 保持运行
}

func startServer(addr string) {
	t := p2p.NewTCPTransport(addr)

	// ✅ 在“监听的那个 server”上注册 handler
	t.OnMessage("PING", func(msg p2p.Message) {
		fmt.Println("🎯 Received PING:", msg.Data)
	})

	// 注册处理逻辑
	t.OnMessage("STORE_FILE", func(msg p2p.Message) {
		fmt.Println("📥 Received file content to store.")
		err := Store(msg.Data)
		if err != nil {
			fmt.Println("❌ Store error:", err)
		}
	})

	if err := t.Start(); err != nil {
		panic(err)
	}
}

func sendPing(addr string) {
	fmt.Println("⚡️sendPing started")
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("Dial error:", err)
		return
	}
	defer conn.Close()

	msg := p2p.Message{
		Type: "STORE_FILE",
		Data: "This is some file content from client.",
	}

	err = p2p.SendMessage(conn, msg)
	if err != nil {
		fmt.Println("Send error:", err)
	}
}
