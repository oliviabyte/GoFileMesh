package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// 你新加的结构体：用于结构化上传
type StoreFilePayload struct {
	Filename string `json:"filename"`
	Filetype string `json:"filetype"` // optional
	Content  string `json:"content"`
}

// ✅ 文件写入逻辑（后续会用 payload.Filename）
func Store(data string) error {
	var payload StoreFilePayload
	err := json.Unmarshal([]byte(data), &payload)
	if err != nil {
		return fmt.Errorf("invalid JSON format: %v", err)
	}

	if payload.Filename == "" {
		payload.Filename = fmt.Sprintf("data_%d.txt", time.Now().UnixNano())
	}

	err = os.WriteFile(payload.Filename, []byte(payload.Content), 0644)
	if err != nil {
		return err
	}
	fmt.Println("📦 File stored as", payload.Filename)
	return nil
}

func Load(filename string) (string, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return "", err
    }
    return string(data), nil
}

