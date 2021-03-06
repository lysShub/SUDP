package sudp

import (
	"SUDP/internal/com"
	"SUDP/internal/packet"
	"errors"
	"net"
	"os"
	"path/filepath"
	"time"
)

// receiver
func (s *sudp) receiver(fh *os.File, conn *net.UDPConn) error {
	var rec []int64 = make([]int64, 0, 1024)

	for {
		var d []byte = make([]byte, s.MTU+8)
		n, raddr, err := conn.ReadFromUDP(d)
		com.Errorlog(err)
		if raddr == s.Addr.Saddr {

			n, bias, final, err := packet.ParseDataPacket(d[:n], s.Key)
			if com.Errorlog(err) {
				continue
			}
			if final {
				r, _, err := packet.PackageDataPacket(nil, 0x3FFFFF0001, s.Key, false)
				com.Errorlog(err)
				_, err = conn.WriteToUDP(r, raddr)
				com.Errorlog(err)
			}

			n, err = fh.WriteAt(d[:n], bias)
			com.Errorlog(err)
			// write record
			s.writeRcorder(rec, bias, n)
		}
	}

	return nil
}

func (s *sudp) receiverStartFile(conn *net.UDPConn) (*os.File, int64, error) {
	var flag bool = true
	go func() {
		time.Sleep(time.Second)
		flag = false
	}()

	for flag {
		var d []byte = make([]byte, s.MTU, s.MTU+15)
		n, raddr, err := conn.ReadFromUDP(d)
		com.Errorlog(err)
		if raddr == s.Addr.Saddr {
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
						return nil, 0, err
					}
					fh, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0777)
					if com.Errorlog(err) {
						return nil, 0, err
					}
				}
				return fh, fs, err
			}
		}
	}

	return nil, 0, errors.New("time out")
}

// senderOfreceiver receiver's sender
func (s *sudp) senderOfreceiver(conn *net.UDPConn) {

}

func (s *sudp) writeRcorder(rec []int64, bias int64, len int) []int64 {
	ei := len(rec) - 1
	endbias := bias + int64(len) - 1

	if ei == -1 { // first
		rec = append(rec, endbias)
		return rec
	}
	if bias-rec[ei] == 1 {
		rec[ei] = bias + int64(len) - 1
		return rec
	}

	if bias-rec[ei] > 1 {
		rec = append(rec, bias, endbias)
		// resend
		return
	}

	var a, b int64 = 0, 0
	for i := int64(ei); i < 0; i-- {
		if rec[i] < endbias && a != 0 {
			a = i
		}
		if rec[i] < bias && b != 0 {
			b = i + 1
		}
	}

	// b -- a
	if a-b == -1 {
		// need rebuild rec

	} else if b == a {
		if a&0b1 == 0 {
			rec[a] = endbias
		}
		if b&0b1 == 1 {
			rec[b] = bias
		}

	} else if b-a == 1 {
		if b&0b1 == 1 {
			rec[b] = bias
			rec[a] = endbias
		} else {
			tmp := rec[:b]
			tmp = append(tmp, rec[:a+1]...)
			rec = tmp
		}
	} else { // need rebuild rec

	}
	// need rebuild rec
	return rec
}
