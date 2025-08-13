package main

import (
	"fmt"
	"time"

	"flag"
)

func main() {
	// 1. parse command line arguments(what user input)
	opts := parseCommandLineArgs()

	// 2. create and start file server
	fs := createAndStartFileServer(opts.port)

	// 3. if user specified a peer, perform connection and test
	if opts.peer != "" {
		performPeerConnectionAndTest(fs, opts)
	}

	keepServerRunning()
}

type CommandLineOptions struct {
	port     string
	peer     string
	download bool
}

func parseCommandLineArgs() *CommandLineOptions {
	port := flag.String("port", "3000", "Server listening port")
	peer := flag.String("peer", "", "Peer node address to connect (optional)")
	download := flag.Bool("download", false, "Test file download after upload")

	flag.Parse()

	return &CommandLineOptions{
		port:     *port,
		peer:     *peer,
		download: *download,
	}
}

func createAndStartFileServer(port string) *FileServer {
	fs := NewFileServer(":" + port)
	go fs.Start()
	time.Sleep(1 * time.Second) // because the server is running in a goroutine, so we need to wait for it to start
	return fs
}

func performPeerConnectionAndTest(fs *FileServer, opts *CommandLineOptions) {
	// connect to the specified peer
	fs.ConnectTo(opts.peer)
	time.Sleep(500 * time.Millisecond)
	fs.UploadFile(opts.peer)
	time.Sleep(1 * time.Second)

	if opts.download {
		performDownloadTest(fs, opts.peer)
	}
}

func performDownloadTest(fs *FileServer, peerAddr string) {
	fmt.Println("Testing local priority logic...")

	fmt.Println("First file request...")
	fs.DownloadFile(peerAddr, "hello_from_client.txt")

	time.Sleep(500 * time.Millisecond)

	fmt.Println("Second request for same file (should use local priority)...")
	fs.DownloadFile(peerAddr, "hello_from_client.txt")
}

// keepServerRunning keeps the program running
func keepServerRunning() {
	select {}
}
