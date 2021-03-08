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
func (s *SUDP) receiver(fh *os.File, fs int64, conn *net.UDPConn, raddr *net.UDPAddr) error {
	var rec []int64 = make([]int64, 0, 1024)
	var finalPacket, final bool = false, false
	var rd *net.UDPAddr
	var err error
	var n int
	var bias int64

	go s.sendResendPacket(conn, rec, raddr)
	go func() {
		if len(rec) == 2 && rec[1] == fs {
			final = true
			// 发送结束包
			r, _, err := packet.PackageDataPacket(nil, 0x3FFFFF0001, s.Key, false)
			com.Errorlog(err)
			_, err = conn.WriteToUDP(r, raddr)
			com.Errorlog(err)

		} else if rec[1] > fs {
			fmt.Println("怎么回事，太大了")
		}

	}()

	for final {
		var d []byte = make([]byte, s.MTU+8)
		n, rd, err = conn.ReadFromUDP(d)
		if com.Errorlog(err) {
			continue
		}
		if rd == raddr {
			n, bias, finalPacket, err = packet.ParseDataPacket(d[:n], s.Key)
			if com.Errorlog(err) {
				continue
			}
			if finalPacket {
				fmt.Println("主进程传输完成", finalPacket)
			}

			n, err = fh.WriteAt(d[:n], bias)
			com.Errorlog(err)
			// write record
			rec = s.writeRcorder(rec, bias, bias+int64(n)-1)
		}

	}
	return nil
}

// receiverStartFile return file handle, file size
func (s *SUDP) receiverStartFile(conn *net.UDPConn) (*os.File, int64, *net.UDPAddr, bool, error) {
	var flag bool = true

	go func() {
		time.Sleep(time.Second)
		flag = false
	}()

	for flag {
		var d []byte = make([]byte, s.MTU, s.MTU+15)
		n, raddr, err := conn.ReadFromUDP(d)
		if com.Errorlog(err) {
			continue
		}

		n, bias, _, err := packet.ParseDataPacket(d[:n], s.Key)
		if com.Errorlog(err) {
			continue
		}
		if bias>>16 == 0x3FFFFF && bias&0x8000 == 1 {
			fs := int64(d[0])<<32 + int64(d[1])<<24 + int64(d[2])<<16 + int64(d[3])<<8 + int64(d[4])

			path := filepath.ToSlash(s.RBasePath) + `/` + string(d[5:n])
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
		} else if bias == 0x3FFFFFFFFF {
			return nil, 0, nil, true, nil

		}

	}

	return nil, 0, nil, false, errors.New("time out")
}

// senderOfreceiver receiver's sender
func (s *SUDP) senderOfreceiver(conn *net.UDPConn) {

}

// sendResendPacket receiver send resend packet
func (s *SUDP) sendResendPacket(conn *net.UDPConn, rec []int64, raddr *net.UDPAddr) {

	for {
		var flag int64 = rec[len(rec)-1]
		time.Sleep(time.Second)
		var d []byte
		for i := 0; rec[i] <= flag; i++ {
			if i&0b1 == 1 {
				bias := rec[i] + 1
				len := rec[i+1] - rec[i] - 1
				d = append(d, uint8(bias>>32), uint8(bias>>24), uint8(bias>>16), uint8(bias>>8), uint8(bias), uint8(len>>8), uint8(len))
				i++
			}
		}
		p, _, err := packet.PackageDataPacket(d, 0x3FFFFF4000, s.Key, false)
		com.Errorlog(err)
		_, err = conn.WriteToUDP(p, raddr)
		com.Errorlog(err)
	}
}

// sendStartPacket send start packet
func (s *SUDP) sendStartPacket(conn *net.UDPConn) error {

	d, _, err := packet.PackageDataPacket(nil, 0x3FFFFF0000, s.Key, false)
	if com.Errorlog(err) {
		return err
	}
	_, err = conn.Write(d)
	if com.Errorlog(err) {
		return err
	}
	return nil
}

// writeRcorder balala
func (s *SUDP) writeRcorder(rec []int64, bias int64, endbias int64) []int64 {
	fmt.Println(len(rec))

	reclen := len(rec)
	ei := (len(rec)) - 1

	if ei == -1 { // first
		rec = append(rec, 0, endbias)
		return rec
	}
	if bias-rec[ei] == 1 {
		rec[ei] = bias + int64(endbias-bias+1) - 1
		return rec
	}
	if bias-rec[ei] > 1 {
		rec = append(rec, bias, endbias)
		// resend
		return rec
	}

	var a, b int = 0, 0
	for i := ei; i > 0; i-- {
		if rec[i] < endbias && a == 0 {
			a = i
		}
		if rec[i] < bias && b == 0 {
			b = i + 1
		}
	}
	// b -- a
	if a-b == -1 { // rebuild
		var tmp []int64 = make([]int64, b, reclen+2)
		copy(tmp, rec)
		tmp = append(tmp, bias, endbias)
		tmp = append(tmp, rec[b:]...)
		rec = tmp
	} else if b == a {
		if a&0b1 == 1 {
			if a+2 <= ei && rec[a+1]-endbias == 1 {
				var tmp []int64 = make([]int64, b, reclen-2)
				copy(tmp, rec)
				tmp = append(tmp, rec[a+2:]...)
				rec = tmp
			}
			rec[a] = endbias
		}
		if b&0b1 == 0 {
			if b-2 >= 0 && bias-rec[b-1] == 1 { //rebuid
				var tmp []int64 = make([]int64, b-2, reclen-2)
				copy(tmp, rec)
				tmp = append(tmp, rec[a:]...)
				rec = tmp
			} else {
				rec[b] = bias
			}
		}

	} else if a-b == 1 {
		if b&0b1 == 0 {
			rec[b] = bias
			rec[a] = endbias
		} else { //rebuild
			tmp := rec[:b]
			tmp = append(tmp, rec[:a+1]...)
			rec = tmp
		}
	} else { // rebuild

		var tmp []int64 = make([]int64, b, reclen)
		copy(tmp, rec)
		if b&0b1 == 0 {

			tmp = append(tmp, bias)
			if a&0b1 == 0 {
				tmp = append(tmp, rec[a+1:]...)
			} else {
				tmp = append(tmp, endbias)
				if a+1 <= ei {
					tmp = append(tmp, rec[a+1:]...)
				}
			}
			rec = tmp

		} else {
			if a&0b1 == 0 {
				tmp = append(tmp, rec[a+1:]...)
			} else {
				tmp = append(tmp, endbias)
				if a+1 <= ei {
					tmp = append(tmp, rec[a+1:]...)
				}
			}
			rec = tmp
		}
	}
	// need rebuild rec
	return rec
}
