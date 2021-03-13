package main

import (
	sudp "SUDP"
	"SUDP/internal/com"
	"fmt"
	"net"
	"time"
)

func main() {

	laddr, err1 := net.ResolveUDPAddr("udp", ":19986")            //192.168.43.183
	raddr, err2 := net.ResolveUDPAddr("udp", "42.42.42.42:19986") //"172.19.228.218:19986"
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
	S.MTU = 65536  //1072
	S.Speed = 1024 // 1MB
	S.Sconn = conn
	S.SCF = time.Second

	a := time.Now().Unix()

	fmt.Println(S.Send(`D:\android-sdk\system-images\android-30\google_apis\x86\system.img`, 0))
	// fmt.Println(S.Send(`E:\a\assets\gmsstd_stari.apk`, 0))
	b := time.Now().Unix()
	fmt.Println((b - a))
}
