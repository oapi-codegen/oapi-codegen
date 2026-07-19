package codegen

import "fmt"

// ClientMethodVariant is a precomputed view of one generated client method.
// Every operation yields a generic variant (the bodyless method, or the
// "WithBody" method taking an io.Reader when the operation has a request body)
// followed by one typed-body variant per request body that
// RequestBodyDefinition.IsSupportedByClient reports.
//
// All naming, signature, and comment decisions are made here in Go so that
// client.tmpl and client-with-responses.tmpl can render linearly, without the
// repeated `{{if .HasBody}}WithBody{{end}}` suffix dance and four-fragment
// signature assembly the templates previously carried at every declaration,
// implementation, and call site.
type ClientMethodVariant struct {
	// Suffix is appended to the OperationId to form the method name and its
	// request-builder name: "" or "WithBody" for the generic variant,
	// "WithJSONBody" etc. for a typed-body variant.
	Suffix string

	// ArgsDecl is the parameter-declaration fragment that follows the leading
	// ctx/server parameter, with a leading comma when non-empty, e.g.
	// ", id string, params *FooParams, contentType string, body io.Reader".
	ArgsDecl string

	// CallArgs is the argument fragment passed when a method forwards to
	// another function (the request builder, or the wrapped client method),
	// with a leading comma when non-empty, e.g. ", id, params, contentType, body".
	CallArgs string

	// InterfaceComment / MethodComment are the fully rendered Godoc comments
	// (including any deprecation notice) for the ClientInterface declaration and
	// the *Client method implementation respectively. The interface form places
	// a blank "//" line before the deprecation notice; the implementation form
	// does not -- matching the historical output of both sites.
	InterfaceComment string
	MethodComment    string

	// WithResponseInterfaceComment / WithResponseMethodComment are the analogous
	// comments for the ClientWithResponsesInterface declaration and the
	// *ClientWithResponses method implementation.
	WithResponseInterfaceComment string
	WithResponseMethodComment    string
}

// clientMethodComment assembles a rendered method comment from the base Godoc
// comment and an optional deprecation notice. When interfaceStyle is true a
// blank "//" comment line separates the two, mirroring the ClientInterface
// declarations; the method-implementation sites omit it.
func clientMethodComment(base, deprecation string, interfaceStyle bool) string {
	if deprecation == "" {
		return base
	}
	if interfaceStyle {
		return base + "\n//\n" + deprecation
	}
	return base + "\n" + deprecation
}

// ClientMethodVariants returns the precomputed client method variants for this
// operation: the generic variant first, then one per client-supported request
// body. It is consumed by client.tmpl and client-with-responses.tmpl.
func (o OperationDefinition) ClientMethodVariants() []ClientMethodVariant {
	pathDecl := genParamArgs(o.PathParams)
	pathCall := genParamNames(o.PathParams)

	var paramsDecl, paramsCall string
	if o.RequiresParamObject() {
		paramsDecl = fmt.Sprintf(", params *%sParams", o.OperationId)
		paramsCall = ", params"
	}

	deprecation := o.DeprecationComment()

	variants := make([]ClientMethodVariant, 0, 1+len(o.Bodies))

	// Generic variant: bodyless, or "WithBody" taking a raw io.Reader.
	var genericSuffix, genericBodyDecl, genericBodyCall string
	if o.HasBody() {
		genericSuffix = "WithBody"
		genericBodyDecl = ", contentType string, body io.Reader"
		genericBodyCall = ", contentType, body"
	}
	genericClientBase := o.GenerateFunctionComment(o.OperationId, genericSuffix, false)
	genericRespBase := o.GenerateFunctionComment(o.OperationId, genericSuffix+"WithResponse", true)
	variants = append(variants, ClientMethodVariant{
		Suffix:                       genericSuffix,
		ArgsDecl:                     pathDecl + paramsDecl + genericBodyDecl,
		CallArgs:                     pathCall + paramsCall + genericBodyCall,
		InterfaceComment:             clientMethodComment(genericClientBase, deprecation, true),
		MethodComment:                clientMethodComment(genericClientBase, deprecation, false),
		WithResponseInterfaceComment: clientMethodComment(genericRespBase, deprecation, true),
		// The generic ClientWithResponses method implementation historically
		// emits the blank "//" separator before the deprecation notice, unlike
		// the plain-Client method implementation and the typed-body variants
		// below. Preserved verbatim to keep generated output byte-identical.
		WithResponseMethodComment: clientMethodComment(genericRespBase, deprecation, true),
	})

	// Typed-body variants, one per client-supported request body.
	for _, body := range o.Bodies {
		if !body.IsSupportedByClient() {
			continue
		}
		suffix := body.Suffix()
		bodyDecl := fmt.Sprintf(", body %s%sRequestBody", o.OperationId, body.NameTag)
		clientBase := body.GenerateFunctionComment(o.OperationId, o, suffix, false)
		respBase := body.GenerateFunctionComment(o.OperationId, o, suffix+"WithResponse", true)
		variants = append(variants, ClientMethodVariant{
			Suffix:                       suffix,
			ArgsDecl:                     pathDecl + paramsDecl + bodyDecl,
			CallArgs:                     pathCall + paramsCall + ", body",
			InterfaceComment:             clientMethodComment(clientBase, deprecation, true),
			MethodComment:                clientMethodComment(clientBase, deprecation, false),
			WithResponseInterfaceComment: clientMethodComment(respBase, deprecation, true),
			WithResponseMethodComment:    clientMethodComment(respBase, deprecation, false),
		})
	}

	return variants
}

// GenericClientVariant returns the generic (bodyless or "WithBody") client
// method variant, which every operation always has as its first variant. The
// request-builder template uses it for the New{OperationId}Request{Suffix}
// signature.
func (o OperationDefinition) GenericClientVariant() ClientMethodVariant {
	return o.ClientMethodVariants()[0]
}
