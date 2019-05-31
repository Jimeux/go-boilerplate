package app

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

// Encrypt encrypts all fields in ef marked with the encrypt tag.
// ef must be a pointer to a struct with at least one tagged field.
func (m *EncryptionManager) Encrypt(ef EncryptableFields) error {
	t, v, err := getTypeAndValue(ef)
	if err != nil {
		return xerrors.Errorf("failed to parse encrypt input: %w", err)
	}

	if ef.IsEncrypted() {
		return xerrors.New("cannot re-encrypt encrypted struct")
	}

	fields, err := getTaggedFieldNames(v, t)
	if err != nil {
		return xerrors.Errorf("failed to get field names for encryption: %w", err)
	}

	for _, field := range fields {
		f := v.FieldByName(field)

		nonce, err := generateNonce()
		if err != nil {
			return xerrors.Errorf("nonce generation error: %w", err)
		}
		cipherText := m.aeadMap[m.version].Seal(nil, nonce, []byte(f.String()), nil)
		val := setParts(m.version, nonce, cipherText)
		f.SetString(string(val))
	}

	ef.SetEncrypted(true)
	return nil
}

// Decrypt decrypts all fields in ef marked with the encrypt tag.
// ef must be a pointer to a struct with at least one tagged field.
func (m *EncryptionManager) Decrypt(ef EncryptableFields) error {
	t, v, err := getTypeAndValue(ef)
	if err != nil {
		return xerrors.Errorf("failed to parse decrypt input: %w", err)
	}
	if !ef.IsEncrypted() {
		return xerrors.New("trying to decrypt unencrypted struct")
	}

	fields, err := getTaggedFieldNames(v, t)
	if err != nil {
		return xerrors.Errorf("failed to get field names for decryption: %w", err)
	}

	for _, field := range fields {
		f := v.FieldByName(field)

		keyVersion, nonce, val, err := getParts([]byte(f.String()))
		if err != nil {
			return xerrors.Errorf("failed to get keyVersion: %w", err)
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

	ef.SetEncrypted(false)
	return nil
}

func getTypeAndValue(ef EncryptableFields) (reflect.Type, *reflect.Value, error) {
	t := reflect.TypeOf(ef)
	v := reflect.ValueOf(ef)

	if ef == nil {
		return nil, nil, xerrors.New("value is nil")
	}
	if v.Kind() != reflect.Ptr {
		return nil, nil, xerrors.New("value is not a pointer")
	}
	if reflect.ValueOf(ef).IsNil() {
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

func getParts(b []byte) (Version, []byte, []byte, error) {
	if len(b) < 25 {
		return 0, nil, nil, xerrors.New("invalid byte array")
	}
	if err := validateNonce(b[1:25]); err != nil {
		return 0, nil, nil, xerrors.Errorf("invalid nonce during decryption: %w", err)
	}
	return Version(b[0]), b[1:25], b[25:], nil
}

func setParts(keyVersion Version, nonce, val []byte) []byte {
	return append(append([]byte{byte(keyVersion)}, nonce...), val...)
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
