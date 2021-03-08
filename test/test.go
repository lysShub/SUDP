package main

import (
	"SUDP/internal/com"
	"fmt"
)

func main() {
	// var rr [16]byte = [16]byte{
	// 	0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf,
	// }

	// r, _, err := packet.PackageDataPacket([]byte("sdfsaf"), 0x3FFFFF0001, rr, false)
	// com.Errorlog(err)
	// fmt.Println(r)

	// l, bias, final, err := packet.ParseDataPacket(r, rr)
	// com.Errorlog(err)
	// fmt.Println(string(r[:l]))
	// fmt.Println("bias", bias)
	// fmt.Println(final)

	inf, basePath, out, err := com.GetFloderInfo(`E:\a`)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(out)
	fmt.Println(basePath)
	// fmt.Println(inf)
	for i, v := range inf.N {
		fmt.Println(v, inf.S[i], inf.T[i])
	}

}
