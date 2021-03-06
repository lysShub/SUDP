package packet

import (
	"SUDP/internal/com"
	"SUDP/internal/crypto"
	"errors"
	"hash/crc32"
)

// PackageDataPacket pack data parse
// pars: d:origin data; b:bias; k:secret key; final:finally data packet
// make sure parmeter d has enough cap(len+9+16); otherwise it will use 2 times the memory
func PackageDataPacket(d []byte, b int64, k [16]byte, final bool) ([]byte, bool, error) {

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
		return nil, final, err
	}

	return d, final, nil
}

// ParseDataPacket parse data packet
// 注意：使用本包解析后，d[:数据长度]才是原始数据
// return: origin data length; bias; final packet
//
//
func ParseDataPacket(d []byte, k [16]byte) (uint16, int64, bool, error) {

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
	if uint8(1)&uint8(d[len(d)-1]) == 1 { // final data packet
		end = true
	}

	d = d[:len(d)-9-int(l)] // over write
	return uint16(len(d)), b, end, nil
}
