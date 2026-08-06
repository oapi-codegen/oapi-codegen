package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// allGenerators turns on every generator so EmittedComponentNames covers the
// whole surface. Configuration.Validate would reject it (only one server type
// at a time), but the component-name helpers do not care.
var allGenerators = GenerateOptions{
	IrisServer:    true,
	ChiServer:     true,
	FiberServer:   true,
	FiberV3Server: true,
	EchoServer:    true,
	Echo5Server:   true,
	GinServer:     true,
	GorillaServer: true,
	StdHTTPServer: true,
	Strict:        true,
	Client:        true,
	Models:        true,
	EmbeddedSpec:  true,
}

func resolveFor(t *testing.T, oo OutputOptions, g GenerateOptions) ComponentNames {
	t.Helper()
	cn, err := resolveComponentNames(Configuration{PackageName: "api", Generate: g, OutputOptions: oo})
	require.NoError(t, err)
	return cn
}

// TestComponentNamesDefaults pins the default resolution: with nothing
// configured, every name is exactly what oapi-codegen has always emitted.
// This is the guard for the "component-names is a pure no-op when unused"
// property that the committed fixtures also assert byte-for-byte.
func TestComponentNamesDefaults(t *testing.T) {
	cn := resolveFor(t, OutputOptions{}, allGenerators)

	assert.Equal(t, "Client", cn.Client)
	assert.Equal(t, "Client", cn.ClientStem)
	assert.Equal(t, "ClientInterface", cn.ClientInterface)
	assert.Equal(t, "ClientOption", cn.ClientOption)
	assert.Equal(t, "NewClient", cn.NewClient)
	assert.Equal(t, "ClientWithResponses", cn.ClientWithResponses)
	assert.Equal(t, "ClientWithResponsesInterface", cn.ClientWithResponsesInterface)
	assert.Equal(t, "NewClientWithResponses", cn.NewClientWithResponses)
	assert.Equal(t, "RequestEditorFn", cn.RequestEditorFn)
	assert.Equal(t, "HttpRequestDoer", cn.HTTPRequestDoer)
	assert.Equal(t, "WithHTTPClient", cn.WithHTTPClient)
	assert.Equal(t, "WithRequestEditorFn", cn.WithRequestEditorFn)
	assert.Equal(t, "WithBaseURL", cn.WithBaseURL)

	assert.Equal(t, "ServerInterface", cn.ServerInterface)
	assert.Equal(t, "ServerInterfaceWrapper", cn.ServerInterfaceWrapper)
	assert.Equal(t, "MiddlewareFunc", cn.MiddlewareFunc)

	assert.Equal(t, "Handler", cn.Handler)
	assert.Equal(t, "HandlerFromMux", cn.HandlerFromMux)
	assert.Equal(t, "HandlerFromMuxWithBaseURL", cn.HandlerFromMuxWithBaseURL)
	assert.Equal(t, "HandlerWithOptions", cn.HandlerWithOptions)
	assert.Equal(t, "Unimplemented", cn.Unimplemented)

	assert.Equal(t, "RegisterHandlers", cn.RegisterHandlers)
	assert.Equal(t, "RegisterHandlersWithBaseURL", cn.RegisterHandlersWithBaseURL)
	assert.Equal(t, "RegisterHandlersWithOptions", cn.RegisterHandlersWithOptions)
	assert.Equal(t, "RegisterHandlersOptions", cn.RegisterHandlersOptions)

	assert.Equal(t, "StrictServerInterface", cn.StrictServerInterface)
	assert.Equal(t, "StrictHandlerFunc", cn.StrictHandlerFunc)
	assert.Equal(t, "StrictMiddlewareFunc", cn.StrictMiddlewareFunc)
	assert.Equal(t, "NewStrictHandler", cn.NewStrictHandler)
	assert.Equal(t, "NewStrictHandlerWithOptions", cn.NewStrictHandlerWithOptions)
	assert.Equal(t, "strictHandler", cn.StrictHandler)
	assert.Equal(t, "StrictHTTPServerOptions", cn.StrictHTTPServerOptions)
	assert.Equal(t, "StrictGinServerOptions", cn.StrictGinServerOptions)

	assert.Equal(t, "GetSwagger", cn.GetSwagger)
	assert.Equal(t, "GetSpec", cn.GetSpec)
	assert.Equal(t, "GetSpecJSON", cn.GetSpecJSON)
	assert.Equal(t, "PathToRawSpec", cn.PathToRawSpec)
	assert.Equal(t, "swaggerSpec", cn.SwaggerSpec)
	assert.Equal(t, "rawSpec", cn.RawSpec)
	assert.Equal(t, "decodeSpec", cn.DecodeSpec)
	assert.Equal(t, "decodeSpecCached", cn.DecodeSpecCached)

	assert.Equal(t, "ChiServerOptions", cn.ChiServerOptions)
	assert.Equal(t, "GorillaServerOptions", cn.GorillaServerOptions)
	assert.Equal(t, "StdHTTPServerOptions", cn.StdHTTPServerOptions)
	assert.Equal(t, "GinServerOptions", cn.GinServerOptions)
	assert.Equal(t, "FiberServerOptions", cn.FiberServerOptions)
	assert.Equal(t, "IrisServerOptions", cn.IrisServerOptions)

	assert.Equal(t, "RequiredParamError", cn.Errors.RequiredParamError)
	assert.Equal(t, "RequiredHeaderError", cn.Errors.RequiredHeaderError)
	assert.Equal(t, "InvalidParamFormatError", cn.Errors.InvalidParamFormatError)
	assert.Equal(t, "TooManyValuesForParamError", cn.Errors.TooManyValuesForParamError)
	assert.Equal(t, "UnmarshalingParamError", cn.Errors.UnmarshalingParamError)
	assert.Equal(t, "UnescapedCookieParamError", cn.Errors.UnescapedCookieParamError)
	assert.Equal(t, "EchoRouter", cn.Echo.Router)
	assert.Equal(t, "ServeMux", cn.StdHTTP.ServeMux)
	assert.Equal(t, "HandlerMiddlewareFunc", cn.Fiber.HandlerMiddlewareFunc)
	assert.Empty(t, cn.Prefix)
}

