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

type EncryptionManager struct {
	aead cipher.AEAD
}

func NewEncryptionManager(aead cipher.AEAD) *EncryptionManager {
	return &EncryptionManager{aead: aead}
}

// Encrypt encrypts all fields in ef marked with the encrypt tag.
// ef must be a pointer to a struct with at least one tagged field.
func (m *EncryptionManager) Encrypt(ef EncryptableFields) error {
	t, v, err := m.getTypeAndValue(ef)
	if err != nil {
		return xerrors.Errorf("failed to parse encrypt input: %w", err)
	}

	if ef.IsEncrypted() {
		return xerrors.New("cannot re-encrypt encrypted struct")
	}

	fields, err := m.getTaggedFieldNames(v, t)
	if err != nil {
		return xerrors.Errorf("failed to get field names for encryption: %w", err)
	}

	nonce, err := generateNonce()
	if err != nil {
		return err
	}

	for _, field := range fields {
		f := v.FieldByName(field)

		cipherText := m.encrypt(f.String(), nonce)
		f.SetString(cipherText)
	}

	ef.SetEncrypted(true)
	ef.SetNonce(nonce)
	return nil
}

// Decrypt decrypts all fields in ef marked with the encrypt tag.
// ef must be a pointer to a struct with at least one tagged field.
func (m *EncryptionManager) Decrypt(ef EncryptableFields) error {
	t, v, err := m.getTypeAndValue(ef)
	if err != nil {
		return xerrors.Errorf("failed to parse decrypt input: %w", err)
	}

	if !ef.IsEncrypted() {
		return xerrors.New("trying to decrypt unencrypted struct")
	}

	fields, err := m.getTaggedFieldNames(v, t)
	if err != nil {
		return xerrors.Errorf("failed to get field names for decryption: %w", err)
	}

	for _, field := range fields {
		f := v.FieldByName(field)

		plainText, err := m.decrypt(f.String(), ef.GetNonce())
		if err != nil {
			return err
		}
		f.SetString(plainText)
	}

	ef.SetEncrypted(false)
	return nil
}

func (m *EncryptionManager) getTypeAndValue(ef EncryptableFields) (reflect.Type, *reflect.Value, error) {
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
func (m *EncryptionManager) getTaggedFieldNames(v *reflect.Value, t reflect.Type) ([]string, error) {
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

func (m *EncryptionManager) encrypt(val, nonce string) string {
	cipherText := m.aead.Seal(nil, []byte(nonce), []byte(val), nil)
	return string(cipherText)
}

func (m *EncryptionManager) decrypt(val, nonce string) (string, error) {
	if nonce == "" || len([]byte(nonce)) != chacha20poly1305.NonceSizeX {
		return "", xerrors.New("invalid nonce")
	}

	plainText, err := m.aead.Open(nil, []byte(nonce), []byte(val), nil)
	if err != nil {
		return "", xerrors.Errorf("failed to decrypt or authenticate value: %w", err)
	}
	return string(plainText), nil
}

func generateNonce() (string, error) {
	nonce := make([]byte, chacha20poly1305.NonceSizeX)
	if _, err := rand.Read(nonce); err != nil {
		return "", xerrors.Errorf("error generating nonce: %w")
	}
	return string(nonce), nil
}
