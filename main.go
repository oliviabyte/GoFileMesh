// main.go is the entry point of the GoFileMesh system

package main

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"flag"

	"github.com/heyxtanya/GoFileMesh/p2p"
)

type FileAnnouncement struct {
	Filename string `json:"filename"`
	Addr     string `json:"addr"` // eg. "localhost:3001"
}

func main() {
	port := flag.String("port", "3000", "port to listen on")
	peer := flag.String("peer", "", "peer to connect to (optional)")
	download := flag.Bool("download", false, "whether to request the file after upload")
	flag.Parse()

	addr := ":" + *port
	t := startServer(addr) // 👈 拿到返回的 transport 实例（含 peerMap）

	time.Sleep(1 * time.Second)

	if *peer != "" {
		connectToPeer(*peer, t) // 👈 连接并注册 peer

		// 上传文件
		sendPing(*peer)

		// 广播给其他节点（peerMap 中的所有人）
		time.Sleep(500 * time.Millisecond)
		broadcastNewFile("hello_from_client.txt", t) // 广播新文件名

		if *download {
			time.Sleep(1 * time.Second)
			sendGetFile(*peer, "hello_from_client.txt") // 下载副本
		}
	}
	select {} // 保持运行
}

func startServer(addr string) *p2p.TCPTransport {
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

	t.OnMessage("FILE_ANNOUNCEMENT", func(msg p2p.Message, conn net.Conn) {
		var fa FileAnnouncement
		json.Unmarshal([]byte(msg.Data), &fa)
		fmt.Println("🛰  Peer announced new file:", fa.Filename)
		go sendGetFile(fa.Addr, fa.Filename)
	})

	if err := t.Start(); err != nil {
		panic(err)
	}

	return t
}

func sendPing(addr string) {
	fmt.Println("⚡️sendPing started")
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("Dial error:", err)
		return
	}
	defer conn.Close()

	// 原始内容
	rawContent := "This is some structured file content."

	// ✅ 加密 content
	encryptedContent, err := Encrypt(rawContent)
	if err != nil {
		fmt.Println("❌ Encryption error:", err)
		return
	}

	// ✅ 构造结构化 payload
	payload := StoreFilePayload{
		Filename: "hello_from_client.txt",
		Filetype: "text/plain",
		Content:  encryptedContent,
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

	// ✅ 尝试解密内容
	decrypted, err := Decrypt(response.Data)
	if err != nil {
		fmt.Println("❌ Decryption error:", err)
		return
	}
	fmt.Println("📥 File content received (decrypted):", decrypted)
}

// 建立连接并保存
func connectToPeer(addr string, t *p2p.TCPTransport) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("❌ Failed to connect to peer:", err)
		return
	}
	t.AddPeer(conn) // 👈 添加进 peerMap
}

func broadcastNewFile(filename string, t *p2p.TCPTransport) {
	announcement := FileAnnouncement{
		Filename: filename,
		Addr:     t.Addr(), // 👈 这是本节点对外暴露的端口
	}
	bytes, _ := json.Marshal(announcement)

	msg := p2p.Message{
		Type: "FILE_ANNOUNCEMENT",
		Data: string(bytes),
	}

	for _, conn := range t.Peers() {
		p2p.SendMessage(conn, msg)
	}
}
