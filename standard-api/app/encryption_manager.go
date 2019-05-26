package app

import (
	"crypto/cipher"
	"crypto/rand"
	"reflect"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/xerrors"
)

var (
	tagName = "encrypt"
)

type EncryptableFields interface {
	GetNonce() string
	SetNonce(nonce string)
	IsEncrypted() bool
	SetEncrypted(b bool)
}

type Encryptable struct {
	Nonce     string `json:"-"`
	Encrypted bool   `json:"-"`
}

func (e *Encryptable) GetNonce() string {
	return e.Nonce
}
func (e *Encryptable) SetNonce(nonce string) {
	e.Nonce = nonce
}
func (e *Encryptable) IsEncrypted() bool {
	return e.Encrypted
}
func (e *Encryptable) SetEncrypted(b bool) {
	e.Encrypted = b
}

type EncryptionManager struct {
	aead cipher.AEAD
}

func NewEncryptionManager(aead cipher.AEAD) *EncryptionManager {
	return &EncryptionManager{aead: aead}
}

// FIXME 2019-05-26 @Jimeux 適当なリフレクション
func (m *EncryptionManager) Encrypt(ef EncryptableFields) error {
	if ef.IsEncrypted() {
		return xerrors.New("trying to re-encrypt encrypted struct")
	}

	t := reflect.TypeOf(ef)
	v := reflect.ValueOf(ef)
	if v.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
	}

	nonce := string(generateNonce())

	for i := 0; i < v.NumField(); i++ {
		field := t.Elem().Field(i)
		tag := field.Tag.Get(tagName)

		if tag == "true" {
			ciph := m.encrypt(v.Field(i).String(), nonce)
			v.Field(i).SetString(ciph)
		}
	}

	ef.SetEncrypted(true)
	ef.SetNonce(nonce)
	return nil
}

// FIXME 2019-05-26 @Jimeux 適当なリフレクション
func (m *EncryptionManager) Decrypt(ef EncryptableFields) error {
	if !ef.IsEncrypted() {
		return xerrors.New("trying to decrypt unencrypted struct")
	}

	t := reflect.TypeOf(ef)
	v := reflect.ValueOf(ef)
	if v.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
	}

	for i := 0; i < v.NumField(); i++ {
		field := t.Elem().Field(i)
		tag := field.Tag.Get(tagName)

		if tag == "true" {
			plain, err := m.decrypt(v.Field(i).String(), ef.GetNonce())
			if err != nil {
				return err
			}
			v.Field(i).SetString(plain)
		}
	}

	ef.SetEncrypted(false)
	return nil
}

func (m *EncryptionManager) encrypt(val, nonce string) string {
	ciphertext := m.aead.Seal(nil, []byte(nonce), []byte(val), nil)
	return string(ciphertext)
}

func (m *EncryptionManager) decrypt(val, nonce string) (string, error) {
	if err := validateNonce(nonce); err != nil {
		return "", err
	}
	plaintext, err := m.aead.Open(nil, []byte(nonce), []byte(val), nil)
	if err != nil {
		return "", xerrors.Errorf("failed to decrypt or authenticate message: %w", err)
	}
	return string(plaintext), nil
}

func validateNonce(nonce string) error {
	if nonce == "" || len([]byte(nonce)) != chacha20poly1305.NonceSizeX {
		return xerrors.New("invalid nonce")
	}
	return nil
}

func generateNonce() []byte {
	nonce := make([]byte, chacha20poly1305.NonceSizeX)
	_, _ = rand.Read(nonce) // TODO 2019-05-26 @Jimeux error handling
	return nonce
}
