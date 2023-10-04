package issue1302_test

import (
	"encoding/json"
	"testing"

	issue1302 "github.com/deepmap/oapi-codegen/v2/internal/test/issues/issue-1302"
	"github.com/stretchr/testify/assert"
)

func TestIssue1302(t *testing.T) {
	buf, err := json.Marshal(issue1302.Test{})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"Object":{},"BigInteger":{}}`, string(buf))
}
