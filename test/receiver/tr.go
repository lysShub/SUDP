package main

import (
	sudp "SUDP"
	"SUDP/internal/com"
	"fmt"
	"net"
	"time"
)

func main() {

	laddr, err := net.ResolveUDPAddr("udp", ":19986")
	if com.Errorlog(err) {
		return
	}
	conn, err := net.ListenUDP("udp", laddr)
	if com.Errorlog(err) {
		return
	}
	fmt.Println("开始接收")

	// 开始
	var S = new(sudp.SUDP)
	S.Key = [16]byte{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	S.Rconn = conn
	S.SCF = time.Second
	S.Rstorepath = `/mnt/d/OneDrive/code/go/project/SUDP/test/receiver/`
	S.MTU = 1072

	fmt.Println(S.Receive())
}
