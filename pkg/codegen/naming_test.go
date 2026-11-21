package codegen

import (
	"go/format"
	"math/rand"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const namingOptionsSpec = `
openapi: 3.0.3
info: {title: Naming options, version: "1"}
paths:
  /things:
    get:
      operationId: listThings
      parameters:
        - name: page_size
          in: query
          x-go-name: Page_Size
          schema: {type: integer}
      responses:
        '200': {description: ok}
components:
  schemas:
    Level:
      type: string
      enum: [LOG_LEVEL_DEBUG, log_level_info, LOG-LEVEL-WARN, _hidden, http_api]
    Named:
      type: string
      enum: [a, b, c]
      x-enum-varnames: [FIRST_VALUE, second_value]
    Thing:
      type: object
      properties:
        _id: {type: string}
        _UPPER_SNAKE_CASE: {type: string}
        __meta: {type: string}
        _: {type: string}
        user_id: {type: string}
        renamed:
          type: string
          x-go-name: HARD_CODED_ID
`

// generateNamingOptions generates the models and the client of spec with the
// given options, and checks that the result is Go.
func generateNamingOptions(t *testing.T, spec string, configure func(*Configuration)) (string, error) {
	t.Helper()
	resetGlobalState(t)

	swagger, err := openapi3.NewLoader().LoadFromData([]byte(spec))
	require.NoError(t, err)

	opts := Configuration{
		PackageName:   "api",
		Generate:      GenerateOptions{Models: true, Client: true},
		OutputOptions: OutputOptions{SkipPrune: true},
	}
	if configure != nil {
		configure(&opts)
	}
	if err := opts.Validate(); err != nil {
		return "", err
	}

	code, err := Generate(swagger, opts)
	if err != nil {
		return "", err
	}
	_, err = format.Source([]byte(code))
	require.NoError(t, err)
	// gofmt pads declarations into columns; one space is easier to match.
	return strings.Join(strings.Fields(code), " "), nil
}

var namingNormalizers = []NameNormalizerFunction{
	NameNormalizerFunctionToCamelCase,
	NameNormalizerFunctionToCamelCaseWithDigits,
	NameNormalizerFunctionToCamelCaseWithInitialisms,
}

// With no naming option set, every name is what it was before the options
// existed.
func TestNamingOptionsDefault(t *testing.T) {
	code, err := generateNamingOptions(t, namingOptionsSpec, nil)
	require.NoError(t, err)

	for _, want := range []string{
		"LOGLEVELDEBUG Level =",
		"LogLevelInfo ",
		"LOGLEVELWARN ",
		"UnderscoreHidden ",
		"HttpApi ",
		"FIRSTVALUE ",
		"SecondValue ",
		"UnderscoreId ",
		"UnderscoreUPPERSNAKECASE ",
		"UnderscoreUnderscoreMeta ",
		"HARDCODEDID ",
		"PageSize ",
	} {
		assert.Contains(t, code, want)
	}
}

func TestNamingEnumValuesKeepUnderscores(t *testing.T) {
	// The names every normalizer agrees on, and the one they do not.
	common := []string{
		"LOG_LEVEL_DEBUG Level =",
		"Log_Level_Info ",
		// A hyphen is still a separator the normalizer removes.
		"LOGLEVELWARN ",
		// A leading underscore keeps its word, so the constant stays exported.
		"Underscore_Hidden ",
		// Names the spec states go through the same conversion.
		"FIRST_VALUE ",
		"Second_Value ",
		// Nothing but enum constants changes.
		"UnderscoreUPPERSNAKECASE ",
		"HARDCODEDID ",
		"PageSize ",
	}
	perNormalizer := map[NameNormalizerFunction]string{
		NameNormalizerFunctionToCamelCase:                "Http_Api ",
		NameNormalizerFunctionToCamelCaseWithDigits:      "Http_Api ",
		NameNormalizerFunctionToCamelCaseWithInitialisms: "HTTP_API ",
	}

	for _, normalizer := range namingNormalizers {
		t.Run(string(normalizer), func(t *testing.T) {
			code, err := generateNamingOptions(t, namingOptionsSpec, func(c *Configuration) {
				c.OutputOptions.NameNormalizer = string(normalizer)
				c.OutputOptions.Naming.EnumValues.KeepUnderscores = true
			})
			require.NoError(t, err)

			for _, want := range append(common, perNormalizer[normalizer]) {
				assert.Contains(t, code, want)
			}
		})
	}
}

// Keeping underscores can make two values share a name that they did not share
// before, or stop them sharing one. Either way no value is lost.
func TestNamingEnumValuesKeepUnderscoresCollisions(t *testing.T) {
	const spec = `
openapi: 3.0.3
info: {title: t, version: "1"}
paths: {}
components:
  schemas:
    Level:
      type: string
      enum: [a_b, A_B, ab, a-b]
`
	code, err := generateNamingOptions(t, spec, func(c *Configuration) {
		c.OutputOptions.Naming.EnumValues.KeepUnderscores = true
	})
	require.NoError(t, err)

	assert.Contains(t, code, `A_B Level = "a_b"`)
	assert.Contains(t, code, `A_B1 Level = "A_B"`)
	assert.Contains(t, code, `Ab Level = "ab"`)
	assert.Contains(t, code, `AB Level = "a-b"`)
}

func TestNamingEnumValuesKeepUnderscoresRejectsOldEnumConflicts(t *testing.T) {
	_, err := generateNamingOptions(t, namingOptionsSpec, func(c *Configuration) {
		c.OutputOptions.Naming.EnumValues.KeepUnderscores = true
		c.Compatibility.OldEnumConflicts = true
	})
	require.ErrorContains(t, err, "keep-underscores")
	require.ErrorContains(t, err, "old-enum-conflicts")
}

// An enum name is converted twice, with colliding names told apart in
// between. The second conversion must leave alone what the first one made of
// an ordinary name, underscores included, or keeping them would not survive
// it.
func TestNamingEnumValuesKeepUnderscoresSecondPassIsStable(t *testing.T) {
	const (
		first = "abcXYZ"
		rest  = "abcdipXYZDIP0189___"
	)
	words := []string{"id", "ID", "api", "http", "Url", "nil", "type"}

	for _, normalizer := range namingNormalizers {
		t.Run(string(normalizer), func(t *testing.T) {
			resetGlobalState(t)
			nameNormalizer = NameNormalizers[normalizer]
			globalState.options.OutputOptions.Naming.EnumValues.KeepUnderscores = true

			r := rand.New(rand.NewSource(2682))
			for range 50000 {
				name := string(first[r.Intn(len(first))])
				for range r.Intn(10) {
					if r.Intn(4) == 0 {
						name += words[r.Intn(len(words))]
					} else {
						name += string(rest[r.Intn(len(rest))])
					}
				}

				once := SanitizeGoIdentity(enumValueNameToGoName(name))
				twice := enumValueNameToGoName(once)
				require.Equal(t, once, twice, "enum name %q", name)
			}
		})
	}
}

func TestNamingPropertiesLeadingUnderscoreDrop(t *testing.T) {
	perNormalizer := map[NameNormalizerFunction][]string{
		NameNormalizerFunctionToCamelCase:                {"Id ", "UserId "},
		NameNormalizerFunctionToCamelCaseWithDigits:      {"Id ", "UserId "},
		NameNormalizerFunctionToCamelCaseWithInitialisms: {"ID ", "UserID "},
	}

	for _, normalizer := range namingNormalizers {
		t.Run(string(normalizer), func(t *testing.T) {
			code, err := generateNamingOptions(t, namingOptionsSpec, func(c *Configuration) {
				c.OutputOptions.NameNormalizer = string(normalizer)
				c.OutputOptions.Naming.Properties.LeadingUnderscore = LeadingUnderscoreDrop
			})
			require.NoError(t, err)

			for _, want := range append([]string{
				"UPPERSNAKECASE ",
				"Meta ",
				// A property that is only underscores has nothing else to be
				// named after.
				"Underscore ",
				// Nothing but struct fields changes.
				"UnderscoreHidden ",
			}, perNormalizer[normalizer]...) {
				assert.Contains(t, code, want)
			}
			assert.NotContains(t, code, "UnderscoreId")
			assert.NotContains(t, code, "UnderscoreUPPERSNAKECASE")
			assert.NotContains(t, code, "UnderscoreMeta")
		})
	}
}

// Dropping the underscore brings back the collision that spelling it out
// avoided, which is reported like any other pair of colliding properties.
func TestNamingPropertiesLeadingUnderscoreDropCollision(t *testing.T) {
	const spec = `
openapi: 3.0.3
info: {title: t, version: "1"}
paths: {}
components:
  schemas:
    Thing:
      type: object
      properties:
        _id: {type: string}
        id: {type: string}
`
	_, err := generateNamingOptions(t, spec, func(c *Configuration) {
		c.OutputOptions.Naming.Properties.LeadingUnderscore = LeadingUnderscoreDrop
	})
	require.ErrorContains(t, err, "property '_id'")
	require.ErrorContains(t, err, "property 'id'")
}

func TestNamingExplicitNamesVerbatim(t *testing.T) {
	for _, normalizer := range namingNormalizers {
		t.Run(string(normalizer), func(t *testing.T) {
			code, err := generateNamingOptions(t, namingOptionsSpec, func(c *Configuration) {
				c.OutputOptions.NameNormalizer = string(normalizer)
				c.OutputOptions.Naming.ExplicitNames = ExplicitNamesVerbatim
			})
			require.NoError(t, err)

			for _, want := range []string{
				`FIRST_VALUE Named = "a"`,
				// Unexported, as written.
				`second_value Named = "b"`,
				// The value x-enum-varnames does not reach is named as usual.
				`C Named = "c"`,
				"HARD_CODED_ID ",
				"Page_Size ",
				// Names the spec does not state are still derived.
				"LOGLEVELDEBUG ",
				"UnderscoreUPPERSNAKECASE ",
			} {
				assert.Contains(t, code, want)
			}
		})
	}
}

// A name emitted as written does not get the type name an enum's other
// constants are prefixed with.
func TestNamingExplicitNamesVerbatimSkipsEnumPrefix(t *testing.T) {
	code, err := generateNamingOptions(t, namingOptionsSpec, func(c *Configuration) {
		c.OutputOptions.Naming.ExplicitNames = ExplicitNamesVerbatim
		c.Compatibility.AlwaysPrefixEnumValues = true
	})
	require.NoError(t, err)

	assert.Contains(t, code, `FIRST_VALUE Named = "a"`)
	assert.Contains(t, code, `second_value Named = "b"`)
	assert.Contains(t, code, `NamedC Named = "c"`)
	assert.Contains(t, code, "LevelLOGLEVELDEBUG ")
}

// A derived name gives way to a name emitted as written, wherever in the enum
// the two are.
func TestNamingExplicitNamesVerbatimWinsCollisions(t *testing.T) {
	const spec = `
openapi: 3.0.3
info: {title: t, version: "1"}
paths: {}
components:
  schemas:
    Before:
      type: string
      enum: [x, y, Kept]
      x-enum-varnames: [Kept, Other]
    After:
      type: integer
      enum: [1, 2]
      x-enumNames: [N1]
`
	code, err := generateNamingOptions(t, spec, func(c *Configuration) {
		c.OutputOptions.Naming.ExplicitNames = ExplicitNamesVerbatim
		c.OutputOptions.Naming.EnumValues.KeepUnderscores = true
	})
	require.NoError(t, err)

	assert.Contains(t, code, `Kept Before = "x"`)
	assert.Contains(t, code, `Kept1 Before = "Kept"`)
	assert.Contains(t, code, "N1 After = 1")
	assert.Contains(t, code, "N2 After = 2")
}

func TestNamingExplicitNamesVerbatimRejects(t *testing.T) {
	const template = `
openapi: 3.0.3
info: {title: t, version: "1"}
paths: {}
components:
  schemas:
    Level:
      type: string
      enum: [a, b]
      x-enum-varnames: `

	for name, tc := range map[string]struct {
		varnames string
		want     string
	}{
		"not an identifier": {`["A B", C]`, "must be a valid Go identifier"},
		"a keyword":         {`[func, C]`, "must be a valid Go identifier"},
		"code":              {`["A = \"a\"\n)\nfunc init() { panic(1) }\nconst (\nB", C]`, "must be a valid Go identifier"},
		"a duplicate":       {`[Same, Same]`, `names two values "Same"`},
	} {
		t.Run(name, func(t *testing.T) {
			verbatim := func(c *Configuration) {
				c.OutputOptions.Naming.ExplicitNames = ExplicitNamesVerbatim
			}
			_, err := generateNamingOptions(t, template+tc.varnames, verbatim)
			require.ErrorContains(t, err, tc.want)

			// Without the option the same names are sanitized instead.
			_, err = generateNamingOptions(t, template+tc.varnames, nil)
			require.NoError(t, err)
		})
	}
}

func TestNamingOptionsValidate(t *testing.T) {
	_, err := generateNamingOptions(t, namingOptionsSpec, func(c *Configuration) {
		c.OutputOptions.Naming.ExplicitNames = "literal"
	})
	require.ErrorContains(t, err, "naming.explicit-names")

	_, err = generateNamingOptions(t, namingOptionsSpec, func(c *Configuration) {
		c.OutputOptions.Naming.Properties.LeadingUnderscore = "keep"
	})
	require.ErrorContains(t, err, "naming.properties.leading-underscore")
}

func TestNamingOptionsTogether(t *testing.T) {
	code, err := generateNamingOptions(t, namingOptionsSpec, func(c *Configuration) {
		c.OutputOptions.Naming = NamingOptions{
			EnumValues:    EnumValueNamingOptions{KeepUnderscores: true},
			Properties:    PropertyNamingOptions{LeadingUnderscore: LeadingUnderscoreDrop},
			ExplicitNames: ExplicitNamesVerbatim,
		}
	})
	require.NoError(t, err)

	for _, want := range []string{
		"LOG_LEVEL_DEBUG ",
		"Log_Level_Info ",
		"FIRST_VALUE ",
		"second_value ",
		"UPPERSNAKECASE ",
		"HARD_CODED_ID ",
		"Page_Size ",
	} {
		assert.Contains(t, code, want)
	}
}
