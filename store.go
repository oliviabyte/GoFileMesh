package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// 新加的结构体：用于结构化上传
type StoreFilePayload struct {
	Filename string `json:"filename"`
	Filetype string `json:"filetype"` // optional
	Content  string `json:"content"`
	Hash string `json:"hash"` // SHA256 哈希
}

func Store(data string) error {
	// 解析结构化 JSON 数据
	var payload StoreFilePayload
	err := json.Unmarshal([]byte(data), &payload)
	if err != nil {
		return fmt.Errorf("❌ JSON decode error: %v", err)
	}

	// 用 hash 作为唯一 ID 进行去重判断
	filename := payload.Hash + ".txt"
	if _, err := os.Stat(filename); err == nil {
		fmt.Println("⚠️  File with this hash already exists, skipping write.")
		return nil // 不重复写入
	}

	// 写入新文件
	err = os.WriteFile(filename, []byte(payload.Content), 0644)
	if err != nil {
		return err
	}

	fmt.Println("📦 File stored as", filename)
	return nil
}

func Load(filename string) (string, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return "", err
    }
    return string(data), nil
}

