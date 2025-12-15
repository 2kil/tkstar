/*
 * @Author: 2Kil
 * @Date: 2025-09-28 11:10:39
 * @LastEditors: 2Kil
 * @LastEditTime: 2025-12-15 10:20:36
 * @Description:结构体定义
 */

package tkstar

import "math/big"

//密钥对
type KeyPair struct {
	PublicKey  *PublicKey
	PrivateKey *PrivateKey
}

type PublicKey struct {
	N *big.Int // 大素数乘积
	E *big.Int // 加密指数
}

type PrivateKey struct {
	N *big.Int // 同公钥的N
	D *big.Int // 解密指数
}
