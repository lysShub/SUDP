package file

import (
	"io"
	"os"

	"SUDP/internal/packet"
)

// read / write file

// ReadFile 读取文件数据; 返回数据包，是否最后包。
// 参数d应该有足够的容量(len+15); 否则会浪费内存。
func ReadFile(fh *os.File, d []byte, bias int64, key *[16]byte) ([]byte, int, bool, error) {

	_, err := fh.ReadAt(d, bias)
	if err != nil {
		if err == io.EOF {
			fi, err := fh.Stat()
			if err != nil {
				return nil, 0, false, err
			}
			if fi.Size()-bias == 1 {
				d = nil
				return nil, 0, true, nil
			}
			d = make([]byte, fi.Size()-bias, fi.Size()-bias+9)
			_, err = fh.ReadAt(d, bias)
			if err != nil {
				return nil, 0, false, err
			}
			return packet.PackageDataPacket(d, bias, *key, true)

		}
		return nil, 0, false, err
	}
	return packet.PackageDataPacket(d, bias, *key, false)
}

// ReadSpecifyData read specify data
func ReadSpecifyData() {

}

// WriteFile as
func WriteFile() {}
