package ecdsafile

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoreEcdsaPrivateKey_RoundTrip(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	encoded, err := StoreEcdsaPrivateKey(key)
	require.NoError(t, err)

	block, _ := pem.Decode(encoded)
	require.NotNil(t, block)

	// The key is marshalled as SEC1, which is labelled "EC PRIVATE KEY".
	// "PRIVATE KEY" is the PKCS#8 label.
	assert.Equal(t, "EC PRIVATE KEY", block.Type)

	parsed, err := x509.ParseECPrivateKey(block.Bytes)
	require.NoError(t, err)
	assert.True(t, key.Equal(parsed))

	loaded, err := LoadEcdsaPrivateKey(encoded)
	require.NoError(t, err)
	assert.True(t, key.Equal(loaded))
}

func TestStoreEcdsaPublicKey_RoundTrip(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	encoded, err := StoreEcdsaPublicKey(&key.PublicKey)
	require.NoError(t, err)

	block, _ := pem.Decode(encoded)
	require.NotNil(t, block)
	assert.Equal(t, "PUBLIC KEY", block.Type)

	loaded, err := LoadEcdsaPublicKey(encoded)
	require.NoError(t, err)
	assert.True(t, key.PublicKey.Equal(loaded))
}
