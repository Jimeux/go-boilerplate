package app

import (
	"bytes"
	"crypto/cipher"
	"crypto/rand"
	"reflect"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/xerrors"
)

const (
	tagName  = "encrypt"
	tagValue = "true"
)

var (
	magicBytes    = []byte{0xBE, 0xEF}
	magicBytesLen = len(magicBytes)
)

type (
	Version byte
	Key     []byte
	KeyMap  map[Version]Key
	aeadMap map[Version]cipher.AEAD
)

type EncryptionManager struct {
	aeadMap aeadMap
	version Version
}

func NewEncryptionManager(version Version, keyMap KeyMap) (*EncryptionManager, error) {
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
	return &EncryptionManager{
		aeadMap: aeadMap,
		version: version,
	}, nil
}

// Encrypt encrypts all fields in ef marked with the encrypt tag.
// ef must be a pointer to a struct with at least one tagged field.
func (m *EncryptionManager) Encrypt(i interface{}) error {
	t, v, err := getReflectedTypeAndValue(i)
	if err != nil {
		return xerrors.Errorf("failed to parse encrypt input: %w", err)
	}

	fields, err := getTaggedFieldNames(v, t)
	if err != nil {
		return xerrors.Errorf("failed to get field names for encryption: %w", err)
	}

	for _, field := range fields {
		f := v.FieldByName(field)
		val := []byte(f.String())
		if isEncrypted(val) {
			return xerrors.New("cannot re-encrypt encrypted value")
		}

		nonce, err := generateNonce()
		if err != nil {
			return xerrors.Errorf("nonce generation error: %w", err)
		}
		cipherText := m.aeadMap[m.version].Seal(nil, nonce, val, nil)
		encVal := setParts(m.version, nonce, cipherText)
		f.SetString(string(encVal))
	}
	return nil
}

// Decrypt decrypts all fields in ef marked with the encrypt tag.
// ef must be a pointer to a struct with at least one tagged field.
func (m *EncryptionManager) Decrypt(i interface{}) error {
	t, v, err := getReflectedTypeAndValue(i)
	if err != nil {
		return xerrors.Errorf("failed to parse decrypt input: %w", err)
	}

	fields, err := getTaggedFieldNames(v, t)
	if err != nil {
		return xerrors.Errorf("failed to get field names for decryption: %w", err)
	}

	for _, field := range fields {
		f := v.FieldByName(field)
		val := []byte(f.String())
		if !isEncrypted(val) {
			return xerrors.New("trying to decrypt unencrypted value")
		}

		keyVersion, nonce, val, err := getParts(val)
		if err != nil {
			return xerrors.Errorf("failed to get parts from value: %w", err)
		}

		aead, ok := m.aeadMap[keyVersion]
		if !ok {
			return xerrors.Errorf("unknown keyVersion %d during decryption", keyVersion)
		}

		plainText, err := aead.Open(nil, nonce, val, nil)
		if err != nil {
			return xerrors.Errorf("failed to decrypt or authenticate value: %w", err)
		}
		f.SetString(string(plainText))
	}
	return nil
}

func getReflectedTypeAndValue(i interface{}) (reflect.Type, *reflect.Value, error) {
	t := reflect.TypeOf(i)
	v := reflect.ValueOf(i)

	if i == nil {
		return nil, nil, xerrors.New("value is nil")
	}
	if v.Kind() != reflect.Ptr {
		return nil, nil, xerrors.New("value is not a pointer")
	}
	if reflect.ValueOf(i).IsNil() {
		return nil, nil, xerrors.New("value is a nil pointer")
	}

	v = reflect.Indirect(v)
	return t, &v, nil
}

// getTaggedFieldNames returns field names from struct v that are
// marked with encrypt=true meta tag.
func getTaggedFieldNames(v *reflect.Value, t reflect.Type) ([]string, error) {
	if v.Kind() != reflect.Struct {
		return nil, xerrors.New("cannot encrypt non-struct value")
	}

	var fields []string
	for i := 0; i < v.NumField(); i++ {
		field := t.Elem().Field(i)
		tag := field.Tag.Get(tagName)

		if tag == tagValue {
			if field.Type.Kind() != reflect.String {
				return nil, xerrors.New("encrypt fields must be of type string")
			}
			fields = append(fields, field.Name)
		}
	}

	if len(fields) == 0 {
		return nil, xerrors.New("struct has no encryptable fields marked with encrypt tag")
	}
	return fields, nil
}

func isEncrypted(val []byte) bool {
	return bytes.Equal(val[0:magicBytesLen], magicBytes)
}

func getParts(b []byte) (Version, []byte, []byte, error) {
	startNonce := magicBytesLen + 1
	endNonce := startNonce + chacha20poly1305.NonceSizeX

	if len(b) < magicBytesLen+1+chacha20poly1305.NonceSizeX {
		return 0, nil, nil, xerrors.New("invalid byte array")
	}
	if err := validateNonce(b[startNonce:endNonce]); err != nil {
		return 0, nil, nil, xerrors.Errorf("invalid nonce: %w", err)
	}
	return Version(b[magicBytesLen]), b[startNonce:endNonce], b[endNonce:], nil
}

func setParts(keyVersion Version, nonce, val []byte) []byte {
	startNonce := magicBytesLen + 1
	endNonce := startNonce + chacha20poly1305.NonceSizeX

	out := make([]byte, magicBytesLen+1+chacha20poly1305.NonceSizeX+len(val))

	copy(out[0:magicBytesLen], magicBytes) // magicBytes
	out[magicBytesLen] = byte(keyVersion)  // version
	copy(out[startNonce:endNonce], nonce)  // nonce
	copy(out[endNonce:], val)              // value

	return out
}

func generateNonce() ([]byte, error) {
	nonce := make([]byte, chacha20poly1305.NonceSizeX)
	if _, err := rand.Read(nonce); err != nil {
		return nil, xerrors.Errorf("error generating nonce: %w")
	}
	return nonce, nil
}

func validateNonce(nonce []byte) error {
	if nonce == nil || len(nonce) != chacha20poly1305.NonceSizeX {
		return xerrors.New("invalid nonce")
	}
	return nil
}
