package codegen

import (
	"errors"
	"fmt"
	"go/token"
	"sort"
	"strings"
	"unicode"
)

// ComponentNames customizes the fixed, spec-independent package-level
// identifiers that oapi-codegen emits (`ServerInterface`, `Client`,
// `GetSwagger`, `RequiredParamError`, ...).
//
// It exists for three reasons:
//
//  1. A spec whose components collide with a generated name (a schema called
//     `Client`, say) can rename the component instead of the schema.
//  2. Two generate runs can share one Go package by giving each a `prefix`;
//     without that every fixed name -- including the unexported ones -- would
//     be declared twice.
//  3. Users who simply want different names (house style, clarity in godoc).
//
// Resolution happens once, in Generate, and fills this struct in place. Three
// layers, in order:
//
//  1. Defaults: the names oapi-codegen has always emitted.
//  2. Prefix: every name still at its default gets Prefix prepended. Exported
//     names are prefixed verbatim (`PetStore` + `ServerInterface` ->
//     `PetStoreServerInterface`); unexported names (`swaggerSpec`, `rawSpec`,
//     `decodeSpec`, `decodeSpecCached`, `strictHandler`) lower the prefix's
//     first letter so they stay unexported (`petStoreSwaggerSpec`).
//  3. Explicit overrides: used verbatim, never prefixed.
//
// Fields carrying a YAML tag are user-configurable. Fields tagged `yaml:"-"`
// are resolution slots only: they are either derived from a configurable root
// (so renaming the root renames its family) or prefix-only. Which is which is
// documented on each field; the complete mapping is in the README.
type ComponentNames struct {
	// Prefix is prepended to every component name still at its default. It
	// must be a valid Go identifier starting with a letter. This is the
	// one-line fix for emitting two specs into a single Go package.
	Prefix string `yaml:"prefix,omitempty"`

	// ---- client (generate.client) ----

	// Client is the generated HTTP client struct. Root of the client family:
	// the derived names below default from it, so renaming it renames them
	// all. Supersedes the older `client-type-name`, which renames only this
	// type and leaves the rest of the family alone.
	Client string `yaml:"client,omitempty"`
	// ClientInterface is derived: <Client>Interface.
	ClientInterface string `yaml:"-"`
	// ClientOption is derived: <Client>Option.
	ClientOption string `yaml:"-"`
	// NewClient is derived: New<Client>.
	NewClient string `yaml:"-"`
	// ClientWithResponses is derived: <Client>WithResponses.
	ClientWithResponses string `yaml:"-"`
	// ClientWithResponsesInterface is derived: <ClientWithResponses>Interface.
	ClientWithResponsesInterface string `yaml:"-"`
	// NewClientWithResponses is derived: New<ClientWithResponses>.
	NewClientWithResponses string `yaml:"-"`
	// RequestEditorFn is prefix-only.
	RequestEditorFn string `yaml:"-"`
	// HTTPRequestDoer is prefix-only. Its default spelling is the historical
	// `HttpRequestDoer`.
	HTTPRequestDoer string `yaml:"-"`
	// WithHTTPClient is prefix-only.
	WithHTTPClient string `yaml:"-"`
	// WithRequestEditorFn is prefix-only.
	WithRequestEditorFn string `yaml:"-"`
	// WithBaseURL is prefix-only.
	WithBaseURL string `yaml:"-"`

	// ---- server, shared across every framework ----

	// ServerInterface is the interface the user implements. Root of its
	// family.
	ServerInterface string `yaml:"server-interface,omitempty"`
	// ServerInterfaceWrapper is derived: <ServerInterface>Wrapper.
	ServerInterfaceWrapper string `yaml:"-"`
	// MiddlewareFunc is the per-framework middleware type alias.
	MiddlewareFunc string `yaml:"middleware-func,omitempty"`

	// ---- net/http family: chi, gorilla, std-http ----

	// Handler is the root of the chi-family handler constructors.
	Handler string `yaml:"handler,omitempty"`
	// HandlerFromMux is derived: <Handler>FromMux.
	HandlerFromMux string `yaml:"-"`
	// HandlerFromMuxWithBaseURL is derived: <Handler>FromMuxWithBaseURL.
	HandlerFromMuxWithBaseURL string `yaml:"-"`
	// HandlerWithOptions is derived: <Handler>WithOptions.
	HandlerWithOptions string `yaml:"-"`
	// Unimplemented is the chi-only stub server implementation.
	Unimplemented string `yaml:"unimplemented,omitempty"`

	// ---- register-handlers family: echo, gin, fiber, iris ----

	// RegisterHandlers is the root of the route-registration family.
	RegisterHandlers string `yaml:"register-handlers,omitempty"`
	// RegisterHandlersWithBaseURL is derived: <RegisterHandlers>WithBaseURL.
	RegisterHandlersWithBaseURL string `yaml:"-"`
	// RegisterHandlersWithOptions is derived: <RegisterHandlers>WithOptions.
	RegisterHandlersWithOptions string `yaml:"-"`
	// RegisterHandlersOptions is derived: <RegisterHandlers>Options (echo).
	RegisterHandlersOptions string `yaml:"-"`

	// ---- strict server (generate.strict-server) ----

	// StrictServerInterface is the strict-mode server interface.
	StrictServerInterface string `yaml:"strict-server-interface,omitempty"`
	// StrictHandlerFunc is prefix-only.
	StrictHandlerFunc string `yaml:"-"`
	// StrictMiddlewareFunc is prefix-only.
	StrictMiddlewareFunc string `yaml:"-"`
	// NewStrictHandler is prefix-only.
	NewStrictHandler string `yaml:"-"`
	// NewStrictHandlerWithOptions is prefix-only.
	NewStrictHandlerWithOptions string `yaml:"-"`
	// StrictHandler is prefix-only, and unexported (`strictHandler`).
	StrictHandler string `yaml:"-"`
	// StrictHTTPServerOptions is prefix-only (chi, gorilla, std-http).
	StrictHTTPServerOptions string `yaml:"-"`
	// StrictGinServerOptions is prefix-only (gin).
	StrictGinServerOptions string `yaml:"-"`

	// ---- embedded spec (generate.embedded-spec) ----

	// GetSwagger is the deprecated spec accessor.
	GetSwagger string `yaml:"get-swagger,omitempty"`
	// GetSpec is the spec accessor.
	GetSpec string `yaml:"get-spec,omitempty"`
	// GetSpecJSON returns the raw embedded JSON.
	GetSpecJSON string `yaml:"get-spec-json,omitempty"`
	// PathToRawSpec is prefix-only. NOTE that this name is also part of the
	// cross-package protocol used by import-mapping: a generated package that
	// $refs another calls `<pkg>.PathToRawSpec`, using the *default* name,
	// because the referencing config cannot know the referenced config's
	// component names. Renaming it in a package that others $ref will break
	// those callers.
	PathToRawSpec string `yaml:"-"`
	// SwaggerSpec is prefix-only, and unexported (`swaggerSpec`).
	SwaggerSpec string `yaml:"-"`
	// RawSpec is prefix-only, and unexported (`rawSpec`).
	RawSpec string `yaml:"-"`
	// DecodeSpec is prefix-only, and unexported (`decodeSpec`).
	DecodeSpec string `yaml:"-"`
	// DecodeSpecCached is prefix-only, and unexported (`decodeSpecCached`).
	DecodeSpecCached string `yaml:"-"`

	// ---- per-framework option structs (prefix-only) ----

	// ChiServerOptions is prefix-only.
	ChiServerOptions string `yaml:"-"`
	// GorillaServerOptions is prefix-only.
	GorillaServerOptions string `yaml:"-"`
	// StdHTTPServerOptions is prefix-only.
	StdHTTPServerOptions string `yaml:"-"`
	// GinServerOptions is prefix-only.
	GinServerOptions string `yaml:"-"`
	// FiberServerOptions is prefix-only.
	FiberServerOptions string `yaml:"-"`
	// IrisServerOptions is prefix-only.
	IrisServerOptions string `yaml:"-"`

	// ---- nested groups (depth cap: exactly one level) ----

	// Errors names the parameter-binding error types emitted by the
	// net/http-family servers (chi, gorilla, std-http).
	Errors ErrorComponentNames `yaml:"errors,omitempty"`
	// Echo names the echo-specific types users interact with.
	Echo EchoComponentNames `yaml:"echo,omitempty"`
	// StdHTTP names the std-http-specific types users interact with.
	StdHTTP StdHTTPComponentNames `yaml:"stdhttp,omitempty"`
	// Fiber names the fiber-specific types users interact with.
	Fiber FiberComponentNames `yaml:"fiber,omitempty"`
}

