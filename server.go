package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"

	"github.com/heyxtanya/GoFileMesh/p2p"
)

type FileAnnouncement struct {
	Filename string `json:"filename"`
	Addr     string `json:"addr"`
}

type FileServer struct {
	Transport *p2p.TCPTransport
}

// 构造函数：注册 handler
func NewFileServer(addr string) *FileServer {
	t := p2p.NewTCPTransport(addr)
	fs := &FileServer{Transport: t}

	// 注册 handler
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
		go fs.DownloadFile(fa.Addr, fa.Filename)
	})

	return fs
}

func (fs *FileServer) Start() {
	if err := fs.Transport.Start(); err != nil {
		panic(err)
	}
}

func (fs *FileServer) ConnectTo(addr string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("❌ Failed to connect to peer:", err)
		return
	}
	fs.Transport.AddPeer(conn)
}

func (fs *FileServer) BroadcastFile(filename string) {
	announcement := FileAnnouncement{
		Filename: filename,
		Addr:     fs.Transport.Addr(),
	}
	bytes, _ := json.Marshal(announcement)

	msg := p2p.Message{
		Type: "FILE_ANNOUNCEMENT",
		Data: string(bytes),
	}

	for _, conn := range fs.Transport.Peers() {
		p2p.SendMessage(conn, msg)
	}
}

func (fs *FileServer) UploadFile(addr string) {
	fmt.Println("⚡️sendPing started")
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("Dial error:", err)
		return
	}
	defer conn.Close()

	rawContent := "This is some structured file content."
	encryptedContent, err := Encrypt(rawContent)
	if err != nil {
		fmt.Println("❌ Encryption error:", err)
		return
	}

	hash := sha256.Sum256([]byte(encryptedContent))
	hashStr := hex.EncodeToString(hash[:])

	payload := StoreFilePayload{
		Filename: "hello_from_client.txt",
		Filetype: "text/plain",
		Content:  encryptedContent,
		Hash:     hashStr,
	}

	bytes, _ := json.Marshal(payload)

	msg := p2p.Message{
		Type: "STORE_FILE",
		Data: string(bytes),
	}
	p2p.SendMessage(conn, msg)
}

func (fs *FileServer) DownloadFile(addr, filename string) {
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
	p2p.SendMessage(conn, msg)

	response, err := p2p.ReceiveMessage(conn)
	if err != nil {
		fmt.Println("Receive error:", err)
		return
	}

	decrypted, err := Decrypt(response.Data)
	if err != nil {
		fmt.Println("❌ Decryption error:", err)
		return
	}
	fmt.Println("📥 File content received (decrypted):", decrypted)
}