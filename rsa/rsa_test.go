package rsa

import (
	"testing"
	"github.com/iwind/TeaGo/Tea"
)

func TestRSA_Encrypt(t *testing.T) {
	r, err := NewRSA(Tea.Root+"/certs/app.crt", Tea.Root+"/certs/app.key")
	if err != nil {
		t.Fatal(err)
	}

	result, err := r.Encrypt([]byte("daa55f6c66ad07a25366786cb66a537a@1532860807@218.6.161.211"))
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(result))

	decryptedResult, err := r.Decrypt([]byte(result))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(decryptedResult))
}

func TestRSA_Decrypt(t *testing.T) {
	r, err := NewRSA(Tea.Root+"/certs/app.crt", Tea.Root+"/certs/app.key")
	if err != nil {
		t.Fatal(err)
	}
	data := "VS61GfuZz0Et5Y4nyigIwMZk4IydiY3FvpwRuJPUndRvWlDmQCRIFI6si2MjRaGFmsJxuPsBlHA1GHrlRBWONoczyro5VuPNRyRr/A+Om6tY5HxFe30L/kD0IfloeEX+kDtoEnsj8vTOKSwP+CvyMdb8wF4kBiFDzqWXLSgruhA="
	result, err := r.Decrypt([]byte(data))
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(result))
}
