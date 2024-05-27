/*
 * @Author: 2Kil
 * @Date: 2024-05-27 23:07:41
 * @LastEditors: 2Kil
 * @LastEditTime: 2024-05-28 00:08:27
 * @Description: 文本加密解密
 */
package star

import (
	"encoding/base64"
	"fmt"
	"strings"
)

/**
 * @description:异或加密
 * @param {*} text
 * @param {string} key
 * @return {*}
 */
func EorEncode(text, key string) string {
	base64Text := base64.StdEncoding.EncodeToString([]byte(text))
	base64Text = strings.ReplaceAll(base64Text, "1", "！")
	base64Text = strings.ReplaceAll(base64Text, "3", "#")
	base64Text = strings.ReplaceAll(base64Text, "4", "￥")
	base64Text = strings.ReplaceAll(base64Text, "5", "%")
	base64Text = strings.ReplaceAll(base64Text, "8", "*")
	base64Text = strings.ReplaceAll(base64Text, "9", "（")
	btText := []byte(base64Text)
	btKey := []byte(key)
	//补全字符
	if len(btText) > len(btKey) {
		for i := 0; len(btText) > len(btKey); i++ {
			btKey = append(btKey, btText[0])
		}
	} else if len(btText) < len(btKey) {
		for i := 0; len(btText) < len(btKey); i++ {
			btText = append(btText, btKey[0])
		}
	}
	if len(btText) != len(btKey) {
		return ""
	}
	result := make([]byte, len(btText))
	for i := range btText {
		result[i] = btText[i] ^ btKey[i]
	}
	return base64.StdEncoding.EncodeToString(result)
}

func EorDecode(ciphertext, key string) (string, error) {
	btCiphertext, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	btKey := []byte(key)
	// 补全字符
	if len(btCiphertext) > len(btKey) {
		for i := 0; len(btCiphertext) > len(btKey); i++ {
			btKey = append(btKey, btCiphertext[0])
		}
	} else if len(btCiphertext) < len(btKey) {
		for i := 0; len(btCiphertext) < len(btKey); i++ {
			btCiphertext = append(btCiphertext, btKey[0])
		}
	}
	if len(btCiphertext) != len(btKey) {
		return "", fmt.Errorf("unequal lengths of ciphertext and key")
	}

	result := make([]byte, len(btCiphertext))
	for i := range btCiphertext {
		result[i] = btCiphertext[i] ^ btKey[i]
	}

	// 反转替换
	reversedBase64Text := string(result)
	reversedBase64Text = strings.ReplaceAll(reversedBase64Text, "！", "1")
	reversedBase64Text = strings.ReplaceAll(reversedBase64Text, "#", "3")
	reversedBase64Text = strings.ReplaceAll(reversedBase64Text, "￥", "4")
	reversedBase64Text = strings.ReplaceAll(reversedBase64Text, "%", "5")
	reversedBase64Text = strings.ReplaceAll(reversedBase64Text, "*", "8")
	reversedBase64Text = strings.ReplaceAll(reversedBase64Text, "（", "9")

	return reversedBase64Text, nil
}
