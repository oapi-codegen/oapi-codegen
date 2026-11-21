package codegen

import (
	"go/format"
	"math/rand"
	"regexp"
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

// A parameter becomes a field of the operation's parameters struct, and the
// client and server read that field by the parameter's own name for it. The
// two have to agree whatever the naming options are.
func TestNamingParametersMatchTheirFields(t *testing.T) {
	const spec = `
openapi: 3.0.3
info: {title: t, version: "1"}
paths:
  /things/{_thing}:
    get:
      operationId: getThing
      parameters:
        - {name: _thing, in: path, required: true, schema: {type: string}}
        - {name: _id, in: query, schema: {type: string}}
        - {name: _trace, in: header, schema: {type: string}}
        - {name: _session, in: cookie, schema: {type: string}}
        - {name: _named, in: query, x-go-name: _Named_Query, schema: {type: string}}
      responses:
        '200': {description: ok}
`
	for name, tc := range map[string]struct {
		configure func(*Configuration)
		want      []string
	}{
		"default": {
			configure: func(*Configuration) {},
			want: []string{
				"UnderscoreId *string", "params.UnderscoreId",
				"UnderscoreTrace *string", "params.UnderscoreTrace",
				"UnderscoreSession *string", "params.UnderscoreSession",
				"UnderscoreNamedQuery *string", "params.UnderscoreNamedQuery",
				"underscoreThing string",
			},
		},
		"leading underscores dropped": {
			configure: func(c *Configuration) {
				c.OutputOptions.Naming.Properties.LeadingUnderscore = LeadingUnderscoreDrop
			},
			want: []string{
				"Id *string", "params.Id",
				"Trace *string", "params.Trace",
				"Session *string", "params.Session",
				"NamedQuery *string", "params.NamedQuery",
				// A path parameter is an argument, not a field.
				"underscoreThing string",
			},
		},
		"explicit names verbatim": {
			configure: func(c *Configuration) {
				c.OutputOptions.Naming.ExplicitNames = ExplicitNamesVerbatim
			},
			want: []string{
				"UnderscoreId *string", "params.UnderscoreId",
				"_Named_Query *string", "params._Named_Query",
			},
		},
		"both": {
			configure: func(c *Configuration) {
				c.OutputOptions.Naming.Properties.LeadingUnderscore = LeadingUnderscoreDrop
				c.OutputOptions.Naming.ExplicitNames = ExplicitNamesVerbatim
			},
			want: []string{
				"Id *string", "params.Id",
				"_Named_Query *string", "params._Named_Query",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			for _, server := range []func(*GenerateOptions){
				func(g *GenerateOptions) { g.StdHTTPServer = true },
				func(g *GenerateOptions) { g.EchoServer = true },
				func(g *GenerateOptions) { g.ChiServer = true; g.Strict = true },
			} {
				code, err := generateNamingOptions(t, spec, func(c *Configuration) {
					server(&c.Generate)
					tc.configure(c)
				})
				require.NoError(t, err)
				for _, want := range tc.want {
					assert.Contains(t, code, want)
				}
				// Every use of a field of the parameters struct names a
				// field the struct has.
				structBody := code[strings.Index(code, "type GetThingParams struct {"):]
				structBody = structBody[:strings.Index(structBody, "}")]
				for _, use := range regexp.MustCompile(`\bparams\.(\w+)`).FindAllStringSubmatch(code, -1) {
					assert.Contains(t, structBody, " "+use[1]+" ", "params.%s is not a field of GetThingParams", use[1])
				}
			}
		})
	}
}

// The type name put in front of an enum's derived constants can produce a
// name that the spec states for another value, which keeps it.
func TestNamingExplicitNamesVerbatimPrefixCollision(t *testing.T) {
	const spec = `
openapi: 3.0.3
info: {title: t, version: "1"}
paths: {}
components:
  schemas:
    Level:
      type: string
      enum: [a, b, b1]
      x-enum-varnames: [LevelB]
`
	for range 20 {
		code, err := generateNamingOptions(t, spec, func(c *Configuration) {
			c.OutputOptions.Naming.ExplicitNames = ExplicitNamesVerbatim
			c.Compatibility.AlwaysPrefixEnumValues = true
		})
		require.NoError(t, err)

		assert.Contains(t, code, `LevelB Level = "a"`)
		assert.Contains(t, code, `LevelB1 Level = "b1"`)
		assert.Contains(t, code, `LevelB2 Level = "b"`)
		assert.Equal(t, 3, strings.Count(code, "return true"))
	}
}

// The strict server copies each path parameter into a field of the request
// object. A name emitted as written need not be exported, and the field and
// the assignment to it have to spell it the same way.
func TestNamingExplicitNamesVerbatimStrictPathParameter(t *testing.T) {
	const spec = `
openapi: 3.0.3
info: {title: t, version: "1"}
paths:
  /things/{id}:
    get:
      operationId: getThing
      parameters:
        - {name: id, in: path, required: true, x-go-name: thing_id, schema: {type: string}}
      responses:
        '200': {description: ok}
`
	for name, server := range map[string]func(*GenerateOptions){
		"std-http": func(g *GenerateOptions) { g.StdHTTPServer = true },
		"fiber":    func(g *GenerateOptions) { g.FiberServer = true },
		"iris":     func(g *GenerateOptions) { g.IrisServer = true },
	} {
		t.Run(name, func(t *testing.T) {
			code, err := generateNamingOptions(t, spec, func(c *Configuration) {
				server(&c.Generate)
				c.Generate.Strict = true
				c.OutputOptions.Naming.ExplicitNames = ExplicitNamesVerbatim
			})
			require.NoError(t, err)

			assert.Contains(t, code, "type GetThingRequestObject struct { thing_id string")
			assert.Contains(t, code, "request.thing_id = thing_id")
		})
	}
}

// A path parameter is a function argument. Naming it as a struct field would
// let "_ctx" become ctx, an argument the client methods already have.
func TestNamingPropertiesLeadingUnderscoreDropLeavesPathParameters(t *testing.T) {
	const spec = `
openapi: 3.0.3
info: {title: t, version: "1"}
paths:
  /things/{_ctx}:
    get:
      operationId: getThing
      parameters:
        - {name: _ctx, in: path, required: true, schema: {type: string}}
        - {name: _ctx, in: query, schema: {type: string}}
      responses:
        '200': {description: ok}
`
	code, err := generateNamingOptions(t, spec, func(c *Configuration) {
		c.Generate.StdHTTPServer = true
		c.Generate.Strict = true
		c.OutputOptions.Naming.Properties.LeadingUnderscore = LeadingUnderscoreDrop
	})
	require.NoError(t, err)

	assert.Contains(t, code, "GetThing(ctx context.Context, underscoreCtx string, params *GetThingParams")
	assert.Contains(t, code, "request.UnderscoreCtx = underscoreCtx")
	// The query parameter of the same name is a field, and is renamed.
	assert.Contains(t, code, "type GetThingParams struct { Ctx *string")
	assert.Contains(t, code, "params.Ctx")
}

// The field a parameter gets in the parameters struct takes its name from the
// parameter, so the two cannot differ. They used to be derived separately,
// from extensions that were not the same.
func TestParameterFieldIsNamedByTheParameter(t *testing.T) {
	const spec = `
openapi: 3.0.3
info: {title: t, version: "1"}
paths:
  /things:
    get:
      operationId: listThings
      parameters:
        # x-go-name on the schema names its type, not this parameter's field
        - name: kind
          in: query
          schema: {$ref: '#/components/schemas/Kind'}
        - name: honored
          in: query
          x-go-name: lowerName
          x-oapi-codegen-only-honour-go-name: true
          schema: {type: string}
      responses:
        '200': {description: ok}
components:
  schemas:
    Kind:
      type: string
      x-go-name: RenamedKind
`
	code, err := generateNamingOptions(t, spec, func(c *Configuration) {
		c.Compatibility.AllowUnexportedStructFieldNames = true
	})
	require.NoError(t, err)

	assert.Contains(t, code, "type ListThingsParams struct { Kind *RenamedKind")
	assert.Contains(t, code, "params.Kind")
	assert.NotContains(t, code, "params.RenamedKind")
	assert.Contains(t, code, "lowerName *string")
	assert.Contains(t, code, "params.lowerName")
	assert.NotContains(t, code, "params.LowerName")
}