// ErrorComponentNames names the parameter-binding error types that the
// net/http-family server wrappers (chi, gorilla, std-http) hand to
// ErrorHandlerFunc. They are part of the API a user's error handler switches
// on, hence individually configurable.
type ErrorComponentNames struct {
	RequiredParamError         string `yaml:"required-param-error,omitempty"`
	RequiredHeaderError        string `yaml:"required-header-error,omitempty"`
	InvalidParamFormatError    string `yaml:"invalid-param-format-error,omitempty"`
	TooManyValuesForParamError string `yaml:"too-many-values-for-param-error,omitempty"`
	UnmarshalingParamError     string `yaml:"unmarshaling-param-error,omitempty"`
	UnescapedCookieParamError  string `yaml:"unescaped-cookie-param-error,omitempty"`
}

// EchoComponentNames names echo-specific generated types.
type EchoComponentNames struct {
	// Router is the `EchoRouter` interface accepted by RegisterHandlers.
	Router string `yaml:"router,omitempty"`
}

// StdHTTPComponentNames names std-http-specific generated types.
type StdHTTPComponentNames struct {
	// ServeMux is the `ServeMux` interface abstracting [http.ServeMux].
	ServeMux string `yaml:"serve-mux,omitempty"`
}

// FiberComponentNames names fiber-specific generated types.
type FiberComponentNames struct {
	// HandlerMiddlewareFunc is the per-handler middleware type users supply
	// via FiberServerOptions.HandlerMiddlewares.
	HandlerMiddlewareFunc string `yaml:"handler-middleware-func,omitempty"`
}

