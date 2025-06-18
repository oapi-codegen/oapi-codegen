package issue1914

import (
	"io"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testServer = "https://localhost"

func TestNewPostPetRequestWithTextBody_CorrectlyMarshalsUUID(t *testing.T) {
	id := uuid.New()

	req, err := NewPostPetRequestWithTextBody(testServer, id)
	require.NoError(t, err)

	defer req.Body.Close() //nolint:errcheck

	bytes, err := io.ReadAll(req.Body)
	require.NoError(t, err)

	require.NotEmpty(t, bytes)

	assert.Equal(t, id.String(), string(bytes))
}

func TestNewPostPet1234RequestWithTextBody_CorrectlyMarshalsFloat(t *testing.T) {
	var id float32 = 1234.1

	req, err := NewPostPet1234RequestWithTextBody(testServer, id)
	require.NoError(t, err)

	defer req.Body.Close() //nolint:errcheck

	bytes, err := io.ReadAll(req.Body)
	require.NoError(t, err)

	require.NotEmpty(t, bytes)

	assert.Equal(t, "1234.1", string(bytes))
}
