package encrypt

import (
	"testing"

	"golang.org/x/xerrors"
)

type (
	Tagged struct {
		Unaffected string
		Value1     string `encrypt:"true"`
		Value2     string `encrypt:"true"`
	}
	NoTags struct {
		Value string
	}
	WrongType struct {
		Value int `encrypt:"true"`
	}
)

func TestTagEncrypter_Encrypt(t *testing.T) {
	keyMap := KeyMap{1: []byte("itWouldBeBadIfSomebodyFoundThis!")}
	encrypter, _ := NewTagEncrypter(KeyVersion(1), keyMap)

	t.Run("encrypts struct with tagged fields", func(t *testing.T) {
		tagged := &Tagged{"unaffected", "value1", "value2"}
		err := encrypter.Encrypt(tagged)

		if err != nil {
			t.Errorf("expected no error: %v", err)
		}
		if !encrypted([]byte(tagged.Value1)) || !encrypted([]byte(tagged.Value2)) {
			t.Errorf("tagged values must be encrypted: val1 %s, val2 %s", tagged.Value1, tagged.Value2)
		}
		if encrypted([]byte(tagged.Unaffected)) {
			t.Errorf("untagged values must not be changed: val %s", tagged.Unaffected)
		}
	})
	t.Run("structs must contain at least one tagged value", func(t *testing.T) {
		tagged := &NoTags{"value"}
		err := encrypter.Encrypt(tagged)

		if !xerrors.Is(err, ErrNoTaggedFields) {
			t.Errorf("expected ErrNoTaggedFields but got %v", err)
		}
	})
	t.Run("only struct pointers must be accepted as input", func(t *testing.T) {
		if err := encrypter.Encrypt(1); err == nil {
			t.Errorf("non-struct value was accepted")
		}
		i := 1
		if err := encrypter.Encrypt(&i); err == nil {
			t.Errorf("non-struct pointer value was accepted")
		}
		var s *string
		if err := encrypter.Encrypt(s); err == nil {
			t.Errorf("nil pointer value was accepted")
		}
		if err := encrypter.Encrypt(nil); err == nil {
			t.Errorf("nil must be rejected without panic")
		}
	})
	t.Run("can only encrypt string values", func(t *testing.T) {
		tagged := &WrongType{1}
		if err := encrypter.Encrypt(tagged); err == nil {
			t.Errorf("int value accepted")
		}
	})
}

func TestTagEncrypter_Decrypt(t *testing.T) {
	keyMap := KeyMap{1: []byte("itWouldBeBadIfSomebodyFoundThis!")}
	encrypter, _ := NewTagEncrypter(KeyVersion(1), keyMap)

	t.Run("encrypts struct with tagged fields", func(t *testing.T) {
		tagged := &Tagged{"unaffected", "value1", "value2"}
		_ = encrypter.Encrypt(tagged)

		err := encrypter.Decrypt(tagged)

		if err != nil {
			t.Errorf("expected no error: %v", err)
		}
		if encrypted([]byte(tagged.Value1)) || encrypted([]byte(tagged.Value2)) {
			t.Errorf("tagged values must be unencrypted: val1 %s, val2 %s", tagged.Value1, tagged.Value2)
		}
		if encrypted([]byte(tagged.Unaffected)) {
			t.Errorf("untagged values must not be changed: val %s", tagged.Unaffected)
		}
	})
	t.Run("structs must contain at least one tagged value", func(t *testing.T) {
		tagged := &NoTags{"value"}
		err := encrypter.Decrypt(tagged)

		if !xerrors.Is(err, ErrNoTaggedFields) {
			t.Errorf("expected ErrNoTaggedFields but got %v", err)
		}
	})
}
