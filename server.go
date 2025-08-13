/*
 * GoFileMesh - File Server Core
 *
 * Registers message handlers and routes incoming messages to their business logic.
 * Coordinates file operations and network communication.
 *
 * Architecture: Business logic layer
 */

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"

	"github.com/heyxtanya/GoFileMesh/p2p"
	"github.com/heyxtanya/GoFileMesh/store"
)

// FileAnnouncement represents a file broadcast message
// when a node has a new file, it will announce it to other nodes
type FileAnnouncement struct {
	Filename string `json:"filename"`
	Addr     string `json:"addr"`
}

// FileServer coordinates network transport, file storage and encryption
type FileServer struct {
	Transport *p2p.TCPTransport
}

// NewFileServer creates and initializes a file server
func NewFileServer(addr string) *FileServer {
	t := p2p.NewTCPTransport(addr)
	fs := &FileServer{Transport: t}
	fs.registerMessageHandlers()
	return fs
}

// Start starts the TCP listener
func (fs *FileServer) Start() {
	if err := fs.Transport.Start(); err != nil {
		panic(err) // if the start fails, the program will crash (because this is a critical function)
	}
}

// ConnectTo establishes connection to a specified node
func (fs *FileServer) ConnectTo(addr string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("Failed to connect to peer:", err)
		return
	}
	fs.Transport.AddPeer(conn)
}

// UploadFile replicates file to all connected nodes
func (fs *FileServer) UploadFile(addr string) {
	fmt.Println("Start file replication to all nodes")

	rawContent := "This is some structured file content."
	encryptedContent, err := Encrypt(rawContent)
	if err != nil {
		fmt.Println("Encryption error:", err)
		return
	}

	hash := sha256.Sum256([]byte(encryptedContent))
	hashStr := hex.EncodeToString(hash[:])

	payload := store.StoreFilePayload{
		Filename: "hello_from_client.txt",
		Filetype: "text/plain",
		Content:  encryptedContent,
		Hash:     hashStr,
	}

	bytes, _ := json.Marshal(payload)

	// Store locally first
	fmt.Println("Store file locally...")
	err = store.Store(string(bytes))
	if err != nil {
		fmt.Println("Local storage failed:", err)
		return
	}
	fmt.Println("File stored locally")

	// Broadcast file availability
	fmt.Println("Broadcast file availability...")
	fs.BroadcastFile(hashStr + ".txt")

	// Replicate to all connected nodes
	fmt.Println("Start file replication to all nodes...")
	peers := fs.Transport.Peers()
	if len(peers) == 0 {
		fmt.Println("No connected nodes, skipping replication")
		return
	}

	msg := p2p.Message{
		Type: "STORE_FILE",
		Data: string(bytes),
	}

	for _, conn := range peers {
		peerAddr := conn.RemoteAddr().String()
		fmt.Printf("Copy file to node: %s\n", peerAddr)

		err := p2p.SendMessage(conn, msg)
		if err != nil {
			fmt.Printf("Failed to copy to node %s: %v\n", peerAddr, err)
		} else {
			fmt.Printf("Successfully copied to node: %s\n", peerAddr)
		}
	}

	fmt.Printf("File replication complete! Copied to %d nodes\n", len(peers))
}

// DownloadFile implements local-first, network-fallback strategy
func (fs *FileServer) DownloadFile(addr, filename string) {
	fmt.Println("Requesting file:", filename)

	// Local-first: check if file exists locally
	if store.Has(filename) {
		fmt.Printf("File %s already exists locally, reading directly\n", filename)
		content, err := store.Load(filename)
		if err != nil {
			fmt.Printf("Local read failed: %v\n", err)
		} else {
			decrypted, err := Decrypt(content)
			if err != nil {
				fmt.Printf("Decryption failed: %v\n", err)
			} else {
				fmt.Printf("Local file content (decrypted): %s\n", decrypted)
			}
		}
		return
	}

	// Network fallback: fetch from network if not local
	fmt.Printf("File %s not found locally, fetching from network...\n", filename)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("Connection failed:", err)
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
		fmt.Println("Failed to receive response:", err)
		return
	}

	decrypted, err := Decrypt(response.Data)
	if err != nil {
		fmt.Println("Decryption failed:", err)
		return
	}

	fmt.Printf("Network fetched file content (decrypted): %s\n", decrypted)

	// Store locally for next access
	fmt.Printf("Storing file locally for next access...\n")

	hash := sha256.Sum256([]byte(response.Data))
	hashStr := hex.EncodeToString(hash[:])

	payload := store.StoreFilePayload{
		Filename: filename,
		Filetype: "text/plain",
		Content:  response.Data,
		Hash:     hashStr,
	}

	bytes, _ := json.Marshal(payload)
	err = store.Store(string(bytes))
	if err != nil {
		fmt.Printf("Local storage failed: %v\n", err)
	} else {
		fmt.Printf("File stored locally, next access will use local priority\n")
	}
}

// DeleteFile requests file deletion from specified node
func (fs *FileServer) DeleteFile(addr, filename string) {
	fmt.Println("Requesting delete for file:", filename)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("Dial error:", err)
		return
	}
	defer conn.Close()

	msg := p2p.Message{
		Type: "DELETE_FILE",
		Data: filename,
	}

	p2p.SendMessage(conn, msg)

	response, err := p2p.ReceiveMessage(conn)
	if err != nil {
		fmt.Println("Receive error:", err)
		return
	}

	fmt.Println("Delete response:", response.Data)
}

// registerMessageHandlers registers all message handlers
func (fs *FileServer) registerMessageHandlers() {
	fs.Transport.OnMessage("STORE_FILE", func(msg p2p.Message, conn net.Conn) {
		fmt.Println("Received file content to store.")
		err := store.Store(msg.Data)
		if err != nil {
			fmt.Println("Store error:", err)
		}
	})

	fs.Transport.OnMessage("GET_FILE", func(msg p2p.Message, conn net.Conn) {
		fmt.Println("Received GET_FILE request:", msg.Data)

		content, err := store.Load(msg.Data)
		if err != nil {
			fmt.Println("Load error:", err)
			return
		}

		response := p2p.Message{
			Type: "FILE_CONTENT",
			Data: content,
		}
		p2p.SendMessage(conn, response)
	})

	fs.Transport.OnMessage("HAS_FILE", func(msg p2p.Message, conn net.Conn) {
		has := store.Has(msg.Data)
		response := p2p.Message{
			Type: "HAS_FILE_RESPONSE",
			Data: fmt.Sprintf("%v", has),
		}
		p2p.SendMessage(conn, response)
	})

	fs.Transport.OnMessage("DELETE_FILE", func(msg p2p.Message, conn net.Conn) {
		fmt.Println("Received DELETE_FILE request:", msg.Data)

		err := store.Delete(msg.Data)
		if err != nil {
			fmt.Println("Delete error:", err)
			response := p2p.Message{
				Type: "DELETE_FILE_RESPONSE",
				Data: fmt.Sprintf("error: %v", err),
			}
			p2p.SendMessage(conn, response)
			return
		}

		response := p2p.Message{
			Type: "DELETE_FILE_RESPONSE",
			Data: "success",
		}
		p2p.SendMessage(conn, response)
	})

	fs.Transport.OnMessage("FILE_ANNOUNCEMENT", func(msg p2p.Message, conn net.Conn) {
		var fa FileAnnouncement
		json.Unmarshal([]byte(msg.Data), &fa)
		fmt.Println("Peer announced new file:", fa.Filename)

		go fs.DownloadFile(fa.Addr, fa.Filename)
	})
}

// BroadcastFile notifies all connected nodes about file availability
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
