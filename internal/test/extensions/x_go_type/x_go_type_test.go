package extensionsxgotype

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// issue #1957: x-go-type and x-go-type-skip-optional-pointer should be usable together.
// When an optional field is annotated with both, the generated field must be a non-pointer
// value type (not *googleuuid.UUID), even though the field is optional.
func TestGeneratedCode(t *testing.T) {
	t.Run("For an object", func(t *testing.T) {
		t.Run("A required field should be a non-pointer", func(t *testing.T) {
			theType := TypeWithOptionalField{
				AtRequired: uuid.New(),
			}

			require.NotZero(t, theType.AtRequired)
		})

		t.Run("An optional field with x-go-type-skip-optional-pointer should be a non-pointer", func(t *testing.T) {
			theType := TypeWithOptionalField{
				AtRequired: uuid.New(),
			}

			require.NotZero(t, theType.AtRequired)
		})
	})

	t.Run("For a query parameter", func(t *testing.T) {
		t.Run("An optional field with x-go-type-skip-optional-pointer should be a non-pointer", func(t *testing.T) {

			u := uuid.New()

			theType := GetRootParams{
				At: u,
			}

			require.NotZero(t, theType.At)
		})
	})

	t.Run("For a field with an AllOf", func(t *testing.T) {
		t.Run("An optional field with x-go-type-skip-optional-pointer should be a non-pointer", func(t *testing.T) {

			u := uuid.New()

			theType := TypeWithAllOf{
				Id: u,
			}

			require.NotZero(t, theType.Id)
		})
	})
}
