package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestTransformer(t *testing.T) {
	key := "momsbestpicture"

	pathkey := ContentAddressiblePathTransformer(key)

	fmt.Println(pathkey.Pathname)

	expectedOriginal := "6804429f74181a63c50c3d81d733a12f14a353ff"

	expectedPathname := "68044/29f74/181a6/3c50c/3d81d/733a1/2f14a/353ff"

	if pathkey.Pathname != expectedPathname {
		t.Errorf("have %s, want %s", pathkey.Pathname, expectedPathname)
	}

	if pathkey.Filename != expectedOriginal {
		t.Errorf("have wrong original %s, want origninal %s", pathkey.Pathname, expectedOriginal)
	}
}

func TestStoreDeleteKey(t *testing.T) {
	options := StoreOptions{
		PathTransformer: ContentAddressiblePathTransformer,
	}

	s := NewStore(options)

	key := "momsspecials"

	data := []byte("some jpeg bytes")

	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	if err := s.Delete(key); err != nil {
		t.Error(err)
	}
}

func TestStore(t *testing.T) {
	options := StoreOptions{
		PathTransformer: ContentAddressiblePathTransformer,
	}

	s := NewStore(options)

	key := "momsspecials"

	data := []byte("some jpeg bytes")

	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	if ok := s.Has(key); !ok {
		t.Errorf("Expected to have key %s", key)
	}

	read, err := s.Read(key)

	if err != nil {
		t.Error(err)
	}

	// b, readAllErr := ioutil.ReadAll(read)
	b, readAllErr := io.ReadAll(read)

	if readAllErr != nil {
		t.Error(readAllErr)
	}

	fmt.Println(string(b))

	if string(b) != string(data) {
		t.Errorf("want %s, have %s", data, b)
	}

	s.Delete(key)

}
