package sudp

import (
	"SUDP/internal/com"
	"net"
	"os"
)

// Addr duplex link
type Addr struct {
	raddr *net.UDPAddr
	laddr *net.UDPAddr
}

type sudp struct {

	// sender
	speed    int      // send/receive real time speed, KB/s, renewal cycle: 1s
	key      [16]byte // secret key
	addr     Addr     // addr
	mtu      int      // variable Byte
	basePath string   // base path

	// receiver
	recoder []int64 // recode write, [1 3 7 9] mean write at 1-3 and 7-9;

	// common

}

// send send data packet
func (c *sudp) Send(fh *os.File) error {
	conn, err := net.DialUDP("udp", c.addr.laddr, c.addr.raddr)
	if com.Errorlog(err) {
		return err
	}
	c.sender(fh, conn)

}

//
func (c *sudp) Receive() {

}
