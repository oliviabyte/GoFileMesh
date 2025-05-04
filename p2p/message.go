package p2p

import (
	"encoding/gob"
	"net"
)

type Message struct {
	Type string
	Data string
}

// 把文件通过tcp发送到server
func SendMessage(conn net.Conn, msg Message) error {
	encoder := gob.NewEncoder(conn)
	return encoder.Encode(msg)
}

func ReceiveMessage(conn net.Conn) (Message, error) {
	var msg Message
	decoder := gob.NewDecoder(conn)
	err := decoder.Decode(&msg)
	return msg, err
}
