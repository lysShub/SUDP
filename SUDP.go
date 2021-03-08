package sudp

import (
	"SUDP/internal/com"
	"errors"
	"fmt"
	"net"
	"os"
	"time"
)

// SUDP sudp
type SUDP struct {

	// sender
	Speed     int          // send/receive real time speed, KB/s, renewal cycle: 1s
	MTU       int          // variable Byte
	SBasePath string       // sender's base path
	Ssendpath string       // file or floder path
	Sconn     *net.UDPConn // sender's udp conn

	// receiver
	recoder    []int64      // recode write;
	Rconn      *net.UDPConn // receiver's udp conn
	RBasePath  string       //receiver's base path
	Rstorepath string       // store floder path

	// common
	SCF time.Duration // Speed control frequency
	Key [16]byte      // secret key
}

// Sinit sss
func (s *SUDP) Sinit(base string, conn *net.UDPConn) {
	s.Speed = 1024 // 1MB/s
	s.Key = [16]byte{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	s.MTU = 1372
	// s.SBasePath = base
	s.Sconn = conn
	s.SCF = time.Duration(time.Second)

	// s.RBasePath = `E:/a/`
}

//
func (s *SUDP) Rinit() {
	s.Key = [16]byte{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	// s.Rconn = conn
	s.SCF = time.Second
	// s.RBasePath = `/mnt/d/OneDrive/code/go/project/SUDP/test/receiver/`
}

// Send send data packet
func (s *SUDP) Send(filepath string, startBias int64) error {

	fi, err := os.Stat(s.Ssendpath)
	if com.Errorlog(err) {
		return err
	}

	if fi.Mode().IsDir() { // floder
		inf, basepath, out, err := com.GetFloderInfo(filepath)
		if err != nil {
			return err
		} else if out != nil {
			var s string
			for _, v := range out {
				s = com.Wrap() + s + v
			}
			return errors.New("以下文件无法读取或非正常文件" + s)
		}

		for _, p := range inf.N {
			fh, err := os.Open(basepath + p)
			if err != nil {
				return err
			}
			if err := s.sender(fh, s.Sconn, 0, false); err != nil {
				return err
			}
		}

	} else { // file
		fh, err := os.Open(filepath)
		if err != nil {
			return err
		}
		return s.sender(fh, s.Sconn, 0, true)
	}

	return nil

}

// Receive receive data packet
func (s *SUDP) Receive() {

	for {
		fh, fs, raddr, transEnd, err := s.receiverStartFile(s.Rconn)
		if com.Errorlog(err) {
			continue
		}
		fmt.Println("开始, 文件大小", fs)

		if transEnd {
			fmt.Println("传输结束")
			return
		}

		err = s.receiver(fh, fs, s.Rconn, raddr)
		if com.Errorlog(err) {
			continue
		}

	}
}
