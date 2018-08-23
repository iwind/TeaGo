package byteutil

import (
	"testing"
	"encoding/base64"
)

func TestEncrypt(t *testing.T) {
	encrypted, err := Encrypt([]byte("Hello, World"), []byte("hello"))
	if err != nil {
		t.Fatal(err)
	}

	t.Log(len(encrypted), base64.StdEncoding.EncodeToString(encrypted))

	decrypted, err := Decrypt(encrypted, []byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(decrypted), string(decrypted))
}
