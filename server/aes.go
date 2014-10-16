// @description wechat 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/chanxuehong/wechat for the canonical source repository
// @license     https://github.com/chanxuehong/wechat/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package server

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
)

// 把整数 n 格式化成 4 字节的网络字节序
func encodeNetworkBytesOrder(n int, orderBytes []byte) {
	if len(orderBytes) != 4 {
		panic("the length of orderBytes must be equal to 4")
	}
	orderBytes[0] = byte(n >> 24)
	orderBytes[1] = byte(n >> 16)
	orderBytes[2] = byte(n >> 8)
	orderBytes[3] = byte(n)
}

// 从 4 字节的网络字节序里解析出整数
func decodeNetworkBytesOrder(orderBytes []byte) (n int) {
	if len(orderBytes) != 4 {
		panic("the length of orderBytes must be equal to 4")
	}
	n = int(orderBytes[0])<<24 |
		int(orderBytes[1])<<16 |
		int(orderBytes[2])<<8 |
		int(orderBytes[3])
	return
}

func encryptMsg(random [16]byte, rawXMLMsg []byte, AppId string, AESKey [32]byte) (encryptMsg []byte) {
	const BLOCK_SIZE = 32 // PKCS#7

	buf := make([]byte, 20+len(rawXMLMsg)+len(AppId)+BLOCK_SIZE)
	plain := buf[:20]
	pad := buf[len(buf)-BLOCK_SIZE:]

	// 拼接
	copy(plain, random[:])
	encodeNetworkBytesOrder(len(rawXMLMsg), plain[16:20])
	plain = append(plain, rawXMLMsg...)
	plain = append(plain, AppId...)

	// PKCS#7 补位
	amountToPad := BLOCK_SIZE - len(plain)%BLOCK_SIZE
	pad = pad[:amountToPad]
	for i := 0; i < amountToPad; i++ {
		pad[i] = byte(amountToPad)
	}
	plain = append(plain, pad...)

	// 加密
	block, err := aes.NewCipher(AESKey[:])
	if err != nil {
		panic(err)
	}
	mode := cipher.NewCBCEncrypter(block, AESKey[:16])
	mode.CryptBlocks(plain, plain)

	encryptMsg = plain
	return
}

func decryptMsg(encryptMsg []byte, AppId string, AESKey [32]byte) (random [16]byte, rawXMLMsg []byte, err error) {
	const BLOCK_SIZE = 32 // PKCS#7

	// 解密
	if len(encryptMsg) < BLOCK_SIZE {
		err = errors.New("encryptMsg too short")
		return
	}
	if len(encryptMsg)%BLOCK_SIZE != 0 {
		err = errors.New("encryptMsg is not a multiple of the block size")
		return
	}

	block, err := aes.NewCipher(AESKey[:])
	if err != nil {
		panic(err)
	}
	mode := cipher.NewCBCDecrypter(block, AESKey[:16])

	plain := make([]byte, len(encryptMsg))
	mode.CryptBlocks(plain, encryptMsg)

	// PKCS#7 去除补位
	amountToPad := int(plain[len(plain)-1])
	if amountToPad < 1 || amountToPad > BLOCK_SIZE {
		err = errors.New("the amount to pad is invalid")
		return
	}
	plain = plain[:len(plain)-amountToPad]

	// 反拼装
	if len(plain) <= 20 {
		err = errors.New("plain too short")
		return
	}
	msgLen := decodeNetworkBytesOrder(plain[16:20])
	if msgLen < 0 {
		err = fmt.Errorf("invalid msg length: %d", msgLen)
		return
	}
	msgEnd := 20 + msgLen
	if msgEnd >= len(plain) {
		err = fmt.Errorf("msg length is too large: %d", msgLen)
		return
	}

	AppIdHave := string(plain[msgEnd:])
	if AppIdHave != AppId { // crypto/subtle.ConstantTimeCompare ???
		err = errors.New("AppId mismatch")
		return
	}

	random[0] = plain[0]
	random[1] = plain[1]
	random[2] = plain[2]
	random[3] = plain[3]
	random[4] = plain[4]
	random[5] = plain[5]
	random[6] = plain[6]
	random[7] = plain[7]
	random[8] = plain[8]
	random[9] = plain[9]
	random[10] = plain[10]
	random[11] = plain[11]
	random[12] = plain[12]
	random[13] = plain[13]
	random[14] = plain[14]
	random[15] = plain[15]

	rawXMLMsg = plain[20:msgEnd]
	return
}