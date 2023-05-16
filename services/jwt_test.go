package services

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"
)

var (
	privKeyPath = "../private_key.pem"
	pubKeyPath  = "../public_key.pem"
)

func TestNewJWT(t *testing.T) {
	/* private-key */
	signBytes, err := os.ReadFile(privKeyPath)
	if err != nil {
		t.Fatal(err)
	}
	var block *pem.Block
	if block, _ = pem.Decode(signBytes); block == nil {
		t.Fatal(ErrKeyMustBePEMEncoded)
	}

	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKCS1PrivateKey(block.Bytes); err != nil {
		t.Fatal(err)
	}

	var ok bool
	if _, ok = parsedKey.(*rsa.PrivateKey); !ok {
		t.Fatal(ErrNotRSAPrivateKey)
	}

	/* public-key */
	signBytes, err = os.ReadFile(pubKeyPath)
	if err != nil {
		t.Fatal(err)
	}

	if block, _ = pem.Decode(signBytes); block == nil {
		t.Fatal(ErrKeyMustBePEMEncoded)
	}
	if parsedKey, err = x509.ParsePKIXPublicKey(block.Bytes); err != nil {
		t.Fatal(err)
	}
	if _, ok = parsedKey.(*rsa.PublicKey); !ok {
		t.Fatal(ErrNotRSAPublicKey)
	}

}
