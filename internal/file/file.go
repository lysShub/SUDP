package file

import (
	"io"
	"os"

	"SUDP/internal/packet"
)

// read / write file

// ReadFile retrun read data len
// parmeter d should has enougeh cap(len+8)
// return: packet; finally packet; error
func ReadFile(fh *os.File, d []byte, bias int64, key *[16]byte) ([]byte, bool, error) {

	_, err := fh.ReadAt(d, bias)
	if err != nil {
		if err == io.EOF {
			fi, err := fh.Stat()
			if err != nil {
				return nil, false, err
			}
			if fi.Size()-bias == 1 {
				d = nil
				return nil, true, nil
			}
			d = make([]byte, fi.Size()-bias, fi.Size()-bias+9)
			_, err = fh.ReadAt(d, bias)
			if err != nil {
				return nil, false, err
			}
			return packet.PackageDataPacket(d, bias, *key, true)

		}
		return nil, false, err
	}
	return packet.PackageDataPacket(d, bias, *key, false)
}

// ReadSpecifyData read specify data
func ReadSpecifyData() {

}

// WriteFile as
func WriteFile() {}