// TestComponentNamesPrefix covers the prefix layer, including the
// unexported-name casing rule that keeps swaggerSpec & friends unexported.
func TestComponentNamesPrefix(t *testing.T) {
	cn := resolveFor(t, OutputOptions{
		ComponentNames: ComponentNames{Prefix: "PetStore"},
	}, allGenerators)

	// Exported names take the prefix verbatim.
	assert.Equal(t, "PetStoreServerInterface", cn.ServerInterface)
	assert.Equal(t, "PetStoreServerInterfaceWrapper", cn.ServerInterfaceWrapper)
	assert.Equal(t, "PetStoreClient", cn.Client)
	assert.Equal(t, "PetStoreClientInterface", cn.ClientInterface)
	assert.Equal(t, "NewPetStoreClient", cn.NewClient)
	assert.Equal(t, "NewPetStoreClientWithResponses", cn.NewClientWithResponses)
	assert.Equal(t, "PetStoreGetSwagger", cn.GetSwagger)
	assert.Equal(t, "PetStoreEchoRouter", cn.Echo.Router)
	assert.Equal(t, "PetStoreServeMux", cn.StdHTTP.ServeMux)
	assert.Equal(t, "PetStoreRequiredParamError", cn.Errors.RequiredParamError)

	// Unexported names lower the prefix's first letter so they stay unexported.
	assert.Equal(t, "petStoreSwaggerSpec", cn.SwaggerSpec)
	assert.Equal(t, "petStoreRawSpec", cn.RawSpec)
	assert.Equal(t, "petStoreDecodeSpec", cn.DecodeSpec)
	assert.Equal(t, "petStoreDecodeSpecCached", cn.DecodeSpecCached)
	assert.Equal(t, "petStoreStrictHandler", cn.StrictHandler)
}

// TestComponentNamesOverridesWithoutPrefix checks that an explicit override
// replaces the default, and that its family derives from it.
func TestComponentNamesOverridesWithoutPrefix(t *testing.T) {
	cn := resolveFor(t, OutputOptions{
		ComponentNames: ComponentNames{
			ServerInterface: "MyServer",
			GetSwagger:      "GetMySpec",
			Errors:          ErrorComponentNames{RequiredParamError: "MissingParam"},
			Echo:            EchoComponentNames{Router: "MyEchoRouter"},
		},
	}, allGenerators)

	assert.Equal(t, "MyServer", cn.ServerInterface)
	assert.Equal(t, "MyServerWrapper", cn.ServerInterfaceWrapper) // derived from the override
	assert.Equal(t, "GetMySpec", cn.GetSwagger)
	assert.Equal(t, "GetSpec", cn.GetSpec) // sibling untouched
	assert.Equal(t, "MissingParam", cn.Errors.RequiredParamError)
	assert.Equal(t, "RequiredHeaderError", cn.Errors.RequiredHeaderError)
	assert.Equal(t, "MyEchoRouter", cn.Echo.Router)
}

