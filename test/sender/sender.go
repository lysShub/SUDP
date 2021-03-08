package main

import (
	sudp "SUDP"
	"SUDP/internal/com"
	"fmt"
	"net"
	"os"
	"time"
)

func main() {

	laddr, err1 := net.ResolveUDPAddr("udp", ":19986") //192.168.43.183
	raddr, err2 := net.ResolveUDPAddr("udp", "172.19.226.94:19986")
	if com.Errorlog(err1, err2) {
		return
	}
	conn, err := net.DialUDP("udp", laddr, raddr)
	if com.Errorlog(err) {
		return
	}

	// 开始

	var S = new(sudp.SUDP)
	S.Key = [16]byte{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	S.MTU = 1372
	S.Speed = 1024 // 1MB
	S.RBasePath = `E:/a/`
	S.Rconn = conn
	S.SCF = time.Second

	fh, err := os.Open(`E:/a/classes.dex`)
	if com.Errorlog(err) {
		return
	}
	fmt.Println(S.Send(fh, 0))

}
