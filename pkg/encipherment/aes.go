package encipherment

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

// AesEncrypt 加密
// key length must 16, 24, or 32 bytes to select
func AesEncrypt(originalText string, key string) string {
	// 转成字节数组
	originalData := []byte(originalText)
	keyData := []byte(key)

	// 分组密钥
	block, err := aes.NewCipher(keyData)
	if err != nil {
		panic(fmt.Sprintf("key 长度必须 16/24/32长度: %s", err.Error()))
	}
	// 获取密钥快长度
	blockSize := block.BlockSize()
	// 补全码
	origData := PKCS7Padding(originalData, blockSize)
	// 加密模式
	blockMode := cipher.NewCBCEncrypter(block, keyData[:blockSize])
	// 创建数组
	encrypted := make([]byte, len(origData))
	// 加密
	blockMode.CryptBlocks(encrypted, origData)
	//使用RawURLEncoding 不要使用StdEncoding
	//不要使用StdEncoding  放在url参数中回导致错误
	return base64.RawURLEncoding.EncodeToString(encrypted)

}

// AesDecrypt aes解密
func AesDecrypt(ciphertext string, key string) string {
	//使用RawURLEncoding 不要使用StdEncoding
	//不要使用StdEncoding  放在url参数中回导致错误
	decryptedByte, _ := base64.RawURLEncoding.DecodeString(ciphertext)
	k := []byte(key)

	// 分组密钥
	block, err := aes.NewCipher(k)
	if err != nil {
		panic(fmt.Sprintf("key 长度必须 16/24/32长度: %s", err.Error()))
	}
	// 获取密钥块的长度
	blockSize := block.BlockSize()
	// 加密模式
	blockMode := cipher.NewCBCDecrypter(block, k[:blockSize])
	// 创建数组
	orig := make([]byte, len(decryptedByte))
	// 解密
	blockMode.CryptBlocks(orig, decryptedByte)
	// 去补全码
	orig = PKCS7UnPadding(orig)
	return string(orig)
}

// PKCS7Padding 补码
func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padText...)
}

// PKCS7UnPadding 去码
func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unPadding := int(origData[length-1])
	return origData[:(length - unPadding)]
}