// TestComponentNamesPrefixAppliesToOverrides pins the uniform-prefix rule:
// the prefix is prepended to every resolved name, an explicit override
// included. It is independent of the overrides -- it affects everything or
// nothing.
func TestComponentNamesPrefixAppliesToOverrides(t *testing.T) {
	cn := resolveFor(t, OutputOptions{
		ComponentNames: ComponentNames{
			Prefix:          "PetStore",
			Client:          "MyClient",
			ServerInterface: "MyServer",
			GetSwagger:      "GetMySpec",
			Errors:          ErrorComponentNames{RequiredParamError: "MissingParam"},
			Echo:            EchoComponentNames{Router: "MyEchoRouter"},
		},
	}, allGenerators)

	assert.Equal(t, "PetStoreMyClient", cn.Client)
	// ... and the family derives from the prefixed override, prefixed once.
	assert.Equal(t, "PetStoreMyClientInterface", cn.ClientInterface)
	assert.Equal(t, "NewPetStoreMyClient", cn.NewClient)
	assert.Equal(t, "PetStoreMyServer", cn.ServerInterface)
	assert.Equal(t, "PetStoreMyServerWrapper", cn.ServerInterfaceWrapper)
	assert.Equal(t, "PetStoreGetMySpec", cn.GetSwagger)
	assert.Equal(t, "PetStoreGetSpec", cn.GetSpec) // untouched sibling, prefixed too
	assert.Equal(t, "PetStoreMissingParam", cn.Errors.RequiredParamError)
	assert.Equal(t, "PetStoreMyEchoRouter", cn.Echo.Router)
}

// TestComponentNamesDerivation covers each derivation family: renaming a root
// renames its whole family.
func TestComponentNamesDerivation(t *testing.T) {
	cn := resolveFor(t, OutputOptions{
		ComponentNames: ComponentNames{
			Client:           "PetAPI",
			ServerInterface:  "PetServer",
			Handler:          "PetHandler",
			RegisterHandlers: "MountPetRoutes",
		},
	}, allGenerators)

	assert.Equal(t, "PetAPI", cn.Client)
	assert.Equal(t, "PetAPIInterface", cn.ClientInterface)
	assert.Equal(t, "PetAPIOption", cn.ClientOption)
	assert.Equal(t, "NewPetAPI", cn.NewClient)
	assert.Equal(t, "PetAPIWithResponses", cn.ClientWithResponses)
	assert.Equal(t, "PetAPIWithResponsesInterface", cn.ClientWithResponsesInterface)
	assert.Equal(t, "NewPetAPIWithResponses", cn.NewClientWithResponses)

	assert.Equal(t, "PetServerWrapper", cn.ServerInterfaceWrapper)

	assert.Equal(t, "PetHandlerFromMux", cn.HandlerFromMux)
	assert.Equal(t, "PetHandlerFromMuxWithBaseURL", cn.HandlerFromMuxWithBaseURL)
	assert.Equal(t, "PetHandlerWithOptions", cn.HandlerWithOptions)

	assert.Equal(t, "MountPetRoutesWithBaseURL", cn.RegisterHandlersWithBaseURL)
	assert.Equal(t, "MountPetRoutesWithOptions", cn.RegisterHandlersWithOptions)
	assert.Equal(t, "MountPetRoutesOptions", cn.RegisterHandlersOptions)

	// Names outside a derivation family stay at their defaults.
	assert.Equal(t, "RequestEditorFn", cn.RequestEditorFn)
	assert.Equal(t, "StrictHandlerFunc", cn.StrictHandlerFunc)
}

// TestComponentNamesLegacyClientTypeName pins the backwards-compatible
// behavior of the older `client-type-name` knob: it renames the client struct
// and nothing else, so existing generated code is unchanged.
func TestComponentNamesLegacyClientTypeName(t *testing.T) {
	cn := resolveFor(t, OutputOptions{ClientTypeName: "APIClient"}, allGenerators)

	assert.Equal(t, "APIClient", cn.Client)
	assert.Equal(t, "ClientInterface", cn.ClientInterface)
	assert.Equal(t, "ClientOption", cn.ClientOption)
	assert.Equal(t, "NewClient", cn.NewClient)
	assert.Equal(t, "ClientWithResponses", cn.ClientWithResponses)
	assert.Equal(t, "NewClientWithResponses", cn.NewClientWithResponses)

	// component-names.client, by contrast, renames the whole family.
	cn = resolveFor(t, OutputOptions{
		ComponentNames: ComponentNames{Client: "APIClient"},
	}, allGenerators)
	assert.Equal(t, "APIClient", cn.Client)
	assert.Equal(t, "APIClientInterface", cn.ClientInterface)
	assert.Equal(t, "NewAPIClient", cn.NewClient)

	// The prefix applies to the legacy knob too.
	cn = resolveFor(t, OutputOptions{
		ClientTypeName: "APIClient",
		ComponentNames: ComponentNames{Prefix: "PetStore"},
	}, allGenerators)
	assert.Equal(t, "PetStoreAPIClient", cn.Client)
	assert.Equal(t, "PetStoreClientInterface", cn.ClientInterface)
}

