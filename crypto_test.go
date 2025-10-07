package main

import (
	"bytes"
	"fmt"
	"testing"
)

// func TestNewEncryptionKey(t *testing.T) {
// 	key := newEncryptionKey()

// 	for i := 0; i < len(key); i++ {
// 		if key[i] == 0x0 {
// 			t.Error("0 bytes")
// 		}
// 	}
// }

func TestCopyEncryptDecrypt(t *testing.T) {
	payload := "Foo not bar"

	src := bytes.NewReader([]byte(payload))

	dst := new(bytes.Buffer)

	key := newEncryptionKey()

	_, err := copyEncrypt(key, src, dst)

	if err != nil {
		t.Error(err)
	}

	fmt.Println(dst.String())

	out := new(bytes.Buffer)

	nw, err := copyDecrypt(key, dst, out)

	if err != nil {
		t.Error(err)
	}

	if nw != (16 + len(payload)) {
		t.Fail()
	}

	if out.String() != payload {
		t.Error("decryption failed!!")
	}

	fmt.Println(out.String())
}
