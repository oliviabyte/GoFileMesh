package main
// 在 sendPing() 中加密 payload.Content
// 在 sendGetFile() 中接收到内容后解密

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// 写死一个密钥（32字节 = 256位）
var secretKey = []byte("thisis32bitlongpassphraseimusing") // 32 bytes key

// Padding 用于补齐文本长度，为了确保数据块长度是AES块大小(16字节)的倍数
func pkcs7Padding(data []byte, blockSize int) []byte {
	pad := blockSize - len(data)%blockSize
	padding := bytes.Repeat([]byte{byte(pad)}, pad)
	return append(data, padding...)
}

// 去除 Padding，为了在解密后移除PKCS#7填充
func pkcs7Unpadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("data is empty")
	}
	pad := int(data[length-1])
	return data[:(length - pad)], nil
}

// 加密函数，Encrypt 使用 AES-CBC 加密，并返回 base64 编码字符串
func Encrypt(plaintext string) (string, error) {
	// 创建一个新的AES加密器
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", err
	}

	plainBytes := pkcs7Padding([]byte(plaintext), block.BlockSize())
	ciphertext := make([]byte, aes.BlockSize+len(plainBytes))

	// 生成随机IV并放在密文的前16字节
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	// 创建一个CBC模式的加密器
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plainBytes)

	// 将二进制数据转换为base64编码的字符串,使加密后的二进制数据可以在文本环境中安全传输
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// 解密函数，Decrypt 解密 base64 编码的 AES 密文
func Decrypt(cipherBase64 string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(cipherBase64)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	plainBytes, err := pkcs7Unpadding(ciphertext)
	if err != nil {
		return "", err
	}

	return string(plainBytes), nil
}
