package sudp

import (
	"SUDP/internal/com"
	"SUDP/internal/file"
	"SUDP/internal/packet"
	"errors"
	"fmt"
	"net"
	"os"
	"time"
)

func (s *SUDP) sender(fh *os.File, name string, conn *net.UDPConn, bias int64, endfile bool) error {

	if err := s.sstart(fh, name, conn, bias); err != nil {
		return err
	}
	var end *bool
	var e bool = false
	end = &e

	rs := make(chan []byte, 128) // 重发数据信息管道
	go s.sreceiver(conn, end, rs)
	go s.sendResendData(conn, fh, rs)

	var d []byte
	for {
		d = make([]byte, s.MTU, s.MTU+25)

		d, n, final, err := file.ReadFile(fh, d, bias, &s.Key)
		if com.Errorlog(err) {
			continue
		}

		_, err = conn.Write(d)
		if com.Errorlog(err) {
			continue
		}

		bias = bias + int64(n)
		time.Sleep(s.speedToDelay())
		if final {
			fmt.Println("发送了最后包")
			break
		}
	}

	for {
		if *end {
			break
		}
		time.Sleep(time.Second)
	}

	return nil
}

func (s *SUDP) sstart(fh *os.File, name string, conn *net.UDPConn, bias int64) error {

	fi, _ := fh.Stat()
	var infop []byte
	infop = append(infop, uint8(fi.Size()>>32), uint8(fi.Size()>>24), uint8(fi.Size()>>16), uint8(fi.Size()>>8), uint8(fi.Size()))
	d, _, _, err := packet.PackageDataPacket(append(infop, []byte(name)...), 0x3FFFFF0000, s.Key, false)
	if err != nil {
		return err
	}
	_, err = conn.Write(d)
	if err != nil {
		return err
	}
	// 接收开始包
	for i := 0; i < 16; i++ {
		conn.SetReadDeadline(time.Now().Add(time.Second))
		n, err := conn.Read(d)
		if err != nil { // 超时
			return err
		}
		_, bias, _, err = packet.ParseDataPacket(d[:n], s.Key)
		if err != nil { // 解析包出错
			return err
		}
		if bias == 0x3FFFFF0000 {
			return nil
		}
	}
	return errors.New("exception")
}

// sreceiver 接收发送端的数据
func (s *SUDP) sreceiver(conn *net.UDPConn, end *bool, rs chan []byte) {
	var d []byte
	for {

		d = make([]byte, s.MTU+25)
		conn.SetReadDeadline(time.Now().Add(time.Minute)) //重置
		n, err := conn.Read(d)
		if com.Errorlog(err) {
			continue
		}
		n, bias, _, err := packet.ParseDataPacket(d[:n], s.Key)
		if com.Errorlog(err) {
			continue
		}
		if bias == 0x3FFFFF4000 { //重发包
			for i := 0; i < n/7; i++ {
				rs <- d[i*7 : (i+1)*7]
			}
		} else if bias == 0x3FFFFF0001 { //结束包
			fmt.Println("收到文件传输结束包")
			*end = true
		}

	}

}

// 发送重发数据包
// 根据重发数据包数量进行速度控制
func (s *SUDP) sendResendData(conn *net.UDPConn, fh *os.File, rs chan []byte) {
	var d []byte = make([]byte, 7)
	var bias, len int64
	var resendLen int64 //上一秒重复数据大小
	go func() {
		for { // speed control strategy
			if resendLen != 0 {
				s.Speed = s.Speed + 512 // KB/s 快增长
			}
			s.Speed = s.Speed + 512 // KB/s 快增长
			resendLen = 0
			time.Sleep(s.SCF)
		}
	}()
	for {
		d = <-rs
		bias = int64(d[5])<<40 + int64(d[4])<<32 + int64(d[3])<<24 + int64(d[2])<<16 + int64(d[1])<<8 + int64(d[0])
		len = int64(d[7])<<8 + int64(d[6])
		resendLen = resendLen + len // rcorde resend data size

		p := make([]byte, len, len+25)
		p, _, _, err := file.ReadFile(fh, p, bias, &s.Key)
		if com.Errorlog(err) {
			continue
		}
		_, err = conn.Write(p)
		com.Errorlog(err)
	}

}

// speedToDelay  获取延时(毫秒)
func (s *SUDP) speedToDelay() time.Duration {
	return time.Duration(1000000 * s.MTU / s.Speed)
}

func (s *SUDP) sSendEndTranfer() error {
	conn := s.Sconn
	var d []byte = make([]byte, 0)

	d, _, _, err := packet.PackageDataPacket(nil, 0x3FFFFFFFFF, s.Key, false)
	fmt.Println("发送了传输任务结束包")
	if err != nil {
		return err
	}
	for i := 0; i < 4; i++ {
		_, err = conn.Write(d)
		time.Sleep(time.Millisecond * 100)
	}
	return err
}
