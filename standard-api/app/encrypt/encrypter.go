package encrypt

import (
	"crypto/cipher"
	"crypto/rand"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/xerrors"
)

type (
	KeyVersion byte
	Key        []byte
	KeyMap     map[KeyVersion]Key
	aeadMap    map[KeyVersion]cipher.AEAD
)

// Encrypter is a type for managing encryption with an AEAD cipher.
// It supports refreshing keys and prevents re-encrypting of encrypted values.
type Encrypter struct {
	aeadMap aeadMap
	version KeyVersion
}

// NewEncrypter create an Encrypter that manages an AEAD cipher instance
// for each key in keyMap, where version is used as the default key version.
func NewEncrypter(version KeyVersion, keyMap KeyMap) (*Encrypter, error) {
	if _, ok := keyMap[version]; !ok {
		return nil, xerrors.Errorf("key not provided for version %d", version)
	}

	aeadMap := make(aeadMap)
	for version, key := range keyMap {
		aead, err := chacha20poly1305.NewX(key)
		if err != nil {
			return nil, xerrors.Errorf("Failed to instantiate XChaCha20-Poly1305 with given key: %w", err)
		}
		aeadMap[version] = aead
	}

	return &Encrypter{
		aeadMap: aeadMap,
		version: version,
	}, nil
}

func (m *Encrypter) Encrypt(b []byte) (EncryptedValue, error) {
	if encrypted(b) {
		return nil, xerrors.New("value is already encrypted")
	}

	nonce := make([]byte, chacha20poly1305.NonceSizeX)
	if _, err := rand.Read(nonce); err != nil {
		return nil, xerrors.Errorf("error generating nonce: %w")
	}

	ciphertext := m.aeadMap[m.version].Seal(nil, nonce, b, nil)
	return NewEncryptedValue(byte(m.version), nonce, ciphertext), nil
}

func (m *Encrypter) Decrypt(b []byte) ([]byte, error) {
	ev, err := EncryptedValueFromByteSlice(b)
	if err != nil {
		return nil, xerrors.Errorf("could not decrypt string: %w", err)
	}

	aead, ok := m.aeadMap[KeyVersion(ev.KeyVersion())]
	if !ok {
		return nil, xerrors.Errorf("unknown keyVersion %d during decryption", ev.KeyVersion())
	}

	plainText, err := aead.Open(nil, ev.Nonce(), ev.Value(), nil)
	if err != nil {
		return nil, xerrors.Errorf("failed to decrypt or authenticate value: %w", err)
	}

	return plainText, nil
}