// setDefault assigns def to dst when the user did not supply a value. Every
// resolved field is assigned exactly once per Generate, so a stale value from
// a previous run is always overwritten (see the globalState caveat in
// Generate).
func setDefault(dst *string, def string) {
	if *dst == "" {
		*dst = def
	}
}

// prefixUnexported prepends prefix to an unexported default name, lowering the
// prefix's first letter and raising the name's so the result stays unexported:
// "PetStore" + "swaggerSpec" -> "petStoreSwaggerSpec".
func prefixUnexported(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return LowercaseFirstCharacter(prefix) + UppercaseFirstCharacter(name)
}

// resolve returns cn with every name filled in: defaults, then the prefix
// applied to whatever is still at its default, then family derivation from the
// resolved roots. Explicit overrides are never prefixed.
func (cn ComponentNames) resolve() ComponentNames {
	// exported prefixes an exported default name.
	exp := func(name string) string { return cn.Prefix + name }
	// unexp prefixes an unexported default name, preserving unexportedness.
	unexp := func(name string) string { return prefixUnexported(cn.Prefix, name) }

	// --- client family ---
	setDefault(&cn.Client, exp(defaultClientTypeName))
	setDefault(&cn.ClientInterface, cn.Client+"Interface")
	setDefault(&cn.ClientOption, cn.Client+"Option")
	setDefault(&cn.NewClient, "New"+cn.Client)
	setDefault(&cn.ClientWithResponses, cn.Client+"WithResponses")
	setDefault(&cn.ClientWithResponsesInterface, cn.ClientWithResponses+"Interface")
	setDefault(&cn.NewClientWithResponses, "New"+cn.ClientWithResponses)
	setDefault(&cn.RequestEditorFn, exp("RequestEditorFn"))
	setDefault(&cn.HTTPRequestDoer, exp("HttpRequestDoer"))
	setDefault(&cn.WithHTTPClient, exp("WithHTTPClient"))
	setDefault(&cn.WithRequestEditorFn, exp("WithRequestEditorFn"))
	setDefault(&cn.WithBaseURL, exp("WithBaseURL"))

	// --- shared server family ---
	setDefault(&cn.ServerInterface, exp("ServerInterface"))
	setDefault(&cn.ServerInterfaceWrapper, cn.ServerInterface+"Wrapper")
	setDefault(&cn.MiddlewareFunc, exp("MiddlewareFunc"))

	// --- net/http family ---
	setDefault(&cn.Handler, exp("Handler"))
	setDefault(&cn.HandlerFromMux, cn.Handler+"FromMux")
	setDefault(&cn.HandlerFromMuxWithBaseURL, cn.Handler+"FromMuxWithBaseURL")
	setDefault(&cn.HandlerWithOptions, cn.Handler+"WithOptions")
	setDefault(&cn.Unimplemented, exp("Unimplemented"))

	// --- register-handlers family ---
	setDefault(&cn.RegisterHandlers, exp("RegisterHandlers"))
	setDefault(&cn.RegisterHandlersWithBaseURL, cn.RegisterHandlers+"WithBaseURL")
	setDefault(&cn.RegisterHandlersWithOptions, cn.RegisterHandlers+"WithOptions")
	setDefault(&cn.RegisterHandlersOptions, cn.RegisterHandlers+"Options")

	// --- strict server ---
	setDefault(&cn.StrictServerInterface, exp("StrictServerInterface"))
	setDefault(&cn.StrictHandlerFunc, exp("StrictHandlerFunc"))
	setDefault(&cn.StrictMiddlewareFunc, exp("StrictMiddlewareFunc"))
	setDefault(&cn.NewStrictHandler, exp("NewStrictHandler"))
	setDefault(&cn.NewStrictHandlerWithOptions, exp("NewStrictHandlerWithOptions"))
	setDefault(&cn.StrictHandler, unexp("strictHandler"))
	setDefault(&cn.StrictHTTPServerOptions, exp("StrictHTTPServerOptions"))
	setDefault(&cn.StrictGinServerOptions, exp("StrictGinServerOptions"))

	// --- embedded spec ---
	setDefault(&cn.GetSwagger, exp("GetSwagger"))
	setDefault(&cn.GetSpec, exp("GetSpec"))
	setDefault(&cn.GetSpecJSON, exp("GetSpecJSON"))
	setDefault(&cn.PathToRawSpec, exp("PathToRawSpec"))
	setDefault(&cn.SwaggerSpec, unexp("swaggerSpec"))
	setDefault(&cn.RawSpec, unexp("rawSpec"))
	setDefault(&cn.DecodeSpec, unexp("decodeSpec"))
	setDefault(&cn.DecodeSpecCached, unexp("decodeSpecCached"))

	// --- per-framework option structs ---
	setDefault(&cn.ChiServerOptions, exp("ChiServerOptions"))
	setDefault(&cn.GorillaServerOptions, exp("GorillaServerOptions"))
	setDefault(&cn.StdHTTPServerOptions, exp("StdHTTPServerOptions"))
	setDefault(&cn.GinServerOptions, exp("GinServerOptions"))
	setDefault(&cn.FiberServerOptions, exp("FiberServerOptions"))
	setDefault(&cn.IrisServerOptions, exp("IrisServerOptions"))

	// --- groups ---
	setDefault(&cn.Errors.RequiredParamError, exp("RequiredParamError"))
	setDefault(&cn.Errors.RequiredHeaderError, exp("RequiredHeaderError"))
	setDefault(&cn.Errors.InvalidParamFormatError, exp("InvalidParamFormatError"))
	setDefault(&cn.Errors.TooManyValuesForParamError, exp("TooManyValuesForParamError"))
	setDefault(&cn.Errors.UnmarshalingParamError, exp("UnmarshalingParamError"))
	setDefault(&cn.Errors.UnescapedCookieParamError, exp("UnescapedCookieParamError"))
	setDefault(&cn.Echo.Router, exp("EchoRouter"))
	setDefault(&cn.StdHTTP.ServeMux, exp("ServeMux"))
	setDefault(&cn.Fiber.HandlerMiddlewareFunc, exp("HandlerMiddlewareFunc"))

	return cn
}

