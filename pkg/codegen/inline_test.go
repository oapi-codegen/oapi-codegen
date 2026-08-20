// Copyright 2019 DeepMap, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package codegen

import (
	"bytes"
	"compress/lzw"
	"encoding/base64"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// rawSpecFixture and compressedSpecFixture are a known-good pair: rawSpecFixture
// compressed with compressSpec must always produce exactly compressedSpecFixture.
//
// compress/lzw (like compress/flate before it) makes no promise that its
// compressed output is byte-stable across Go versions - only that the format
// remains valid and round-trips correctly. If a future Go release changes
// that output, this test - not a `go generate` diff in some unrelated repo -
// is what should catch it.
const rawSpecFixture = `{"openapi":"3.0.3","info":{"title":"Test API","version":"1.0.0"},"paths":{"/ping":{"get":{"operationId":"getPing","responses":{"200":{"description":"pong","content":{"application/json":{"schema":{"type":"object","properties":{"message":{"type":"string"}}}}}}}}}}}`

const compressedSpecFixture = "APeIeAOnjJswcNKI0CFihgsYLmaIYCEijRszbxYKpJOGDpsyC0VQKTOHDoggUJJMFGGnjJw5ad64CRnj4UMRfSjCCUMHzRyNIl4kdHMG6JkydIASdMkzppskZEIepQPFYlGKckjCkTmHJFAZMGAAJUNyjJw0cDjKDLmV6MoxMukYTKpDIEI4bNKMaSrzhZo5a+uKmDMGTZk2YYDSyVMw5BsxasqMSapTzlI5HL0KbkNyTpijihmDZFjyrNs+qFOr7hMQ"

// TestCompressSpec_KnownGood pins compressSpec's output for a fixed input.
// If this test starts failing after a Go toolchain upgrade, that confirms
// compress/lzw's output has changed underneath us, and any regenerated
// *.gen.go embedding a spec will show a spurious diff.
func TestCompressSpec_KnownGood(t *testing.T) {
	got, err := compressSpec([]byte(rawSpecFixture))
	require.NoError(t, err)
	assert.Equal(t, compressedSpecFixture, got)
}

// TestCompressSpec_RoundTrip confirms the known-good compressed fixture
// decodes back to the known-good raw fixture, using the same lzw.Reader
// settings (order, litWidth) as the decodeSpec function generated into
// *.gen.go by inline.tmpl. If these settings ever drift apart, decoding a
// real embedded spec at runtime would panic/error, and this is where that
// would first show up.
func TestCompressSpec_RoundTrip(t *testing.T) {
	compressed, err := base64.StdEncoding.DecodeString(compressedSpecFixture)
	require.NoError(t, err)

	zr := lzw.NewReader(bytes.NewReader(compressed), lzw.LSB, 8)
	defer func() {
		require.NoError(t, zr.Close())
	}()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, zr)
	require.NoError(t, err)

	assert.Equal(t, rawSpecFixture, buf.String())
}
