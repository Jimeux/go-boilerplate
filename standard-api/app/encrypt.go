package app

type EncryptableFields interface {
	IsEncrypted() bool
	SetEncrypted(b bool)
}

type Encryptable struct {
	Encrypted bool `json:"-"`
}

func (e *Encryptable) IsEncrypted() bool {
	return e.Encrypted
}
func (e *Encryptable) SetEncrypted(b bool) {
	e.Encrypted = b
}
