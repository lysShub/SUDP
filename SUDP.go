package sudp

import (
	"SUDP/internal/com"
	"fmt"
	"net"
	"os"
	"time"
)

// Addr duplex link
type Addr struct {
	Saddr       *net.UDPAddr // sender's WAN addr
	Raddr       *net.UDPAddr // receiver's WAN addr
	SLocalPort  uint16       // sender's Lan IP
	RLocaloPort uint16       // receiver's Lan IP
}

type sudp struct {

	// sender
	Speed     int          // send/receive real time speed, KB/s, renewal cycle: 1s
	Key       [16]byte     // secret key
	Addr      Addr         // addr
	MTU       int          // variable Byte
	SBasePath string       // sender's base path
	Sconn     *net.UDPConn // sender's udp conn

	// receiver
	recoder   []int64      // recode write, [1 3 7 9] mean write at 1-3 and 7-9;
	Rconn     *net.UDPConn // receiver's udp conn
	RBasePath string       //receiver's base path

	// common
	SCF time.Duration // Speed control frequency
}

func (s *sudp) Sinit(base string, conn *net.UDPConn) {
	s.Speed = 1024 // 1MB/s
	s.Key = [16]byte{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	s.MTU = 1372
	s.SBasePath = base
	s.Sconn = conn
	s.SCF = time.Duration(time.Second)

}

// send send data packet
func (s *sudp) Send(fh *os.File, startBias int64) error {
	// laddr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(int(c.addr.sLocalPort)))
	// if com.Errorlog(err) {
	// 	return err
	// }
	// conn, err := net.DialUDP("udp", laddr, c.addr.raddr)
	// if com.Errorlog(err) {
	// 	return err
	// }
	return s.sender(fh, s.Sconn, 0)

}

// Receive receive data packet
func (s *sudp) Receive() {

	for {
		fh, fs, err := s.receiverStartFile(s.Rconn)
		fmt.Println("开始, 文件大小", fs)

		if com.Errorlog(err) {
			continue
		}
		err = s.receiver(fh, fs, s.Rconn)
		if com.Errorlog(err) {
			continue
		}
	}

}
