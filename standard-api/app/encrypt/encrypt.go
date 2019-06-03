package encrypt

import (
	"bytes"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/xerrors"
)

var (
	magicBytes      = []byte{0xBE, 0xEF}
	magicBytesLen   = len(magicBytes)
	keyVersionIndex = magicBytesLen
	keyVersionLen   = 1
	nonceLen        = chacha20poly1305.NonceSizeX
	nonceStart      = magicBytesLen + keyVersionLen
	nonceEnd        = nonceStart + nonceLen
	paddingLen      = magicBytesLen + keyVersionLen + nonceLen
)

type EncryptedValue []byte

func NewEncryptedValue(keyVersion byte, nonce, val []byte) EncryptedValue {
	data := make([]byte, paddingLen+len(val))

	copy(data[0:magicBytesLen], magicBytes) // magic bytes
	data[keyVersionIndex] = keyVersion      // key version
	copy(data[nonceStart:nonceEnd], nonce)  // nonce
	copy(data[nonceEnd:], val)              // value

	return EncryptedValue(data)
}

func FromByteSlice(val []byte) (EncryptedValue, error) {
	if len(val) < paddingLen {
		return nil, xerrors.New("invalid encrypted value")
	}
	if !Encrypted(val) {
		return nil, xerrors.New("cannot re-encrypt encrypted value")
	}
	return EncryptedValue(val), nil
}

func Encrypted(b []byte) bool {
	return bytes.Equal(b[0:magicBytesLen], magicBytes)
}

func (v EncryptedValue) String() string {
	return string(v)
}

func (v EncryptedValue) KeyVersion() byte {
	return v[keyVersionIndex]
}

func (v EncryptedValue) Nonce() []byte {
	return v[nonceStart:nonceEnd]
}

func (v EncryptedValue) Value() []byte {
	return v[nonceEnd:]
}
