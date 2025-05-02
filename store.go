package main

import (
	"fmt"
	"os"
	"time"
)

func Store(data string) error {
	// 用时间戳生成唯一文件名（简化处理）
	filename := fmt.Sprintf("data_%d.txt", time.Now().UnixNano())
	err := os.WriteFile(filename, []byte(data), 0644)
	if err != nil {
		return err
	}
	fmt.Println("📦 File stored as", filename)
	return nil
}
