/*
 * @Author: 2Kil
 * @Date: 2025-09-28 11:10:39
 * @LastEditors: 2Kil
 * @LastEditTime: 2025-12-15 02:35:53
 * @Description:加密解密及文本处理相关
 */

package tkstar

import (
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"

	"crypto/aes"
	"crypto/cipher"
)

// TextGetKeyPair 生成指定位数的RSA密钥对。
func TextGetKeyPair(bits int) (*KeyPair, error) {
	// 生成两个大素数
	p, err := crand.Prime(crand.Reader, bits/2)
	if err != nil {
		return nil, err
	}
	q, err := crand.Prime(crand.Reader, bits/2)
	if err != nil {
		return nil, err
	}

	n := new(big.Int).Mul(p, q)
	pMinus1 := new(big.Int).Sub(p, big.NewInt(1))
	qMinus1 := new(big.Int).Sub(q, big.NewInt(1))
	phi := new(big.Int).Mul(pMinus1, qMinus1)

	// 动态生成与phi互质的e值
	var e *big.Int

	// 候选e值列表
	candidateEs := []*big.Int{
		big.NewInt(65537 * 7),
		big.NewInt(65537 * 5),
		big.NewInt(65537 * 3),
		big.NewInt(65537), // 标准RSA加密指数
		big.NewInt(17),
		big.NewInt(7),
	}

	// 尝试找到与phi互质的e
	for _, candidate := range candidateEs {
		gcd := new(big.Int).GCD(nil, nil, candidate, phi)
		if gcd.Cmp(big.NewInt(1)) == 0 {
			e = candidate
			break
		}
	}

	if e == nil {
		return nil, fmt.Errorf("无法找到与phi互质的e值")
	}

	// 计算解密指数d
	d := new(big.Int).ModInverse(e, phi)
	if d == nil {
		return nil, fmt.Errorf("无法计算模逆")
	}

	return &KeyPair{
		PublicKey: &PublicKey{
			N: n,
			E: e,
		},
		PrivateKey: &PrivateKey{
			N: n,
			D: d,
		},
	}, nil
}

// TextEncrypt 使用公钥对明文进行加密。
// 注意：这是一个基础的RSA实现，没有使用填充方案（如OAEP），
// 在生产环境中，推荐使用 `crypto/rsa` 包以获得更高的安全性。
func TextEncrypt(pub *PublicKey, plaintext []byte) ([]byte, error) {
	plaintextBig := new(big.Int).SetBytes(plaintext)
	if plaintextBig.Cmp(pub.N) >= 0 {
		return nil, fmt.Errorf("明文数据过大，无法加密")
	}

	// 使用公钥进行加密（幂运算）
	cipherBig := new(big.Int).Exp(plaintextBig, pub.E, pub.N)

	return cipherBig.Bytes(), nil
}

// TextDecrypt 使用私钥解密密文。
func TextDecrypt(priv *PrivateKey, ciphertext []byte) ([]byte, error) {
	cipherBig := new(big.Int).SetBytes(ciphertext)
	plaintextBig := new(big.Int).Exp(cipherBig, priv.D, priv.N)
	return plaintextBig.Bytes(), nil
}

// TextVerify 验证明文的哈希值是否与解密后的哈希值匹配。
// 这个函数适用于签名验证场景，即用公钥解密签名，然后与原文哈希比较。
func TextVerify(plaintext []byte, decryptedHash []byte) bool {
	originalHash := sha256.Sum256(plaintext)
	return hex.EncodeToString(originalHash[:]) == hex.EncodeToString(decryptedHash[:])
}

// TextAesEncrypt 使用AES-CFB模式加密文本。
// 它接受一个明文字符串和一个密钥（16、24或32字节），
// 返回一个经过URL安全的Base64编码的密文字符串和可能的错误。
func TextAesEncrypt(plainText, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", fmt.Errorf("创建密码块失败: %w", err)
	}

	plainTextBytes := []byte(plainText)
	cipherText := make([]byte, aes.BlockSize+len(plainTextBytes))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(crand.Reader, iv); err != nil {
		return "", fmt.Errorf("生成初始化向量失败: %w", err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainTextBytes)

	// 使用URL安全的Base64编码，避免了手动替换字符的需要和相关错误。
	text := base64.URLEncoding.EncodeToString(cipherText)

	return text, nil
}

// TextAesDecrypt 使用AES-CFB模式解密一个经过URL安全Base64编码的字符串。
// 它接受一个密文字符串和一个密钥，返回解密后的明文字符串和可能的错误。
func TextAesDecrypt(cipherText, key string) (string, error) {
	cipherTextBytes, err := base64.URLEncoding.DecodeString(cipherText)
	if err != nil {
		return "", fmt.Errorf("Base64解码失败: %w", err)
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", fmt.Errorf("创建密码块失败: %w", err)
	}

	if len(cipherTextBytes) < aes.BlockSize {
		return "", fmt.Errorf("密文过短，无法解密")
	}

	iv := cipherTextBytes[:aes.BlockSize]
	cipherTextBytes = cipherTextBytes[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherTextBytes, cipherTextBytes)

	return string(cipherTextBytes), nil
}
