// main.go is the entry point of the GoFileMesh system

package main

import (
	"encoding/json"
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

	// ✅ 上传文件
	sendPing("localhost:3000")

	// ✅ 然后再下载文件
	sendGetFile("localhost:3000", "hello_from_client.txt")

	select {} // 保持运行
}

func startServer(addr string) {
	t := p2p.NewTCPTransport(addr)

	// ✅ 先注册所有处理函数
	t.OnMessage("STORE_FILE", func(msg p2p.Message, conn net.Conn) {
		fmt.Println("📥 Received file content to store.")
		err := Store(msg.Data)
		if err != nil {
			fmt.Println("❌ Store error:", err)
		}
	})

	t.OnMessage("GET_FILE", func(msg p2p.Message, conn net.Conn) {
		fmt.Println("📤 Received GET_FILE request:", msg.Data)

		content, err := Load(msg.Data)
		if err != nil {
			fmt.Println("❌ Load error:", err)
			return
		}

		response := p2p.Message{
			Type: "FILE_CONTENT",
			Data: content,
		}
		p2p.SendMessage(conn, response)
	})

	// ✅ 最后再启动服务器
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

	// ✅ 构造结构化 payload
	payload := StoreFilePayload{
		Filename: "hello_from_client.txt",
		Filetype: "text/plain",
		Content:  "This is some structured file content.",
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("❌ JSON marshal error:", err)
		return
	}

	msg := p2p.Message{
		Type: "STORE_FILE",
		Data: string(bytes), // 👈 结构化 JSON 字符串
	}

	err = p2p.SendMessage(conn, msg)
	if err != nil {
		fmt.Println("Send error:", err)
	}
}

func sendGetFile(addr, filename string) {
	fmt.Println("📬 Requesting file:", filename)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("Dial error:", err)
		return
	}
	defer conn.Close()

	msg := p2p.Message{
		Type: "GET_FILE",
		Data: filename,
	}
	err = p2p.SendMessage(conn, msg)
	if err != nil {
		fmt.Println("Send error:", err)
	}

	// 接收返回的消息（模拟服务器直接回发）
	response, err := p2p.ReceiveMessage(conn)
	if err != nil {
		fmt.Println("Receive error:", err)
		return
	}
	fmt.Println("📥 File content received:", response.Data)
}
