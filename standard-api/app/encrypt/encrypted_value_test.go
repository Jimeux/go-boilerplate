package encrypt

import (
	"bytes"
	"fmt"
	"testing"
)

var (
	testKeyVersion = byte(1)
	testNonce      = []byte("-testing-nonce-24-bytes-")
	testValue      = []byte("value")
)

func TestNewEncryptedValue(t *testing.T) {
	str := fmt.Sprintf("%s%s%s%s", magicBytes, []byte{testKeyVersion}, testNonce, testValue)
	ev := NewEncryptedValue(testKeyVersion, testNonce, testValue)

	if ev.KeyVersion() != testKeyVersion {
		t.Errorf("expected KeyVersion %d but got %d", int(testKeyVersion), int(ev.KeyVersion()))
	}
	if !bytes.Equal(ev.Nonce(), testNonce) {
		t.Errorf("expected Nonce %s but got %s", string(testNonce), string(ev.Nonce()))
	}
	if !bytes.Equal(ev.Value(), testValue) {
		t.Errorf("expected Value %s but got %s", string(testValue), ev.Value())
	}
	if ev.String() != str {
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

	ev = NewEncryptedValue(testKeyVersion, testNonce, testValue)
	ev, err = EncryptedValueFromByteSlice([]byte(ev))
	if err != nil {
		t.Errorf("expected no error but got: %v", err)
	}
}
