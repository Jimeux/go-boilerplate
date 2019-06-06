package encrypt

import (
	"bytes"
	"fmt"
	"testing"
)

func TestNewEncryptedValue(t *testing.T) {
	keyVersion := byte(1)
	nonce := []byte("-testing-nonce-24-bytes-")
	value := []byte("value")
	str := fmt.Sprintf("%s%s%s%s", magicBytes, []byte{keyVersion}, nonce, value)

	ev := NewEncryptedValue(keyVersion, nonce, value)

	if ev.KeyVersion() != keyVersion {
		t.Errorf("expected KeyVersion %d but got %d", int(keyVersion), int(ev.KeyVersion()))
	}
	if !bytes.Equal(ev.Nonce(), nonce) {
		t.Errorf("expected Nonce %s but got %s", string(nonce), string(ev.Nonce()))
	}
	if !bytes.Equal(ev.Value(), value) {
		t.Errorf("expected Value %s but got %s", string(value), ev.Value())
	}
	if !bytes.Equal([]byte(ev.String()), []byte(str)) {
		t.Errorf("expected String %s but got %s", str, ev.String())
	}
	if !encrypted(ev) {
		t.Errorf("value was not encrypted: %s", ev.String())
	}
}

func TestEncryptedValueFromByteSlice(t *testing.T) {
	ev, err := EncryptedValueFromByteSlice(make([]byte, 1))
	if err == nil || ev != nil {
		t.Errorf("expected length error but got nil")
	}

	ev, err = EncryptedValueFromByteSlice(make([]byte, paddingLen))
	if err == nil {
		t.Errorf("expected unencrypted error but got nil")
	}

	keyVersion := byte(1)
	nonce := []byte("-testing-nonce-24-bytes-")
	value := []byte("value")
	ev = NewEncryptedValue(keyVersion, nonce, value)

	ev, err = EncryptedValueFromByteSlice([]byte(ev))
	if err != nil {
		t.Errorf("expected no error but got: %v", err)
	}
}
