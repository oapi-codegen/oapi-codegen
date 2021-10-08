package ecdsafile

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

// These are utilities for working with files containing ECDSA public and
// private keys. See this helpful doc for how to generate them:
// https://wiki.openssl.org/index.php/Command_Line_Elliptic_Curve_Operations
//
// The quick cheat sheet below.
// 1) Generate an ECDSA-P256 private key
//    openssl ecparam -name prime256v1 -genkey -noout -out ecprivatekey.pem
// 2) Generate public key from private key
//    openssl ec -in ecprivatekey.pem -pubout -out ecpubkey.pem

// LoadEcdsaPublicKey reads an ECDSA public key from an X509 encoding stored in a PEM encoding.
func LoadEcdsaPublicKey(buf []byte) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode(buf)

	if block == nil {
		return nil, errors.New("no PEM data block found")
	}
	// The public key is loaded via a generic loader. We use X509 key format,
	// which supports multiple types of keys.
	keyIface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("Error loading public key: %w", err)
	}

	// Now, we're assuming the key content is ECDSA, and converting.
	publicKey, ok := keyIface.(*ecdsa.PublicKey)
	if !ok {
		// The cast failed, we might have loaded an RSA file or something.
		return nil, errors.New("file contents were not an ECDSA public key")
	}
	return publicKey, nil
}

// LoadEcdsaPrivateKey reads an ECDSA private key from an X509 encoding stored in a PEM encoding
func LoadEcdsaPrivateKey(buf []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(buf)

	if block == nil {
		return nil, errors.New("no PEM data block found")
	}

	// At this point, we've got a valid PEM data block. PEM is just an encoding,
	// and we're assuming this encoding contains X509 key material.
	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("Error loading private ECDSA key: %w", err)
	}
	return privateKey, nil
}

// StoreEcdsaPublicKey writes an ECDSA public key to a PEM encoding
func StoreEcdsaPublicKey(publicKey *ecdsa.PublicKey) ([]byte, error) {
	encodedKey, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("error x509 encoding public key: %w", err)
	}
	pemEncodedKey := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: encodedKey,
	})
	return pemEncodedKey, nil
}

// StoreEcdsaPrivateKey writes an ECDSA private key to a PEM encoding
func StoreEcdsaPrivateKey(privateKey *ecdsa.PrivateKey) ([]byte, error) {
	encodedKey, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("error x509 encoding private key: %w", err)
	}
	pemEncodedKey := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: encodedKey,
	})
	return pemEncodedKey, nil
}
