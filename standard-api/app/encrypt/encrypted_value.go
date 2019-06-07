package encrypt

import (
	"bytes"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/xerrors"
)

var (
	magicBytes    = []byte{0xBE, 0xEF}
	magicBytesLen = len(magicBytes)

	keyVersionIndex = magicBytesLen
	keyVersionLen   = 1

	nonceLen   = chacha20poly1305.NonceSizeX
	nonceStart = magicBytesLen + keyVersionLen

	valueStart = nonceStart + nonceLen

	paddingLen = magicBytesLen + keyVersionLen + nonceLen
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
// ③ nonce       - A unique value required for both encryption and decryption.
//                  It must be generated uniquely for each ciphertext value.
// ④ value       - The ciphertext value.
type EncryptedValue []byte

// NewEncryptedValue created an EncryptedValue from its constituent parts.
func NewEncryptedValue(version byte, nonce, value []byte) EncryptedValue {
	data := make([]byte, paddingLen+len(value))

	copy(data[0:magicBytesLen], magicBytes)  // magic bytes
	data[keyVersionIndex] = version          // version
	copy(data[nonceStart:valueStart], nonce) // nonce
	copy(data[valueStart:], value)           // value

	return EncryptedValue(data)
}

// EncryptedValueFromByteSlice creates a validated EncryptedValue from a byte slice.
func EncryptedValueFromByteSlice(b []byte) (EncryptedValue, error) {
	if len(b) < paddingLen {
		return nil, xerrors.New("byte slice does not contain encryption padding data")
	}
	if !encrypted(b) {
		return nil, xerrors.New("cannot decrypt unencrypted value")
	}
	return EncryptedValue(b), nil
}

// encrypted is true if the value stored in b is currently encrypted.
func encrypted(b []byte) bool {
	return len(b) >= paddingLen && bytes.Equal(b[0:magicBytesLen], magicBytes)
}

func (v EncryptedValue) String() string {
	return string(v)
}

func (v EncryptedValue) KeyVersion() byte {
	return v[keyVersionIndex]
}

func (v EncryptedValue) Nonce() []byte {
	return v[nonceStart:valueStart]
}

func (v EncryptedValue) Value() []byte {
	return v[valueStart:]
}
