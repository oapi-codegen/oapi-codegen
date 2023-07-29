package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmail_MarshalJSON_Validation(t *testing.T) {
	type requiredEmail struct {
		EmailField Email `json:"email"`
	}

	testCases := map[string]struct {
		email         Email
		expectedJSON  []byte
		expectedError error
	}{
		"it should succeed marshalling a valid email and return valid JSON populated with the email": {
			email:         Email("validemail@openapicodegen.com"),
			expectedJSON:  []byte(`{"email":"validemail@openapicodegen.com"}`),
			expectedError: nil,
		},
		"it should fail marshalling an invalid email and return a validation error": {
			email:         Email("invalidemail"),
			expectedJSON:  nil,
			expectedError: ErrValidationEmail,
		},
		"it should fail marshalling an empty email and return a validation error": {
			email:         Email(""),
			expectedJSON:  nil,
			expectedError: ErrValidationEmail,
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			jsonBytes, err := json.Marshal(requiredEmail{EmailField: tc.email})

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.JSONEq(t, string(tc.expectedJSON), string(jsonBytes))
			}
		})
	}
}

func TestEmail_UnmarshalJSON_RequiredEmail_Validation(t *testing.T) {
	type requiredEmail struct {
		EmailField Email `json:"email"`
	}

	requiredEmailTestCases := map[string]struct {
		jsonStr       string
		expectedEmail Email
		expectedError error
	}{
		"it should succeed validating a valid email during the unmarshal process": {
			jsonStr:       `{"email":"gaben@valvesoftware.com"}`,
			expectedError: nil,
			expectedEmail: func() Email {
				e := Email("gaben@valvesoftware.com")
				return e
			}(),
		},
		"it should fail validating an invalid email": {
			jsonStr:       `{"email":"not-an-email"}`,
			expectedError: ErrValidationEmail,
			expectedEmail: func() Email {
				e := Email("not-an-email")
				return e
			}(),
		},
		"it should fail validating an empty email": {
			jsonStr: `{"email":""}`,
			expectedEmail: func() Email {
				e := Email("")
				return e
			}(),
			expectedError: ErrValidationEmail,
		},
		"it should fail validating a null email": {
			jsonStr: `{"email":null}`,
			expectedEmail: func() Email {
				e := Email("")
				return e
			}(),
			expectedError: ErrValidationEmail,
		},
	}

	for name, tc := range requiredEmailTestCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			b := requiredEmail{}
			err := json.Unmarshal([]byte(tc.jsonStr), &b)
			assert.Equal(t, tc.expectedEmail, b.EmailField)
			assert.ErrorIs(t, err, tc.expectedError)
		})
	}

}

func TestEmail_UnmarshalJSON_NullableEmail_Validation(t *testing.T) {

	type nullableEmail struct {
		EmailField *Email `json:"email,omitempty"`
	}

	nullableEmailTestCases := map[string]struct {
		body          nullableEmail
		jsonStr       string
		expectedEmail *Email
		expectedError error
	}{
		"it should succeed validating a valid email during the unmarshal process": {
			body:          nullableEmail{},
			jsonStr:       `{"email":"gaben@valvesoftware.com"}`,
			expectedError: nil,
			expectedEmail: func() *Email {
				e := Email("gaben@valvesoftware.com")
				return &e
			}(),
		},
		"it should fail validating an invalid email": {
			body:          nullableEmail{},
			jsonStr:       `{"email":"not-an-email"}`,
			expectedError: ErrValidationEmail,
			expectedEmail: func() *Email {
				e := Email("not-an-email")
				return &e
			}(),
		},
		"it should fail validating an empty email": {
			body:          nullableEmail{},
			jsonStr:       `{"email":""}`,
			expectedError: ErrValidationEmail,
			expectedEmail: func() *Email {
				e := Email("")
				return &e
			}(),
		},
		"it should succeed validating a null email": {
			body:          nullableEmail{},
			jsonStr:       `{"email":null}`,
			expectedEmail: nil,
			expectedError: nil,
		},
		"it should succeed validating a missing email": {
			body:          nullableEmail{},
			jsonStr:       `{}`,
			expectedEmail: nil,
			expectedError: nil,
		},
	}

	for name, tc := range nullableEmailTestCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := json.Unmarshal([]byte(tc.jsonStr), &tc.body)
			assert.Equal(t, tc.expectedEmail, tc.body.EmailField)
			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
			}
		})
	}
}
