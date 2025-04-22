package issue1957

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

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