// TestComponentNamesLegacyClientTypeNameShadowed checks that setting both
// knobs is not an error: component-names.client simply wins, and Warnings
// reports the shadowing so the user notices.
func TestComponentNamesLegacyClientTypeNameShadowed(t *testing.T) {
	cfg := Configuration{
		PackageName: "api",
		Generate:    GenerateOptions{Client: true},
		OutputOptions: OutputOptions{
			ClientTypeName: "APIClient",
			ComponentNames: ComponentNames{Client: "OtherClient"},
		},
	}
	require.NoError(t, cfg.Validate())

	cn, err := resolveComponentNames(cfg)
	require.NoError(t, err)
	assert.Equal(t, "OtherClient", cn.Client)
	assert.Equal(t, "OtherClientInterface", cn.ClientInterface)

	warnings := cfg.Warnings()
	require.Contains(t, warnings, "client-type-name")
	assert.Contains(t, warnings["client-type-name"], "deprecated")
	assert.Contains(t, warnings["client-type-name"], "OtherClient")

	// Set alone, it is not warned about -- it still works as it always has.
	cfg.OutputOptions.ComponentNames.Client = ""
	assert.NotContains(t, cfg.Warnings(), "client-type-name")
}

func TestComponentNamesIdentifierValidation(t *testing.T) {
	for _, tc := range []struct {
		name    string
		names   ComponentNames
		wantErr string
	}{
		{"space in name", ComponentNames{ServerInterface: "My Server"}, "server-interface"},
		{"go keyword", ComponentNames{Client: "type"}, "client"},
		{"empty-ish prefix", ComponentNames{Prefix: "9Pets"}, "prefix"},
		{"underscore prefix", ComponentNames{Prefix: "_Pets"}, "must start with a letter"},
		{"nested group", ComponentNames{Echo: EchoComponentNames{Router: "my-router"}}, "echo.router"},
		{"error group", ComponentNames{Errors: ErrorComponentNames{UnmarshalingParamError: "1Bad"}}, "errors.unmarshaling-param-error"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := resolveComponentNames(Configuration{
				PackageName:   "api",
				Generate:      allGenerators,
				OutputOptions: OutputOptions{ComponentNames: tc.names},
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestComponentNamesUniqueness(t *testing.T) {
	// Two roots colliding.
	_, err := resolveComponentNames(Configuration{
		PackageName: "api",
		Generate:    GenerateOptions{ChiServer: true, Client: true},
		OutputOptions: OutputOptions{
			ComponentNames: ComponentNames{ServerInterface: "Thing", Handler: "Thing"},
		},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be unique")
	assert.Contains(t, err.Error(), `"Thing"`)

	// A derived name colliding with an unrelated root.
	_, err = resolveComponentNames(Configuration{
		PackageName: "api",
		Generate:    GenerateOptions{ChiServer: true},
		OutputOptions: OutputOptions{
			ComponentNames: ComponentNames{Handler: "Serve", ServerInterface: "ServeFromMux"},
		},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ServeFromMux")

	// The same collision is not an error when neither generator is enabled.
	_, err = resolveComponentNames(Configuration{
		PackageName: "api",
		Generate:    GenerateOptions{Models: true},
		OutputOptions: OutputOptions{
			ComponentNames: ComponentNames{ServerInterface: "Thing", Handler: "Thing"},
		},
	})
	require.NoError(t, err)
}

// TestEmittedComponentNamesGating checks that the emitted set follows the
// `generate` options: a name only takes part in uniqueness and schema
// collision checks when the config actually declares it.
func TestEmittedComponentNamesGating(t *testing.T) {
	cn := resolveFor(t, OutputOptions{}, allGenerators)

	modelsOnly := EmittedComponentNames(cn, GenerateOptions{Models: true})
	assert.Empty(t, modelsOnly)

	clientOnly := EmittedComponentNames(cn, GenerateOptions{Client: true})
	assert.Contains(t, values(clientOnly), "Client")
	assert.NotContains(t, values(clientOnly), "ServerInterface")

	chi := EmittedComponentNames(cn, GenerateOptions{ChiServer: true})
	assert.Contains(t, values(chi), "Unimplemented")
	assert.Contains(t, values(chi), "ChiServerOptions")
	assert.Contains(t, values(chi), "RequiredParamError")
	assert.NotContains(t, values(chi), "ServeMux")
	assert.NotContains(t, values(chi), "RegisterHandlers")

	echo := EmittedComponentNames(cn, GenerateOptions{EchoServer: true})
	assert.Contains(t, values(echo), "EchoRouter")
	assert.Contains(t, values(echo), "RegisterHandlersOptions")
	assert.NotContains(t, values(echo), "RequiredParamError")

	spec := EmittedComponentNames(cn, GenerateOptions{EmbeddedSpec: true})
	assert.Contains(t, values(spec), "swaggerSpec")
	assert.Contains(t, values(spec), "GetSwagger")
}

func values(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}
