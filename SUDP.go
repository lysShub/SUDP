package sudp

import (
	"net"
	"os"
	"time"
)

// Addr duplex link
type Addr struct {
	raddr *net.UDPAddr
	laddr *net.UDPAddr
}

type sudp struct {

	// sender
	speed    int           // send/receive real time speed, KB/s, renewal cycle: 1s
	handle   chan *os.File // read file handle
	key      [16]byte      // secret key
	addr     Addr          // addr
	mtu      int           // variable Byte
	basePath string        // base path

	// receiver
	recoder []int64 // recode write, [1 3 7 9] mean write at 1-3 and 7-9;

	// common

}

// send send data packet
func (c *sudp) Send() {
	conn, _ := net.DialUDP("udp", c.addr.laddr, c.addr.raddr)

	select {
	case fh := <-c.handle:

		c.sender(fh, conn)

	case <-time.After(5 * time.Millisecond * 1000):
		return
	}
}

//
func (c *sudp) Receive() {

}
