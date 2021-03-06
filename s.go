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

func (c *sudp) sender(fh *os.File, conn *net.UDPConn) error {
	var start, end *bool
	var ii, jj bool = false, false
	start, end = &ii, &jj
	var rs chan []byte
	rs = make(chan []byte, 128)

	fi, _ := fh.Stat()
	name, _ := filepath.Rel(c.basePath, fi.Name())

	// read receiver's reply
	go c.receiverOfSender(conn, start, end, rs)
	// send resend data
	go c.sendResendData(conn, fh, rs)

	var i int64 = 0
	for !*end {
		var d []byte = make([]byte, c.mtu, c.mtu+10)
		if *start { //send file data

			d, final, err := file.ReadFile(fh, d, i, &c.key)
			if com.Errorlog(err) {
				continue
			}
			if final {
				*start = false
				if d == nil {
					continue
				}
				continue
			}
			n, err := conn.Write(d)
			if com.Errorlog(err) {
				continue
			}

			i = i + int64(n)
			time.Sleep(c.speedToDelay())
		} else { // send info packet
			var infop []byte
			infop = append(infop, uint8(fi.Size()>>24), uint8(fi.Size()>>16), uint8(fi.Size()>>8), uint8(fi.Size()))
			d, _, err := packet.PackageDataPacket(append(infop, []byte(name)...), 0, c.key, false)
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
func (c *sudp) receiverOfSender(conn *net.UDPConn, start *bool, end *bool, rs chan []byte) {
	var b []byte = make([]byte, 2000)
	for {
		n, r, err := conn.ReadFromUDP(b)
		if err != nil {
			continue
		}
		if r == c.addr.raddr {
			_, bias, _, _ := packet.ParseDataPacket(b[:n], c.key)
			if bias>>16 == 0x3FFFFF {
				if bias&0x4000 == 1 { // speed control

					if b[0] == 0 {
						c.speed = c.speed - (int(b[1])<<8 + int(b[2]))
					} else if b[0] == 1 {
						c.speed = c.speed + (int(b[1])<<8 + int(b[2]))
					}

				} else if bias&0x2000 == 1 { //resend

				} else if bias&0x1 == 1 { // finish
					*end = true
				} else if bias|0x0 == 0 { //start
					*start = true
				}
			}
		}
	}
}

// sendSpecifyData send resend data
// control speed
func (c *sudp) sendResendData(conn *net.UDPConn, fh *os.File, rs chan []byte) {
	var d []byte = make([]byte, 7)
	var bias, len int64
	var r int64
	go func() {
		for {
			time.Sleep(time.Second)
			c.speed = c.speed + 1024 // 快增长
			fmt.Println(r)
			r = 0
		}
	}()
	for {

		d = <-rs
		bias = int64(d[5])<<40 + int64(d[4])<<32 + int64(d[3])<<24 + int64(d[2])<<16 + int64(d[1])<<8 + int64(d[0])
		len = int64(d[7])<<8 + int64(d[6])
		r = r + len

		p := make([]byte, len, len+15)
		p, _, err := file.ReadFile(fh, p, bias, &c.key)
		if com.Errorlog(err) {
			continue
		}
		_, err = conn.Write(p)
		com.Errorlog(err)
	}

}

// speedToDelay  microseconds
func (c *sudp) speedToDelay() time.Duration {
	return time.Duration(1000000 * c.mtu / c.speed)
}
