package rsa

import (
	"github.com/iwind/TeaGo/files"
	"encoding/pem"
	"crypto/x509"
	"crypto/rsa"
	"encoding/base64"
	"crypto/rand"
	"github.com/iwind/TeaGo/logs"
)

type RSA struct {
	certFile string
	*rsa.PublicKey

	publicFile string

	privateKeyFile string
	*rsa.PrivateKey
}

func NewRSA(certFile string, privateKeyFile string) (*RSA, error) {
	rsaObject := &RSA{
		certFile:       certFile,
		privateKeyFile: privateKeyFile,
	}

	{
		keyFile := files.NewFile(certFile)
		keyBody, err := keyFile.ReadAll()
		if err != nil {
			return nil, err
		}
		block, _ := pem.Decode(keyBody)
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}
		rsaObject.PublicKey = cert.PublicKey.(*rsa.PublicKey)
	}

	{
		keyFile := files.NewFile(privateKeyFile)
		keyBody, err := keyFile.ReadAll()
		if err != nil {
			return nil, err
		}
		block, _ := pem.Decode(keyBody)
		rsaObject.PrivateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
	}

	return rsaObject, nil
}

func NewRSAPair(publicFile string, privateKeyFile string) (*RSA, error) {
	rsaObject := &RSA{
		publicFile:     publicFile,
		privateKeyFile: privateKeyFile,
	}

	{
		keyFile := files.NewFile(publicFile)
		keyBody, err := keyFile.ReadAll()
		if err != nil {
			return nil, err
		}
		block, _ := pem.Decode(keyBody)

		publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			logs.Printf("err1")
			return nil, err
		}
		rsaObject.PublicKey = publicKey.(*rsa.PublicKey)
	}

	{
		keyFile := files.NewFile(privateKeyFile)
		keyBody, err := keyFile.ReadAll()
		if err != nil {
			return nil, err
		}
		block, _ := pem.Decode(keyBody)
		rsaObject.PrivateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
	}

	return rsaObject, nil
}

func (this *RSA) Encrypt(data []byte) ([]byte, error) {
	result, err := rsa.EncryptPKCS1v15(rand.Reader, this.PublicKey, data)
	if err != nil {
		return nil, err
	}
	return []byte(base64.StdEncoding.EncodeToString(result)), err
}

func (this *RSA) Decrypt(data []byte) ([]byte, error) {
	encryptedData, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return nil, err
	}

	dataBytes, err := rsa.DecryptPKCS1v15(rand.Reader, this.PrivateKey, encryptedData)
	if err != nil {
		return nil, err
	}
	return dataBytes, err
}
