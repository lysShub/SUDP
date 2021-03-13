package main

import (
	"SUDP/internal/file"
	"fmt"
	_ "net/http/pprof"
	"os"
	"runtime"
	"testing"
	"time"
)

func main() {

	runtime.GOMAXPROCS(8)

	// go Test()
	Test()
	// http.ListenAndServe(":8080", nil)

}

// TestTest assa
func TestTest(t *testing.T) {
	Test()
}

// BenchmarkTest sa
func BenchmarkTest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Test()
	}
}

// Test ss
func Test() {
	var rr [16]byte = [16]byte{
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf,
	}

	// r, _, err := packet.PackageDataPacket([]byte("sdfsaf"), 0x3FFFFF0001, rr, false)
	// com.Errorlog(err)
	// fmt.Println(r)

	// l, bias, final, err := packet.ParseDataPacket(r, rr)
	// com.Errorlog(err)
	// fmt.Println(string(r[:l]))
	// fmt.Println("bias", bias)
	// fmt.Println(final)

	fh, err := os.Open(`E:\a.mp4`)
	if err != nil {
		fmt.Println(err)
		return
	}
	var r = new(file.Rd)
	r.Fh = fh
	r.Fm = true

	wh, err := os.OpenFile(`F:\aa.mp4`, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		fmt.Println(err)
		return
	}
	var w = new(file.Wt)
	w.Fh = wh
	w.Fm = false

	a := time.Now().UnixNano()
	var i int64 = 0
	// var i int64 = 2017521000
	b := make([]byte, 1024, 1100)
	for {
		// fmt.Println("-------------------------------")

		// fmt.Println(&(b[0]))

		d, l, end, err := r.ReadFile(b, i, rr)
		// fmt.Println(&(d[0]))
		if err != nil {
			fmt.Println(err)
			return
		}

		// 原始写
		// l, bias, end, err := packet.ParseDataPacket(d, rr)
		// if err != nil {
		// 	fmt.Println(err)
		// 	return
		// }
		// _, err = wh.WriteAt(d[:l], bias)
		// if err != nil {
		// 	fmt.Println(err)
		// }

		// 写
		l, end, err = w.WriteFile(d, rr)
		if err != nil {
			fmt.Println(err)
			return
		}

		//
		if end {
			break
		}

		i = i + int64(l)

		// time.Sleep(time.Second)
	}
	bb := time.Now().UnixNano()
	fmt.Println((bb - a) / 1e9)
}
