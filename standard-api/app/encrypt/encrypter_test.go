package encrypt

import (
	"bytes"
	"testing"
)

func TestEncrypter_NewEncrypter(t *testing.T) {
	t.Run("keyMap must contain current version", func(t *testing.T) {
		encrypter, err := NewEncrypter(KeyVersion(1), KeyMap{2: []byte("itWouldBeBadIfSomeoneFoundThis!!")})
		if err == nil {
			t.Errorf("keyMap must contain current key version: %v", err)
		}
		if encrypter != nil {
			t.Errorf("return value must be nil on error")
		}
	})
	t.Run("key length is validated", func(t *testing.T) {
		encrypter, err := NewEncrypter(KeyVersion(1), KeyMap{1: []byte("too-short")})
		if err == nil {
			t.Errorf("incorrect key lengths must be rejected")
		}
		if encrypter != nil {
			t.Errorf("return value must be nil on error")
		}
	})
}
func TestEncrypter_Encrypt(t *testing.T) {
	currentVersion := KeyVersion(1)
	keyMap := KeyMap{
		1: []byte("itWouldBeBadIfSomebodyFoundThis!"),
		2: []byte("itWouldBeBadIfSomeoneFoundThis!!"),
	}
	encrypter, _ := NewEncrypter(currentVersion, keyMap)
	val := []byte("hello")

	t.Run("encrypts byte slice", func(t *testing.T) {
		ev, err := encrypter.Encrypt(val)

		if err != nil {
			t.Errorf("expected no error: %v", err)
		}
		if bytes.Equal(ev, val) || bytes.Equal(ev.Value(), val) {
			t.Errorf("expected value to be transformed: %s", ev)
		}
		if !encrypted(ev) {
			t.Errorf("expected value to be encrypted: %s", ev)
		}
		if ev.KeyVersion() != byte(currentVersion) {
			t.Errorf("expected version %d but got %d", currentVersion, ev.KeyVersion())
		}
	})
	t.Run("does not re-encrypt encrypted values", func(t *testing.T) {
		enc, _ := encrypter.Encrypt(val)
		ev, err := encrypter.Encrypt(enc)
		if err == nil {
			t.Fatal("expected re-encryption error but got nil")
		}
		if ev != nil {
			t.Errorf("expected value to be nil")
		}
	})
}

func TestEncrypter_Decrypt(t *testing.T) {
	currentVersion := KeyVersion(1)
	keyMap := KeyMap{
		1: []byte("itWouldBeBadIfSomebodyFoundThis!"),
		2: []byte("itWouldBeBadIfSomeoneFoundThis!!"),
	}
	encrypter, _ := NewEncrypter(currentVersion, keyMap)
	val := []byte("hello")

	t.Run("decrypts byte slices", func(t *testing.T) {
		ev, _ := encrypter.Encrypt(val)
		b, err := encrypter.Decrypt(ev)
		if err != nil {
			t.Errorf("expected no error: %v", err)
		}
		if !bytes.Equal(b, val) {
			t.Errorf("decrypted value %s does not match original %s", b, val)
		}
		if encrypted(b) {
			t.Errorf("encrypted was true for %s", b)
		}
	})
	t.Run("does not decrypt unencrypted byte slices", func(t *testing.T) {
		b, err := encrypter.Decrypt(val)
		if err == nil {
			t.Errorf("expected does not contain padding error but got nil")
		}
		if b != nil {
			t.Errorf("return value must be nil on error")
		}
	})
	t.Run("fails to authenticate with incorrect nonce", func(t *testing.T) {
		ev, _ := encrypter.Encrypt(val)
		[]byte(ev)[nonceStart] = 0xAD
		b, err := encrypter.Decrypt(ev)
		if err == nil {
			t.Errorf("expected authentication error")
		}
		if b != nil {
			t.Errorf("return value must be nil on error")
		}
	})
	t.Run("cannot decrypt with unknown key version", func(t *testing.T) {
		ev, _ := encrypter.Encrypt(val)
		[]byte(ev)[keyVersionIndex] = 0xAD
		b, err := encrypter.Decrypt(ev)
		if err == nil {
			t.Errorf("expected unknown key error")
		}
		if b != nil {
			t.Errorf("return value must be nil on error")
		}
	})
	t.Run("supports key transitioning", func(t *testing.T) {
		t.SkipNow() // TODO 2019-06-07 @Jimeux
	})
}
