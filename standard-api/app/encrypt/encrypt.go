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

// EncryptedValue is a type for managing ciphertext values in a flexible way.
// It is comprised of four elements:
//
//   magic bytes  version           nonce                      value
// |-------------|-------|-------------------------|---------------------------|
//
// ① magic bytes - A series of bytes used for detecting encryption state.
// ② version     - A byte representing the version of the key used to encrypt
//                  the value. This allows for migrating to new key values.
// ③ nonce       - A unique value required for decryption.
// ④ value       - The ciphertext value.
type EncryptedValue []byte

func NewEncryptedValue(version byte, nonce, value []byte) EncryptedValue {
	data := make([]byte, paddingLen+len(value))

	copy(data[0:magicBytesLen], magicBytes) // magic bytes
	data[keyVersionIndex] = version         // version
	copy(data[nonceStart:nonceEnd], nonce)  // nonce
	copy(data[nonceEnd:], value)            // value

	return EncryptedValue(data)
}

func FromByteSlice(val []byte) (EncryptedValue, error) {
	if len(val) < paddingLen {
		return nil, xerrors.New("invalid encrypted value")
	}
	if !encrypted(val) {
		return nil, xerrors.New("cannot re-encrypt encrypted value")
	}
	return EncryptedValue(val), nil
}

func encrypted(b []byte) bool {
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
