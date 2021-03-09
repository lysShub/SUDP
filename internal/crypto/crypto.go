package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"strconv"
)

/*
* en/de crypt
* !!! the key is also used as the initialization vector, less safety
 */

// CbcEncrypt encrypt
func CbcEncrypt(key []byte, p []byte) error {
	if len(key) != 16 {
		return errors.New("the secret key's length != 16")
	}

	if len(p)%16 != 0 {
		return errors.New("plaintext's length isn't integer multiples of 16，length：" + strconv.Itoa(len(p)))
	}
	block, err := aes.NewCipher(key) //key
	if err != nil {
		return err
	}

	mode := cipher.NewCBCEncrypter(block, key) // key is also used as the initialization vector
	mode.CryptBlocks(p[0:], p)

	return nil
}

// CbcDecrypt decrypt
func CbcDecrypt(key []byte, c []byte) error {
	lenData := len(c)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	if lenData%16 != 0 {
		err := errors.New("the data's length != 16; length is " + strconv.Itoa(lenData))
		if err != nil {
			return err
		}
	}
	mode := cipher.NewCBCDecrypter(block, key)
	mode.CryptBlocks(c[0:], c)

	return nil
}
