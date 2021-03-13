package file

import (
	"fmt"
	"io"
	"os"

	"SUDP/internal/com"
	"SUDP/internal/packet"
)

// Rd  Read 文件读取
type Rd struct {
	// 文件句柄
	Fh *os.File
	// 文件大小
	fs int64
	// 初始化标志
	initflag bool

	// fast mode 快速读取模式，单次读取较大的文件到内存中
	// 建议大文件下启用
	// 固态硬盘中，2GB文件快速读取8s，普通读取35s
	// 机械磁盘中，2GB文件快速读取10s，普通读取46s
	// 一般来说快速读取的瓶颈在CPU封包处(只能使用到单核性能)
	Fm bool
	// block size 快速读取模式下的暂存数据块大小
	bs int64
	// 储存数据块
	coer []byte
	// 记录coer中数据的位置
	rang [2]int64
}

// init 初始化函数，会覆盖
func (s *Rd) init() {
	if !s.initflag {
		fmt.Println("启动")
		fi, err := s.Fh.Stat()

		if com.Errorlog(err) {
			return
		}
		s.fs = int64(fi.Size())     // 初始化文件大小
		s.bs = 4194304              //4194304 4MB
		s.coer = make([]byte, s.bs) // 初始化coer

		s.initflag = true
	}
}
func ReadFile(fh *os.File, d []byte, bias int64, key *[16]byte) ([]byte, int, bool, error) {
	return nil, 0, false, nil
}

// readfile 任意读取，适配最后一包
func (s *Rd) readfile(fh *os.File, d []byte, bias int64, key [16]byte) ([]byte, int, bool, error) {

	_, err := fh.ReadAt(d, bias)
	if err != nil {
		if err == io.EOF {
			if s.fs-bias == 1 {
				d = nil
				return nil, 0, true, nil
			}
			d = make([]byte, s.fs-bias, s.fs-bias+9)
			_, err = fh.ReadAt(d, bias)
			if err != nil {
				return nil, 0, false, err
			}
			return packet.PackageDataPacket(d, bias, key, true)

		}
		return nil, 0, false, err
	}
	return packet.PackageDataPacket(d, bias, key, false)
}

// ReadFile 读取文件；返回：打包好数据包，原始数据长度，是否最后包。
// 参数d应该有足够的容量(len+15); 否则会浪费内存。
func (s *Rd) ReadFile(d []byte, bias int64, key [16]byte) ([]byte, int, bool, error) {
	s.init()

	// 启用快速读取模式
	if s.Fm {
		if bias < s.rang[0] { // 重发数据包
			return s.readfile(s.Fh, d, bias, key)
		}

		l := int64(len(d))
		if s.rang[0] > bias || s.rang[1] < bias+l-1 {
			_, err := s.Fh.ReadAt(s.coer, bias)
			if err != nil {
				if err == io.EOF { // 剩余文件不足以读取16MB的数据块
					return s.readfile(s.Fh, d, bias, key)
				} else if com.Errorlog(err) {
					return nil, 0, false, err
				}
			} else { // 正常读取
				s.rang[0], s.rang[1] = bias, bias+s.bs-1 // -1
				// fmt.Println(&(s.coer[0]), &(s.coer[16777215]))
			}

		}
		copy(d, s.coer[bias-s.rang[0]:])

		// 16MB数据块恰好读完文件数据，且此数据包恰好读完数据块中最后数据
		if bias+l+1 == s.fs {
			return packet.PackageDataPacket(d, bias, key, true)
		}
		return packet.PackageDataPacket(d, bias, key, false)

	}

	// 不启用快速读取模式
	return s.readfile(s.Fh, d, bias, key)
}

// Wt Write 文件写入，传入标准数据包即可
// 传入的正常数据一定要写
type Wt struct {
	// 文件句柄
	Fh *os.File
	// 初始化标志
	initflag bool

	// Fm fast mode 快速写入模式；将接收到的数据存储到内存中至一定值再写入文件
	// 大文件传输开启
	// 固态硬盘中，非快速写入3GB数据耗时55s，快速写入耗时19s，仅解包不写入耗时19s(CPU已成为瓶颈)
	// 机械磁盘中，非快速写入3GB数据耗时67s，快速写入耗时26s
	Fm bool
	// block size 快速写入模式下的暂存数据块大小
	bs int64
	// 储存数据块，存入暂存数据必须连续且小于
	coer []byte
	// 记录coer中数据的位置
	rang [2]int64
	// 记录coer中有效数据长度
	rbias int
}

// init 初始化函数
func (s *Wt) init() {
	if !s.initflag {
		fmt.Println("启动")
		if s.Fm {
			s.bs = 4194304
			s.coer = make([]byte, s.bs, s.bs)
		}
		s.rang = [2]int64{
			0, s.bs,
		}
		s.initflag = true
	}
}

// WriteFile 写入文件，传入完整数据包，返回原始数据长度，是否最后包
// 凡是解包正确的必须写入
func (s *Wt) WriteFile(p []byte, key [16]byte) (int, bool, error) {
	s.init()

	dl, bias, end, err := packet.ParseDataPacket(p, key)
	if err != nil {
		return 0, false, err
	}

	var l int64 = int64(dl)
	if s.Fm {
		if s.rang[1] < bias+l-1 || end { //重置缓存块
			// 写入
			_, err = s.Fh.WriteAt(s.coer[:s.rbias], s.rang[0])

			// 重置
			s.rang[0] = bias
			s.rang[1] = bias + s.bs
			copy(s.coer[0:l], p[:dl])
			s.rbias = dl
			if end { // 清空缓存
				_, err = s.Fh.WriteAt(s.coer[:s.rbias], s.rang[0])
			}

		} else {
			if bias >= s.rang[0] { //存入缓存块
				copy(s.coer[bias-s.rang[0]:bias-s.rang[0]+l], p[:dl])
				s.rbias = s.rbias + dl
			} else { // 重发的数据包 直接写入
				_, err = s.Fh.WriteAt(p[:dl], bias)
			}
		}
	} else {
		_, err = s.Fh.WriteAt(p[:dl], bias)
	}
	return dl, end, err
}
