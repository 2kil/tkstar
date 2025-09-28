/*
 * @Author: 2Kil
 * @Date: 2025-09-28 11:10:39
 * @LastEditors: 2Kil
 * @LastEditTime: 2025-09-28 11:26:11
 * @Description:
 */

package tkstar

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

// ----------------------------------------------
// @description: 生成RSA密钥对
// @param {int} bits
// @return {*}
// ----------------------------------------------
func GenerateKeyPair(bits int) (*KeyPair, error) {
	// 生成两个大素数
	p, err := rand.Prime(rand.Reader, bits/2)
	if err != nil {
		return nil, err
	}
	q, err := rand.Prime(rand.Reader, bits/2+1)
	if err != nil {
		return nil, err
	}

	n := new(big.Int).Mul(p, q)
	pMinus1 := new(big.Int).Sub(p, big.NewInt(1))
	qMinus1 := new(big.Int).Sub(q, big.NewInt(1))
	phi := new(big.Int).Mul(pMinus1, qMinus1)

	// 动态生成与phi互质的e值
	var e *big.Int
	maxAttempts := 100
	attempts := 0

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
		attempts++
		gcd := new(big.Int).GCD(nil, nil, candidate, phi)
		if gcd.Cmp(big.NewInt(1)) == 0 {
			e = candidate
			break
		}
		if attempts >= maxAttempts {
			return nil, fmt.Errorf("超过最大尝试次数，无法找到与phi互质的e")
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

// ----------------------------------------------
// @description: RSA加密
// @param {*PublicKey} pub
// @param {[]byte} plaintext
// @param {int} workFactor
// @return {*}
// ----------------------------------------------
func Encrypt(pub *PublicKey, plaintext []byte, workFactor int) ([]byte, error) {
	// 1. 计算密集型预热，以控制复杂度
	// 使用 crypto/rand 生成随机数据，防止编译器优化
	randomData := make([]byte, 32) // 每次哈希32字节
	for i := 0; i < workFactor; i++ {
		// 生成随机数据
		if _, err := rand.Read(randomData); err != nil {
			return nil, fmt.Errorf("生成随机数失败: %v", err)
		}
		// 执行一次SHA-256哈希
		_ = sha256.Sum256(randomData)
	}

	// 2. 原始加密逻辑保持不变
	hashed := sha256.Sum256(plaintext)
	hashBig := new(big.Int).SetBytes(hashed[:])

	if hashBig.Cmp(pub.N) >= 0 {
		return nil, fmt.Errorf("哈希值过大，无法加密")
	}

	// 幂运算
	cipherBig := new(big.Int).Exp(hashBig, pub.E, pub.N)

	return cipherBig.Bytes(), nil
}

// ----------------------------------------------
// @description: 解密（高效）
// @param {*PrivateKey} priv
// @param {[]byte} ciphertext
// @return {*}
// ----------------------------------------------
func Decrypt(priv *PrivateKey, ciphertext []byte) ([]byte, error) {
	cipherBig := new(big.Int).SetBytes(ciphertext)
	hashBig := new(big.Int).Exp(cipherBig, priv.D, priv.N)
	return hashBig.Bytes(), nil
}

// ----------------------------------------------
// @description: 验证哈希匹配
// @param {[]byte} plaintext
// @param {[]byte} decryptedHash
// @return {*}
// ----------------------------------------------
func Verify(plaintext []byte, decryptedHash []byte) bool {
	originalHash := sha256.Sum256(plaintext)

	fixedHash := make([]byte, 32)
	if len(decryptedHash) > 32 {
		copy(fixedHash, decryptedHash[len(decryptedHash)-32:])
	} else {
		copy(fixedHash[32-len(decryptedHash):], decryptedHash)
	}

	return hex.EncodeToString(originalHash[:]) == hex.EncodeToString(fixedHash)
}
