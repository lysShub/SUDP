package file

import (
	"io"
	"os"

	"SUDP/internal/packet"
)

// read / write file

// ReadFile retrun read data len
// parmeter b should has enougeh cap(len+8)
// parmeter b become full packet
// if len(retur []byte) < (parmeter []byt) mean finall packet
// if return nil,nil; mean last packet exactly is finall packet
func ReadFile(fh *os.File, d []byte, bias int64, key *[16]byte) ([]byte, error) {

	_, err := fh.ReadAt(d, bias)
	if err != nil {
		if err == io.EOF {
			fi, err := fh.Stat()
			if err != nil {
				return nil, err
			}
			if fi.Size()-bias == 1 {
				return nil, nil
			}
			d = make([]byte, fi.Size()-bias, fi.Size()-bias+8)
			_, err = fh.ReadAt(d, bias)
			if err != nil {
				return nil, err
			}
			return packet.PackageDataPacket(d, bias, *key, true)

		} else {
			return nil, err
		}

	}
	return packet.PackageDataPacket(d, bias, *key, false)
}

func WriteFile()
