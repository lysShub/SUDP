package sudp

import (
	"SUDP/internal/com"
	"SUDP/internal/file"
	"SUDP/internal/packet"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

func (s *sudp) sender(fh *os.File, conn *net.UDPConn, bias int64) error {
	var start, end, readyStart *bool
	var xx, yy, zz = false, false, true
	start, end, readyStart = &xx, &yy, &zz
	var rs chan []byte //resend control
	rs = make(chan []byte, 128)

	fi, _ := fh.Stat()
	name, _ := filepath.Rel(s.SBasePath, fi.Name())

	// read receiver's reply
	go s.receiverOfSender(conn, start, end, readyStart, rs)
	// send resend data
	go s.sendResendData(conn, fh, rs)

	for !*end {
		var d []byte = make([]byte, s.MTU, s.MTU+15)
		if *start { //send file data

			d, final, err := file.ReadFile(fh, d, bias, &s.Key)
			if com.Errorlog(err) {
				continue
			}
			if final {
				*start = false
				if d == nil {
					continue
				}
			}
			n, err := conn.Write(d)
			if com.Errorlog(err) {
				continue
			}

			bias = bias + int64(n)
			time.Sleep(s.speedToDelay())
		}
		if *readyStart { // send info packet
			var infop []byte
			infop = append(infop, uint8(fi.Size()>>32), uint8(fi.Size()>>24), uint8(fi.Size()>>16), uint8(fi.Size()>>8), uint8(fi.Size()))
			d, _, err := packet.PackageDataPacket(append(infop, []byte(name)...), 0, s.Key, false)
			if com.Errorlog(err) {
				continue
			}
			_, err = conn.Write(d)
			com.Errorlog(err)
			time.Sleep(time.Millisecond * 50)
		}
	}
	return nil
}

// receiverOfSender sender's receiver
// receive receiver's data ; updata controlSpeed,
func (s *sudp) receiverOfSender(conn *net.UDPConn, start *bool, end, readyStart *bool, rs chan []byte) {
	var b []byte = make([]byte, 2000)
	for {
		n, r, err := conn.ReadFromUDP(b)
		if err != nil {
			continue
		}
		if r == s.Addr.Raddr {
			_, bias, _, _ := packet.ParseDataPacket(b[:n], s.Key)
			if bias>>16 == 0x3FFFFF {
				// if bias&0x4000 == 1 { // speed control
				// 	if b[0] == 0 {
				// 		s.Speed = s.Speed - (int(b[1])<<8 + int(b[2]))
				// 	} else if b[0] == 1 {
				// 		s.Speed = s.Speed + (int(b[1])<<8 + int(b[2]))
				// 	}
				// } else
				if bias&0x2000 == 1 { //resend

				} else if bias&0x1 == 1 { // finish
					*end = true
				} else if bias|0x0 == 0 { //start
					*start = true
					*readyStart = false
				}
			}
		}
	}
}

// sendSpecifyData send resend data
// control speed
func (s *sudp) sendResendData(conn *net.UDPConn, fh *os.File, rs chan []byte) {
	var d []byte = make([]byte, 7)
	var bias, len int64
	var r int64
	go func() {
		for { // speed control strategy
			if r != 0 {
				s.Speed = s.Speed + 512 // KB/s 快增长
				fmt.Println(r)
			}
			r = 0
			time.Sleep(s.SCF)
		}
	}()
	for {
		d = <-rs
		bias = int64(d[5])<<40 + int64(d[4])<<32 + int64(d[3])<<24 + int64(d[2])<<16 + int64(d[1])<<8 + int64(d[0])
		len = int64(d[7])<<8 + int64(d[6])
		r = r + len // rcorde resend data size

		p := make([]byte, len, len+15)
		p, _, err := file.ReadFile(fh, p, bias, &s.Key)
		if com.Errorlog(err) {
			continue
		}
		_, err = conn.Write(p)
		com.Errorlog(err)
	}

}

// speedToDelay  microseconds
func (s *sudp) speedToDelay() time.Duration {
	return time.Duration(1000000 * s.MTU / s.Speed)
}