// suppliedComponentNames returns the user-supplied (non-empty, pre-resolution)
// values keyed by their YAML path, for identifier validation and error
// messages.
func suppliedComponentNames(cn ComponentNames) map[string]string {
	supplied := map[string]string{
		"prefix":                  cn.Prefix,
		"client":                  cn.Client,
		"server-interface":        cn.ServerInterface,
		"middleware-func":         cn.MiddlewareFunc,
		"handler":                 cn.Handler,
		"unimplemented":           cn.Unimplemented,
		"register-handlers":       cn.RegisterHandlers,
		"strict-server-interface": cn.StrictServerInterface,
		"get-swagger":             cn.GetSwagger,
		"get-spec":                cn.GetSpec,
		"get-spec-json":           cn.GetSpecJSON,

		"errors.required-param-error":            cn.Errors.RequiredParamError,
		"errors.required-header-error":           cn.Errors.RequiredHeaderError,
		"errors.invalid-param-format-error":      cn.Errors.InvalidParamFormatError,
		"errors.too-many-values-for-param-error": cn.Errors.TooManyValuesForParamError,
		"errors.unmarshaling-param-error":        cn.Errors.UnmarshalingParamError,
		"errors.unescaped-cookie-param-error":    cn.Errors.UnescapedCookieParamError,

		"echo.router":                   cn.Echo.Router,
		"stdhttp.serve-mux":             cn.StdHTTP.ServeMux,
		"fiber.handler-middleware-func": cn.Fiber.HandlerMiddlewareFunc,
	}
	for k, v := range supplied {
		if v == "" {
			delete(supplied, k)
		}
	}
	return supplied
}

