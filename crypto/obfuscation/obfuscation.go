package obfuscation

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"io"
)

const cipherKeySize = 16

func newCipher(cipherKey []byte) cipher.Stream {
	hash := sha1.Sum(cipherKey)
	block, err := aes.NewCipher(hash[:cipherKeySize])
	if err != nil {
		panic(err.Error())
	}
	return cipher.NewCTR(block, hash[:aes.BlockSize])
}

func NewEncodeDecoder(rw io.ReadWriter, cipherKey []byte) io.ReadWriter {
	return &struct {
		io.Reader
		io.Writer
	}{
		NewDecoder(rw, cipherKey),
		NewEncoder(rw, cipherKey),
	}
}

func NewEncoder(w io.Writer, cipherKey []byte) io.Writer {
	return &cipher.StreamWriter{S: newCipher(cipherKey), W: w}
}

func NewDecoder(r io.Reader, cipherKey []byte) io.Reader {
	return &cipher.StreamReader{S: newCipher(cipherKey), R: r}
}
