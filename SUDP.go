package sudp

import (
	"SUDP/internal/packet"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

// Addr duplex link
type Addr struct {
	raddr *net.UDPAddr
	laddr *net.UDPAddr
}

type sudp struct {

	// sender
	realSpeed    int           // send/receive real time speed, KB/s
	controlSpeed int           // send control speed, renewal cycle: 1s
	handle       chan *os.File // read file handle
	key          [16]byte      // secret key
	addr         Addr          // addr
	mtu          int           // variable Byte
	basePath     string        // base path
	flag         bool          //link useable
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

func (c *sudp) sender(fh *os.File, conn *net.UDPConn) {
	fi, _ := fh.Stat()

	name, err := filepath.Rel(c.basePath, fi.Name())

	go c.senderSreader(conn)

	for {
		if c.flag {

		} else {
			packet.PackageDataPacket([]byte(name))
			conn.Write()

		}
	}

}

// senderSreader sender's reader
// receive receiver's data ; updata controlSpeed,
func (c *sudp) senderSreader(conn *net.UDPConn) {
	var b []byte = make([]byte, 2000)
	for {
		n, r, err := conn.ReadFromUDP(b)
		if err != nil {
			continue
		}
		if r == c.addr.raddr {
			n, bias, _, _ := packet.ParseDataPacket(b[:n], c.key)
			if bias > 100 {
				fmt.Println(b[:n])
			}
		}
	}
}

// speedToDelay  microseconds
func (c *sudp) speedToDelay() time.Duration {
	return time.Duration(1000000 * c.mtu / c.controlSpeed)
}