// validateComponentNameIdentifiers checks that every supplied name is a legal
// Go identifier, and that the prefix additionally begins with a letter (it is
// spliced onto the front of other identifiers).
func validateComponentNameIdentifiers(cn ComponentNames) error {
	var errs []error
	supplied := suppliedComponentNames(cn)
	for _, key := range SortedMapKeys(supplied) {
		name := supplied[key]
		if !token.IsIdentifier(name) {
			errs = append(errs, fmt.Errorf("%s: %q is not a valid Go identifier", key, name))
			continue
		}
		if key == "prefix" && !unicode.IsLetter([]rune(name)[0]) {
			errs = append(errs, fmt.Errorf("prefix: %q must start with a letter", name))
		}
	}
	return errors.Join(errs...)
}

// EmittedComponentNames returns the resolved names that the given
// GenerateOptions will actually emit, keyed by a label naming the component
// (its YAML path when configurable, otherwise the default name plus the root
// it derives from). Names whose emission depends on the spec rather than the
// configuration -- the webhook/callback initiator and receiver families, which
// compose Prefix with "Webhook"/"Callback" -- are deliberately excluded: this
// map drives configuration-time uniqueness and schema-collision checks, which
// must not depend on spec contents.
//
// cn must already be resolved.
func EmittedComponentNames(cn ComponentNames, g GenerateOptions) map[string]string {
	netHTTPFamily := g.ChiServer || g.GorillaServer || g.StdHTTPServer
	echoFamily := g.EchoServer || g.Echo5Server
	fiberFamily := g.FiberServer || g.FiberV3Server
	registerFamily := echoFamily || g.GinServer || fiberFamily || g.IrisServer
	anyServer := netHTTPFamily || registerFamily

	emitted := map[string]string{}
	add := func(enabled bool, label, name string) {
		if enabled {
			emitted[label] = name
		}
	}

	add(g.Client, "client", cn.Client)
	add(g.Client, "client-interface (derived from `client`)", cn.ClientInterface)
	add(g.Client, "client-option (derived from `client`)", cn.ClientOption)
	add(g.Client, "new-client (derived from `client`)", cn.NewClient)
	add(g.Client, "client-with-responses (derived from `client`)", cn.ClientWithResponses)
	add(g.Client, "client-with-responses-interface (derived from `client`)", cn.ClientWithResponsesInterface)
	add(g.Client, "new-client-with-responses (derived from `client`)", cn.NewClientWithResponses)
	add(g.Client, "RequestEditorFn (prefix-only)", cn.RequestEditorFn)
	add(g.Client, "HttpRequestDoer (prefix-only)", cn.HTTPRequestDoer)
	add(g.Client, "WithHTTPClient (prefix-only)", cn.WithHTTPClient)
	add(g.Client, "WithRequestEditorFn (prefix-only)", cn.WithRequestEditorFn)
	add(g.Client, "WithBaseURL (prefix-only)", cn.WithBaseURL)

	add(anyServer, "server-interface", cn.ServerInterface)
	add(anyServer, "server-interface-wrapper (derived from `server-interface`)", cn.ServerInterfaceWrapper)
	add(anyServer, "middleware-func", cn.MiddlewareFunc)

	add(netHTTPFamily, "handler", cn.Handler)
	add(netHTTPFamily, "handler-from-mux (derived from `handler`)", cn.HandlerFromMux)
	add(netHTTPFamily, "handler-from-mux-with-base-url (derived from `handler`)", cn.HandlerFromMuxWithBaseURL)
	add(netHTTPFamily, "handler-with-options (derived from `handler`)", cn.HandlerWithOptions)
	add(netHTTPFamily, "errors.required-param-error", cn.Errors.RequiredParamError)
	add(netHTTPFamily, "errors.required-header-error", cn.Errors.RequiredHeaderError)
	add(netHTTPFamily, "errors.invalid-param-format-error", cn.Errors.InvalidParamFormatError)
	add(netHTTPFamily, "errors.too-many-values-for-param-error", cn.Errors.TooManyValuesForParamError)
	add(netHTTPFamily, "errors.unmarshaling-param-error", cn.Errors.UnmarshalingParamError)
	add(netHTTPFamily, "errors.unescaped-cookie-param-error", cn.Errors.UnescapedCookieParamError)
	add(g.ChiServer, "unimplemented", cn.Unimplemented)
	add(g.ChiServer, "ChiServerOptions (prefix-only)", cn.ChiServerOptions)
	add(g.GorillaServer, "GorillaServerOptions (prefix-only)", cn.GorillaServerOptions)
	add(g.StdHTTPServer, "stdhttp.serve-mux", cn.StdHTTP.ServeMux)
	add(g.StdHTTPServer, "StdHTTPServerOptions (prefix-only)", cn.StdHTTPServerOptions)

	add(registerFamily, "register-handlers", cn.RegisterHandlers)
	add(registerFamily, "register-handlers-with-options (derived from `register-handlers`)", cn.RegisterHandlersWithOptions)
	add(echoFamily, "register-handlers-with-base-url (derived from `register-handlers`)", cn.RegisterHandlersWithBaseURL)
	add(echoFamily, "register-handlers-options (derived from `register-handlers`)", cn.RegisterHandlersOptions)
	add(echoFamily, "echo.router", cn.Echo.Router)
	add(g.GinServer, "GinServerOptions (prefix-only)", cn.GinServerOptions)
	add(fiberFamily, "FiberServerOptions (prefix-only)", cn.FiberServerOptions)
	add(fiberFamily, "fiber.handler-middleware-func", cn.Fiber.HandlerMiddlewareFunc)
	add(g.IrisServer, "IrisServerOptions (prefix-only)", cn.IrisServerOptions)

	add(g.Strict, "strict-server-interface", cn.StrictServerInterface)
	add(g.Strict, "StrictHandlerFunc (prefix-only)", cn.StrictHandlerFunc)
	add(g.Strict, "StrictMiddlewareFunc (prefix-only)", cn.StrictMiddlewareFunc)
	add(g.Strict, "NewStrictHandler (prefix-only)", cn.NewStrictHandler)
	add(g.Strict, "strictHandler (prefix-only)", cn.StrictHandler)
	add(g.Strict && (netHTTPFamily || g.GinServer), "NewStrictHandlerWithOptions (prefix-only)", cn.NewStrictHandlerWithOptions)
	add(g.Strict && netHTTPFamily, "StrictHTTPServerOptions (prefix-only)", cn.StrictHTTPServerOptions)
	add(g.Strict && g.GinServer, "StrictGinServerOptions (prefix-only)", cn.StrictGinServerOptions)

	add(g.EmbeddedSpec, "get-swagger", cn.GetSwagger)
	add(g.EmbeddedSpec, "get-spec", cn.GetSpec)
	add(g.EmbeddedSpec, "get-spec-json", cn.GetSpecJSON)
	add(g.EmbeddedSpec, "PathToRawSpec (prefix-only)", cn.PathToRawSpec)
	add(g.EmbeddedSpec, "swaggerSpec (prefix-only)", cn.SwaggerSpec)
	add(g.EmbeddedSpec, "rawSpec (prefix-only)", cn.RawSpec)
	add(g.EmbeddedSpec, "decodeSpec (prefix-only)", cn.DecodeSpec)
	add(g.EmbeddedSpec, "decodeSpecCached (prefix-only)", cn.DecodeSpecCached)

	return emitted
}

