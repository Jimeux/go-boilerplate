package encrypt

import (
	"reflect"

	"golang.org/x/xerrors"
)

const (
	tagName  = "encrypt"
	tagValue = "true"
)

var (
	ErrNoTaggedFields = xerrors.New("struct has no fields marked with encrypt tag")
)

// TagEncrypter is a type for automatically encrypting/decrypting
// struct fields marked with the 'encrypt' meta-tag via reflection.
type TagEncrypter struct {
	encrypter *Encrypter
}

func NewTagEncrypter(version KeyVersion, keyMap KeyMap) (*TagEncrypter, error) {
	manager, err := NewEncrypter(version, keyMap)
	if err != nil {
		return nil, xerrors.Errorf("could not create TagEncrypter: %w", err)
	}

	return &TagEncrypter{
		encrypter: manager,
	}, nil
}

// Encrypt encrypts all fields in i marked with the encrypt tag.
// i must be a pointer to a struct with at least one tagged field.
func (m *TagEncrypter) Encrypt(i interface{}) error {
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

		encVal, err := m.encrypter.Encrypt([]byte(f.String()))
		if err != nil {
			return xerrors.Errorf("could not encrypt string: %w", err)
		}

		f.SetString(encVal.String())
	}
	return nil
}

// Decrypt decrypts all fields in i marked with the encrypt tag.
// i must be a pointer to a struct with at least one tagged field.
func (m *TagEncrypter) Decrypt(i interface{}) error {
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

		plainText, err := m.encrypter.Decrypt([]byte(f.String()))
		if err != nil {
			return xerrors.Errorf("could not decrypt string: %w", err)
		}

		f.SetString(string(plainText))
	}
	return nil
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
		return nil, ErrNoTaggedFields
	}
	return fields, nil
}
