package sudp

import (
	"SUDP/internal/com"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

// SUDP sudp
type SUDP struct {

	// sender
	Speed int          // send/receive real time speed, KB/s, renewal cycle: 1s
	MTU   int          // variable Bytes
	Sconn *net.UDPConn // sender's udp conn

	// receiver
	// recoder    []int64      //
	Rconn      *net.UDPConn // UDP conn
	Rstorepath string       // 接收端储存路径(文件夹、必须已存在)

	// common
	SCF time.Duration // Speed control frequency, 速度控制更新频率
	Key [16]byte      // 密钥
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
func (s *SUDP) Send(path string, startBias int64) error {

	fi, err := os.Stat(path)
	if com.Errorlog(err) {
		return err
	}
	go func() {
		for {
			fmt.Println("速度：", s.Speed)
			time.Sleep(time.Second)
		}
	}()

	if fi.Mode().IsDir() { // floder
		inf, basepath, out, err := com.GetFloderInfo(path)
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
			if err := s.sender(fh, p, s.Sconn, 0, false); err != nil {
				return err
			}
		}

	} else { // file
		fh, err := os.Open(path)
		if err != nil {
			return err
		}
		err = s.sender(fh, filepath.Base(path), s.Sconn, 0, true)
		if err != nil {
			fmt.Println(err)
		}
	}

	return s.sSendEndTranfer()

}

// Receive receive data packet
func (s *SUDP) Receive() error {

	for {
		fh, fs, raddr, transEnd, err := s.receiverStartFile(s.Rconn)
		if com.Errorlog(err) {
			return err
		}
		fmt.Println("开始, 文件大小", fs)

		if transEnd {
			fmt.Println("收到传输结束包,传输结束")
			return nil
		}

		err = s.receiver(fh, fs, raddr)
		if com.Errorlog(err) {
			continue
		}

	}
}

// Test test
func (s *SUDP) Test() {
	conn := s.Sconn
	conn.Write([]byte("sdsdsd"))
}

// resetmtu 由于封装为数据包的是否会填充，应该重置MTU
func (s *SUDP) resetmtu() {

	if s.MTU%16 != 0 {
		s.MTU = s.MTU - s.MTU%16
	}

}