// validateComponentNameUniqueness reports resolved names that two different
// components share. Two components mapping to one identifier is a
// configuration error, caught here rather than by the Go compiler.
func validateComponentNameUniqueness(cn ComponentNames, g GenerateOptions) error {
	byName := map[string][]string{}
	emitted := EmittedComponentNames(cn, g)
	for _, label := range SortedMapKeys(emitted) {
		byName[emitted[label]] = append(byName[emitted[label]], label)
	}

	var clashes []string
	for name, labels := range byName {
		if len(labels) > 1 {
			sort.Strings(labels)
			clashes = append(clashes, fmt.Sprintf("%q is used by %s", name, strings.Join(labels, " and ")))
		}
	}
	if len(clashes) == 0 {
		return nil
	}
	sort.Strings(clashes)
	return fmt.Errorf("resolved component names must be unique: %s", strings.Join(clashes, "; "))
}

// resolveComponentNames validates and resolves the component names for the
// given configuration. It is the single entry point used by both
// Configuration.Validate and Generate.
func resolveComponentNames(opts Configuration) (ComponentNames, error) {
	oo := opts.OutputOptions
	if err := validateComponentNameIdentifiers(oo.ComponentNames); err != nil {
		return ComponentNames{}, err
	}
	if oo.ClientTypeName != "" {
		if !token.IsIdentifier(oo.ClientTypeName) {
			return ComponentNames{}, fmt.Errorf("client-type-name: %q is not a valid Go identifier", oo.ClientTypeName)
		}
		if oo.ComponentNames.Client != "" && oo.ComponentNames.Client != oo.ClientTypeName {
			return ComponentNames{}, fmt.Errorf(
				"`client-type-name` (%q) and `component-names.client` (%q) are both set and disagree; "+
					"drop `client-type-name`, which `component-names.client` supersedes",
				oo.ClientTypeName, oo.ComponentNames.Client)
		}
	}

	resolved := oo.ComponentNames.resolve()

	// The legacy `client-type-name` renames only the client struct, leaving
	// ClientInterface, NewClient and the rest of the family at their defaults.
	// Applying it after derivation preserves that behavior exactly;
	// `component-names.client` (which is folded in before derivation) is the
	// knob that renames the whole family.
	if oo.ClientTypeName != "" && oo.ComponentNames.Client == "" {
		resolved.Client = oo.ClientTypeName
	}

	if err := validateComponentNameUniqueness(resolved, opts.Generate); err != nil {
		return ComponentNames{}, err
	}
	return resolved, nil
}
