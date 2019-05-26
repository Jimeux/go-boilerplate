package app

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
