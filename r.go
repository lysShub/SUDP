package sudp

import (
	"SUDP/internal/com"
	"SUDP/internal/packet"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

// receiver
func (s *SUDP) receiver(fh *os.File, fs int64, raddr *net.UDPAddr) error {
	var err error

	err = s.sReplyStart(raddr)
	fmt.Println("回复了开始包")
	if err != nil {
		return err
	}

	// 写入记录
	var rec []int64 = make([]int64, 0, 1024)

	// 时候接收到最后数据包，此文件时候传输完成
	var finalPacket, final bool = false, false
	var rd *net.UDPAddr
	var n int
	var bias int64
	var conn *net.UDPConn = s.Rconn

	// 通过查验写入记录，请求需要重发的数据包
	go s.sendResendPacket(conn, &rec, raddr)

	// 检查文件是否传输完成
	go func() {
		for {
			time.Sleep(time.Second)
			// fmt.Println("rec", rec)
			if len(rec) == 2 && rec[0] == 0 && rec[1] == fs-1 {
				final = true
				// 发送结束包 一个文件传输完成
				r, _, _, err := packet.PackageDataPacket(nil, 0x3FFFFF0001, s.Key, false)
				com.Errorlog(err)
				_, err = conn.WriteToUDP(r, raddr)
				com.Errorlog(err)

			}
			// else if rec[len(rec)-1] > fs {
			// 	fmt.Println("怎么回事，太大了")
			// }
			fmt.Println(rec)
		}
	}()

	for !final {

		var d []byte = make([]byte, s.MTU+25)
		n, rd, err = conn.ReadFromUDP(d)
		if com.Errorlog(err) {
			continue
		}
		if rd.IP.Equal(raddr.IP) && rd.Port == raddr.Port {

			n, bias, finalPacket, err = packet.ParseDataPacket(d[:n], s.Key)
			if com.Errorlog(err) {
				continue
			}
			if finalPacket {
				fmt.Println("接收到最后包", finalPacket)
			}

			n, err = fh.WriteAt(d[:n], bias)
			com.Errorlog(err)

			// 更新写入记录器
			rec, _ = s.writeRcorder(rec, bias, bias+int64(n)-1)
		}

	}
	fmt.Println("文件传输完成，退出")
	return nil
}

// sendResendPacket 请求重发数据包
func (s *SUDP) sendResendPacket(conn *net.UDPConn, recorder *[]int64, raddr *net.UDPAddr) {

	for {
		rec := *recorder
		if len(rec) <= 2 {
			time.Sleep(time.Second)
			continue
		}
		var flag int64 = rec[len(rec)-1]
		time.Sleep(time.Second)
		var d []byte
		for i := 0; rec[i] < flag; {
			if i&0b1 == 1 {
				bias := rec[i] + 1
				len := rec[i+1] - rec[i] - 1
				d = append(d, uint8(bias>>32), uint8(bias>>24), uint8(bias>>16), uint8(bias>>8), uint8(bias), uint8(len>>8), uint8(len))
				i = i + 7
			}
			if i >= 520 { // 数据包太大l
				break
			}
		}
		p, _, _, err := packet.PackageDataPacket(d, 0x3FFFFF4000, s.Key, false)
		com.Errorlog(err)
		_, err = conn.WriteToUDP(p, raddr)
		com.Errorlog(err)
	}
}

// sReplyStart 回复开始包
func (s *SUDP) sReplyStart(raddr *net.UDPAddr) error {
	conn := s.Rconn
	d, _, _, err := packet.PackageDataPacket(nil, 0x3FFFFF0000, s.Key, false)
	if err != nil {
		return err
	}
	_, err = conn.WriteToUDP(d, raddr)
	if err != nil {
		return err
	}
	return nil
}

// receiverStartFile 接收文件信息包或传输任务结束包
func (s *SUDP) receiverStartFile(conn *net.UDPConn) (*os.File, int64, *net.UDPAddr, bool, error) {
	var flag bool = true

	go func() {
		time.Sleep(time.Second)
		flag = false
	}()

	for flag {
		var d []byte = make([]byte, s.MTU+25)
		// 设置等待超时
		n, raddr, err := conn.ReadFromUDP(d)
		if com.Errorlog(err) {
			continue
		}

		n, bias, _, err := packet.ParseDataPacket(d[:n], s.Key)
		if com.Errorlog(err) {
			continue
		}
		if bias == 0x3FFFFF0000 { // 接收到文件信息包

			fs := int64(d[0])<<32 + int64(d[1])<<24 + int64(d[2])<<16 + int64(d[3])<<8 + int64(d[4])
			fmt.Println("接收到文件信息包", fs, string(com.ToUtf8(d[6:n])))

			path := s.Rstorepath + `/` + string(d[6:n])
			fh, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0777)
			if os.IsNotExist(err) {
				err = os.MkdirAll(filepath.Dir(path), 0666)
				if com.Errorlog(err) {
					return nil, 0, nil, false, err
				}
				fh, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0777)
				if com.Errorlog(err) {
					return nil, 0, nil, false, err
				}
			}
			return fh, fs, raddr, false, err
		} else if bias == 0x3FFFFFFFFF { //结束传输任务
			fmt.Println("接收到传输任务结束包")
			return nil, 0, nil, true, nil
		}

	}

	return nil, 0, nil, false, errors.New("time out")
}

// writeRcorder 文件写入记录器
// 如果有覆盖写入，返回bool为true
func (s *SUDP) writeRcorder(rec []int64, start, end int64) ([]int64, bool) {
	var l int = len(rec)
	if l == 0 {
		rec = append(rec, start, end)
		return rec, false
	}

	var tmp []int64 = make([]int64, 0, l)
	var ex bool = false
	if rec[l-1]+1 == start { //	//绝大多数情况
		tmp = rec
		tmp[l-1] = end
	} else { //覆盖所有情况

		var max func(x, y int64) int64 = func(x, y int64) int64 {
			if x > y {
				return x
			}
			return y
		}
		var min func(x, y int64) int64 = func(x, y int64) int64 {
			if x < y {
				return x
			}
			return y
		}
		var merged bool = false

		for i := 0; i < l; i = i + 2 {
			if rec[i]-1 > end {
				if !merged {
					tmp = append(tmp, start, end)
					merged = true
				}
				tmp = append(tmp, rec[i], rec[i+1])
			} else if rec[i+1]+1 < start {
				tmp = append(tmp, rec[i], rec[i+1])
			} else { //有重复区间
				start = min(start, rec[i])
				end = max(end, rec[i+1])
				ex = true
			}
		}
		if !merged {
			tmp = append(tmp, start, end)
		}
	}
	return tmp, ex
}
