/*
 * @Author: 2Kil
 * @Date: 2024-05-27 23:07:41
 * @LastEditors: 2Kil
 * @LastEditTime: 2024-05-28 11:15:40
 * @Description: 文本加密解密
 */
package star

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"strings"
)

/**
 * @description: aes加密
 * @param {string} 待加密的文本
 * @param {string} 16,24,32密钥
 * @return {string} 密文
 */
func AesEncrypt(plainText, key string) string {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return ""
	}

	plainTextBytes := []byte(plainText)
	cipherText := make([]byte, aes.BlockSize+len(plainTextBytes))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return ""
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainTextBytes)
	text := base64.StdEncoding.EncodeToString(cipherText)

	//替换base64特殊字符
	text = strings.ReplaceAll(text, "/", "*")
	text = strings.ReplaceAll(text, "==", "#")
	text = strings.ReplaceAll(text, "=", "$")

	return text
}

/**
 * @description: aes解密
 * @param {string} 待解密的文本
 * @param {string} 密钥
 * @return {string} 明文
 */
func AesDecrypt(cipherText, key string) string {
	//替换base64特殊字符
	cipherText = strings.ReplaceAll(cipherText, "*", "/")
	cipherText = strings.ReplaceAll(cipherText, "$", "=")
	cipherText = strings.ReplaceAll(cipherText, "#", "==")

	cipherTextBytes, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return ""
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return ""
	}

	if len(cipherTextBytes) < aes.BlockSize {
		return ""
	}

	iv := cipherTextBytes[:aes.BlockSize]
	cipherTextBytes = cipherTextBytes[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherTextBytes, cipherTextBytes)

	return string(cipherTextBytes)
}
