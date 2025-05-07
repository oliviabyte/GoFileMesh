// main.go is the entry point of the GoFileMesh system

package main

import (
	"time"

	"flag"
)

func main() {
	port := flag.String("port", "3000", "port to listen on")
	peer := flag.String("peer", "", "peer to connect to (optional)")
	download := flag.Bool("download", false, "whether to request the file after upload")
	flag.Parse()

	fs := NewFileServer(":" + *port)
	go fs.Start()

	time.Sleep(1 * time.Second)

	if *peer != "" {
		fs.ConnectTo(*peer)
		fs.UploadFile(*peer)

		time.Sleep(500 * time.Millisecond)
		fs.BroadcastFile("hello_from_client.txt")

		if *download {
			time.Sleep(1 * time.Second)
			fs.DownloadFile(*peer, "hello_from_client.txt")
		}
	}

	select {}
}
