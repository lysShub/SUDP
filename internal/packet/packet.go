package packet

import (
	"SUDP/internal/com"
	"SUDP/internal/crypto"
	"errors"
	"hash/crc32"
)

// PackageDataPacket 打包为数据包
// 参数: d:原始数据; b:偏置; k:密钥; final:最后一个数据包
// 返回：打包后数据包，原始数据长度，是否最后一个数据包
func PackageDataPacket(d []byte, b int64, k [16]byte, final bool) ([]byte, int, bool, error) {
	var dl int = len(d)
	//filling
	var l uint8 = uint8(16 - (len(d)+9)%16)
	var flag bool = false
	for i := uint8(0); i < l; i++ {
		d = append(d, uint8(l))
		flag = true
	}

	// checksum
	s := crc32.ChecksumIEEE(d)
	d = append(d, uint8(s), uint8(s>>8), uint8(s>>16), uint8(s>>24))

	// data bias
	for i := uint8(0); i < 4; i++ {
		d = append(d, uint8(b>>(8*(4-i)-2)))
	}
	d = append(d, uint8(b<<2))

	var end uint8 = 0
	if final {
		end = 1
	}
	if flag {
		d[len(d)-1] = d[len(d)-1] + 2 + end
	} else {
		d[len(d)-1] = d[len(d)-1] + 0 + end
	}

	if err := crypto.CbcEncrypt(k[:], d); com.Errorlog(err) { // encrypto
		d = nil
		return nil, dl, final, err
	}

	return d, dl, final, nil
}

// ParseDataPacket parse data packet
// 注意：使用本包解析后，d[:数据长度]才是原始数据
// 返回: 原始数据长度; 偏置; 是否最后包
func ParseDataPacket(d []byte, k [16]byte) (int, int64, bool, error) {

	if err := crypto.CbcDecrypt(k[:], d); com.Errorlog(err) { // decrypto
		return 0, -1, false, err
	}

	// check sum
	rS := crc32.ChecksumIEEE(d[:len(d)-5])
	if rS != 0x2144DF1C {
		err := errors.New("CRC verify failed")
		com.Errorlog(err)
		return 0, -1, false, err
	}

	var l uint8 = 0
	if d[len(d)-1]&uint8(2) != 0 { //filled
		l = d[len(d)-10] //fill data len
	}

	// bias
	var b int64 = int64(d[len(d)-5])<<30 + int64(d[len(d)-4])<<22 + int64(d[len(d)-3])<<14 + int64(d[len(d)-2])<<6 + int64(d[len(d)-1])>>2

	var end bool = false
	if 1&uint8(d[len(d)-1]) == 1 { // final data packet
		end = true
	}

	d = d[:len(d)-9-int(l)] // over write
	return len(d), b, end, nil
}
