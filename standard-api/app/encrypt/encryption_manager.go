package encrypt

import (
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

// Encrypt encrypts all fields in i marked with the encrypt tag.
// i must be a pointer to a struct with at least one tagged field.
func (m *EncryptionManager) Encrypt(i interface{}) error {
	t, v, err := getReflectedTypeAndValue(i)
	if err != nil {
		return xerrors.Errorf("failed to parse encrypt input: %w", err)
	}

	indexes, err := getTaggedFieldIndexes(v, t)
	if err != nil {
		return xerrors.Errorf("failed to get field names for encryption: %w", err)
	}

	for _, index := range indexes {
		f := v.Field(index)

		encVal, err := m.encryptToString(f.String())
		if err != nil {
			return xerrors.Errorf("could not encrypt string: %w", err)
		}

		f.SetString(encVal)
	}
	return nil
}

func (m *EncryptionManager) encryptToString(s string) (string, error) {
	if encrypted([]byte(s)) {
		return "", xerrors.New("value is already encrypted")
	}

	nonce := make([]byte, chacha20poly1305.NonceSizeX)
	if _, err := rand.Read(nonce); err != nil {
		return "", xerrors.Errorf("error generating nonce: %w")
	}

	ciphertext := m.aeadMap[m.version].Seal(nil, nonce, []byte(s), nil)
	return NewEncryptedValue(byte(m.version), nonce, ciphertext).String(), nil
}

// Decrypt decrypts all fields in i marked with the encrypt tag.
// i must be a pointer to a struct with at least one tagged field.
func (m *EncryptionManager) Decrypt(i interface{}) error {
	t, v, err := getReflectedTypeAndValue(i)
	if err != nil {
		return xerrors.Errorf("failed to parse decrypt input: %w", err)
	}

	indexes, err := getTaggedFieldIndexes(v, t)
	if err != nil {
		return xerrors.Errorf("failed to get field names for decryption: %w", err)
	}

	for _, index := range indexes {
		f := v.Field(index)

		plainText, err := m.decryptFromString(f.String())
		if err != nil {
			return xerrors.Errorf("could not decrypt string: %w", err)
		}

		f.SetString(plainText)
	}
	return nil
}

func (m *EncryptionManager) decryptFromString(s string) (string, error) {
	ev, err := FromByteSlice([]byte(s))
	if err != nil {
		return "", xerrors.Errorf("could not decrypt string: %w", err)
	}

	aead, ok := m.aeadMap[Version(ev.KeyVersion())]
	if !ok {
		return "", xerrors.Errorf("unknown keyVersion %d during decryption", ev.KeyVersion())
	}

	plainText, err := aead.Open(nil, ev.Nonce(), ev.Value(), nil)
	if err != nil {
		return "", xerrors.Errorf("failed to decrypt or authenticate value: %w", err)
	}

	return string(plainText), nil
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

	v = reflect.Indirect(v) // return value v points to
	return t, &v, nil
}

// getTaggedFieldIndexes returns field names from struct v that are
// marked with encrypt=true meta tag.
func getTaggedFieldIndexes(v *reflect.Value, t reflect.Type) ([]int, error) {
	if v.Kind() != reflect.Struct {
		return nil, xerrors.New("cannot encrypt non-struct value")
	}

	var fields []int
	for i := 0; i < v.NumField(); i++ {
		field := t.Elem().Field(i)
		tag := field.Tag.Get(tagName)

		if tag == tagValue {
			if field.Type.Kind() != reflect.String {
				return nil, xerrors.New("encrypt fields must be of type string")
			}
			fields = append(fields, i)
		}
	}

	if len(fields) == 0 {
		return nil, xerrors.New("struct has no encryptable fields marked with encrypt tag")
	}
	return fields, nil
}
