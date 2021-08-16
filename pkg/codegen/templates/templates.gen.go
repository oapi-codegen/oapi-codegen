package templates

import "text/template"

var templates = map[string]string{"additional-properties.tmpl": `{{range .Types}}{{$addType := .Schema.AdditionalPropertiesType.TypeDecl}}

// Getter for additional properties for {{.TypeName}}. Returns the specified
// element and whether it was found
func (a {{.TypeName}}) Get(fieldName string) (value {{$addType}}, found bool) {
    if a.AdditionalProperties != nil {
        value, found = a.AdditionalProperties[fieldName]
    }
    return
}

// Setter for additional properties for {{.TypeName}}
func (a *{{.TypeName}}) Set(fieldName string, value {{$addType}}) {
    if a.AdditionalProperties == nil {
        a.AdditionalProperties = make(map[string]{{$addType}})
    }
    a.AdditionalProperties[fieldName] = value
}

// Override default JSON handling for {{.TypeName}} to handle AdditionalProperties
func (a *{{.TypeName}}) UnmarshalJSON(b []byte) error {
    object := make(map[string]json.RawMessage)
	err := json.Unmarshal(b, &object)
	if err != nil {
		return err
	}
{{range .Schema.Properties}}
    if raw, found := object["{{.JsonFieldName}}"]; found {
        err = json.Unmarshal(raw, &a.{{.GoFieldName}})
        if err != nil {
            return errors.Wrap(err, "error reading '{{.JsonFieldName}}'")
        }
        delete(object, "{{.JsonFieldName}}")
    }
{{end}}
    if len(object) != 0 {
        a.AdditionalProperties = make(map[string]{{$addType}})
        for fieldName, fieldBuf := range object {
            var fieldVal {{$addType}}
            err := json.Unmarshal(fieldBuf, &fieldVal)
            if err != nil {
                return errors.Wrap(err, fmt.Sprintf("error unmarshaling field %s", fieldName))
            }
            a.AdditionalProperties[fieldName] = fieldVal
        }
    }
	return nil
}

// Override default JSON handling for {{.TypeName}} to handle AdditionalProperties
func (a {{.TypeName}}) MarshalJSON() ([]byte, error) {
    var err error
    object := make(map[string]json.RawMessage)
{{range .Schema.Properties}}
{{if not .Required}}if a.{{.GoFieldName}} != nil { {{end}}
    object["{{.JsonFieldName}}"], err = json.Marshal(a.{{.GoFieldName}})
    if err != nil {
        return nil, errors.Wrap(err, fmt.Sprintf("error marshaling '{{.JsonFieldName}}'"))
    }
{{if not .Required}} }{{end}}
{{end}}
    for fieldName, field := range a.AdditionalProperties {
		object[fieldName], err = json.Marshal(field)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("error marshaling '%s'", fieldName))
		}
	}
	return json.Marshal(object)
}
{{end}}
`,
	"chi-handler.tmpl": `// Handler creates http.Handler with routing matching OpenAPI spec.
func Handler(si ServerInterface) http.Handler {
  return HandlerWithOptions(si, ChiServerOptions{})
}

type ChiServerOptions struct {
    BaseURL string
    BaseRouter chi.Router
    Middlewares []MiddlewareFunc
}

// HandlerFromMux creates http.Handler with routing matching OpenAPI spec based on the provided mux.
func HandlerFromMux(si ServerInterface, r chi.Router) http.Handler {
    return HandlerWithOptions(si, ChiServerOptions {
        BaseRouter: r,
    })
}

func HandlerFromMuxWithBaseURL(si ServerInterface, r chi.Router, baseURL string) http.Handler {
    return HandlerWithOptions(si, ChiServerOptions {
        BaseURL: baseURL,
        BaseRouter: r,
    })
}

// HandlerWithOptions creates http.Handler with additional options
func HandlerWithOptions(si ServerInterface, options ChiServerOptions) http.Handler {
r := options.BaseRouter

if r == nil {
r = chi.NewRouter()
}
{{if .}}wrapper := ServerInterfaceWrapper{
Handler: si,
HandlerMiddlewares: options.Middlewares,
}
{{end}}
{{range .}}r.Group(func(r chi.Router) {
r.{{.Method | lower | title }}(options.BaseURL+"{{.Path | swaggerUriToChiUri}}", wrapper.{{.OperationId}})
})
{{end}}
return r
}
`,
	"chi-interface.tmpl": `// ServerInterface represents all server handlers.
type ServerInterface interface {
{{range .}}{{.SummaryAsComment }}
// ({{.Method}} {{.Path}})
{{.OperationId}}(w http.ResponseWriter, r *http.Request{{genParamArgs .PathParams}}{{if .RequiresParamObject}}, params {{.OperationId}}Params{{end}})
{{end}}
}
`,
	"chi-middleware.tmpl": `// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
    Handler ServerInterface
    HandlerMiddlewares []MiddlewareFunc
}

type MiddlewareFunc func(http.HandlerFunc) http.HandlerFunc

{{range .}}{{$opid := .OperationId}}

// {{$opid}} operation middleware
func (siw *ServerInterfaceWrapper) {{$opid}}(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
  {{if or .RequiresParamObject (gt (len .PathParams) 0) }}
  var err error
  {{end}}

  {{range .PathParams}}// ------------- Path parameter "{{.ParamName}}" -------------
  var {{$varName := .GoVariableName}}{{$varName}} {{.TypeDef}}

  {{if .IsPassThrough}}
  {{$varName}} = chi.URLParam(r, "{{.ParamName}}")
  {{end}}
  {{if .IsJson}}
  err = json.Unmarshal([]byte(chi.URLParam(r, "{{.ParamName}}")), &{{$varName}})
  if err != nil {
    http.Error(w, "Error unmarshaling parameter '{{.ParamName}}' as JSON", http.StatusBadRequest)
    return
  }
  {{end}}
  {{if .IsStyled}}
  err = runtime.BindStyledParameter("{{.Style}}",{{.Explode}}, "{{.ParamName}}", chi.URLParam(r, "{{.ParamName}}"), &{{$varName}})
  if err != nil {
    http.Error(w, fmt.Sprintf("Invalid format for parameter {{.ParamName}}: %s", err), http.StatusBadRequest)
    return
  }
  {{end}}

  {{end}}

{{range .SecurityDefinitions}}
  ctx = context.WithValue(ctx, {{.ProviderName | ucFirst}}Scopes, {{toStringArray .Scopes}})
{{end}}

  {{if .RequiresParamObject}}
    // Parameter object where we will unmarshal all parameters from the context
    var params {{.OperationId}}Params

    {{range $paramIdx, $param := .QueryParams}}// ------------- {{if .Required}}Required{{else}}Optional{{end}} query parameter "{{.ParamName}}" -------------
      if paramValue := r.URL.Query().Get("{{.ParamName}}"); paramValue != "" {

      {{if .IsPassThrough}}
        params.{{.GoName}} = {{if not .Required}}&{{end}}paramValue
      {{end}}

      {{if .IsJson}}
        var value {{.TypeDef}}
        err = json.Unmarshal([]byte(paramValue), &value)
        if err != nil {
          http.Error(w, "Error unmarshaling parameter '{{.ParamName}}' as JSON", http.StatusBadRequest)
          return
        }

        params.{{.GoName}} = {{if not .Required}}&{{end}}value
      {{end}}
      }{{if .Required}} else {
          http.Error(w, "Query argument {{.ParamName}} is required, but not found", http.StatusBadRequest)
          return
      }{{end}}
      {{if .IsStyled}}
      err = runtime.BindQueryParameter("{{.Style}}", {{.Explode}}, {{.Required}}, "{{.ParamName}}", r.URL.Query(), &params.{{.GoName}})
      if err != nil {
        http.Error(w, fmt.Sprintf("Invalid format for parameter {{.ParamName}}: %s", err), http.StatusBadRequest)
        return
      }
      {{end}}
  {{end}}

    {{if .HeaderParams}}
      headers := r.Header

      {{range .HeaderParams}}// ------------- {{if .Required}}Required{{else}}Optional{{end}} header parameter "{{.ParamName}}" -------------
        if valueList, found := headers[http.CanonicalHeaderKey("{{.ParamName}}")]; found {
          var {{.GoName}} {{.TypeDef}}
          n := len(valueList)
          if n != 1 {
            http.Error(w, fmt.Sprintf("Expected one value for {{.ParamName}}, got %d", n), http.StatusBadRequest)
            return
          }

        {{if .IsPassThrough}}
          params.{{.GoName}} = {{if not .Required}}&{{end}}valueList[0]
        {{end}}

        {{if .IsJson}}
          err = json.Unmarshal([]byte(valueList[0]), &{{.GoName}})
          if err != nil {
            http.Error(w, "Error unmarshaling parameter '{{.ParamName}}' as JSON", http.StatusBadRequest)
            return
          }
        {{end}}

        {{if .IsStyled}}
          err = runtime.BindStyledParameterWithLocation("{{.Style}}",{{.Explode}}, "{{.ParamName}}", runtime.ParamLocationHeader, valueList[0], &{{.GoName}})
          if err != nil {
            http.Error(w, fmt.Sprintf("Invalid format for parameter {{.ParamName}}: %s", err), http.StatusBadRequest)
            return
          }
        {{end}}

          params.{{.GoName}} = {{if not .Required}}&{{end}}{{.GoName}}

        } {{if .Required}}else {
            http.Error(w, fmt.Sprintf("Header parameter {{.ParamName}} is required, but not found: %s", err), http.StatusBadRequest)
            return
        }{{end}}

      {{end}}
    {{end}}

    {{range .CookieParams}}
      var cookie *http.Cookie

      if cookie, err = r.Cookie("{{.ParamName}}"); err == nil {

      {{- if .IsPassThrough}}
        params.{{.GoName}} = {{if not .Required}}&{{end}}cookie.Value
      {{end}}

      {{- if .IsJson}}
        var value {{.TypeDef}}
        var decoded string
        decoded, err := url.QueryUnescape(cookie.Value)
        if err != nil {
          http.Error(w, "Error unescaping cookie parameter '{{.ParamName}}'", http.StatusBadRequest)
          return
        }

        err = json.Unmarshal([]byte(decoded), &value)
        if err != nil {
          http.Error(w, "Error unmarshaling parameter '{{.ParamName}}' as JSON", http.StatusBadRequest)
          return
        }

        params.{{.GoName}} = {{if not .Required}}&{{end}}value
      {{end}}

      {{- if .IsStyled}}
        var value {{.TypeDef}}
        err = runtime.BindStyledParameter("simple",{{.Explode}}, "{{.ParamName}}", cookie.Value, &value)
        if err != nil {
          http.Error(w, "Invalid format for parameter {{.ParamName}}: %s", http.StatusBadRequest)
          return
        }
        params.{{.GoName}} = {{if not .Required}}&{{end}}value
      {{end}}

      }

      {{- if .Required}} else {
        http.Error(w, "Query argument {{.ParamName}} is required, but not found", http.StatusBadRequest)
        return
      }
      {{- end}}
    {{end}}
  {{end}}

  var handler = func(w http.ResponseWriter, r *http.Request) {
    siw.Handler.{{.OperationId}}(w, r{{genParamNames .PathParams}}{{if .RequiresParamObject}}, params{{end}})
}

  for _, middleware := range siw.HandlerMiddlewares {
    handler = middleware(handler)
  }

  handler(w, r.WithContext(ctx))
}
{{end}}



`,
	"client-with-responses.tmpl": `// ClientWithResponses builds on ClientInterface to offer response payloads
type ClientWithResponses struct {
    ClientInterface
}

// NewClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling
func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
    client, err := NewClient(server, opts...)
    if err != nil {
        return nil, err
    }
    return &ClientWithResponses{client}, nil
}

// WithBaseURL overrides the baseURL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		newBaseURL, err := url.Parse(baseURL)
		if err != nil {
			return err
		}
		c.Server = newBaseURL.String()
		return nil
	}
}

// ClientWithResponsesInterface is the interface specification for the client with responses above.
type ClientWithResponsesInterface interface {
{{range . -}}
{{$hasParams := .RequiresParamObject -}}
{{$pathParams := .PathParams -}}
{{$opid := .OperationId -}}
    // {{$opid}} request{{if .HasBody}} with any body{{end}}
    {{$opid}}{{if .HasBody}}WithBody{{end}}WithResponse(ctx context.Context{{genParamArgs .PathParams}}{{if .RequiresParamObject}}, params *{{$opid}}Params{{end}}{{if .HasBody}}, contentType string, body io.Reader{{end}}, reqEditors... RequestEditorFn) (*{{genResponseTypeName $opid}}, error)
{{range .Bodies}}
    {{$opid}}{{.Suffix}}WithResponse(ctx context.Context{{genParamArgs $pathParams}}{{if $hasParams}}, params *{{$opid}}Params{{end}}, body {{$opid}}{{.NameTag}}RequestBody, reqEditors... RequestEditorFn) (*{{genResponseTypeName $opid}}, error)
{{end}}{{/* range .Bodies */}}
{{end}}{{/* range . $opid := .OperationId */}}
}

{{range .}}{{$opid := .OperationId}}{{$op := .}}
type {{$opid | ucFirst}}Response struct {
    Body         []byte
	HTTPResponse *http.Response
    {{- range getResponseTypeDefinitions .}}
    {{.TypeName}} *{{.Schema.TypeDecl}}
    {{- end}}
}

// Status returns HTTPResponse.Status
func (r {{$opid | ucFirst}}Response) Status() string {
    if r.HTTPResponse != nil {
        return r.HTTPResponse.Status
    }
    return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r {{$opid | ucFirst}}Response) StatusCode() int {
    if r.HTTPResponse != nil {
        return r.HTTPResponse.StatusCode
    }
    return 0
}
{{end}}


{{range .}}
{{$opid := .OperationId -}}
{{/* Generate client methods (with responses)*/}}

// {{$opid}}{{if .HasBody}}WithBody{{end}}WithResponse request{{if .HasBody}} with arbitrary body{{end}} returning *{{$opid}}Response
func (c *ClientWithResponses) {{$opid}}{{if .HasBody}}WithBody{{end}}WithResponse(ctx context.Context{{genParamArgs .PathParams}}{{if .RequiresParamObject}}, params *{{$opid}}Params{{end}}{{if .HasBody}}, contentType string, body io.Reader{{end}}, reqEditors... RequestEditorFn) (*{{genResponseTypeName $opid}}, error){
    rsp, err := c.{{$opid}}{{if .HasBody}}WithBody{{end}}(ctx{{genParamNames .PathParams}}{{if .RequiresParamObject}}, params{{end}}{{if .HasBody}}, contentType, body{{end}}, reqEditors...)
    if err != nil {
        return nil, err
    }
    return Parse{{genResponseTypeName $opid | ucFirst}}(rsp)
}

{{$hasParams := .RequiresParamObject -}}
{{$pathParams := .PathParams -}}
{{$bodyRequired := .BodyRequired -}}
{{range .Bodies}}
func (c *ClientWithResponses) {{$opid}}{{.Suffix}}WithResponse(ctx context.Context{{genParamArgs $pathParams}}{{if $hasParams}}, params *{{$opid}}Params{{end}}, body {{$opid}}{{.NameTag}}RequestBody, reqEditors... RequestEditorFn) (*{{genResponseTypeName $opid}}, error) {
    rsp, err := c.{{$opid}}{{.Suffix}}(ctx{{genParamNames $pathParams}}{{if $hasParams}}, params{{end}}, body, reqEditors...)
    if err != nil {
        return nil, err
    }
    return Parse{{genResponseTypeName $opid | ucFirst}}(rsp)
}
{{end}}

{{end}}{{/* operations */}}

{{/* Generate parse functions for responses*/}}
{{range .}}{{$opid := .OperationId}}

// Parse{{genResponseTypeName $opid | ucFirst}} parses an HTTP response from a {{$opid}}WithResponse call
func Parse{{genResponseTypeName $opid | ucFirst}}(rsp *http.Response) (*{{genResponseTypeName $opid}}, error) {
    bodyBytes, err := ioutil.ReadAll(rsp.Body)
    defer rsp.Body.Close()
    if err != nil {
        return nil, err
    }

    response := {{genResponsePayload $opid}}

    {{genResponseUnmarshal .}}

    return response, nil
}
{{end}}{{/* range . $opid := .OperationId */}}

`,
	"client.tmpl": `// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the server, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the server.
	Server string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HttpRequestDoer

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	RequestEditors []RequestEditorFn
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// Creates a new Client, with reasonable defaults
func NewClient(server string, opts ...ClientOption) (*Client, error) {
    // create a client with sane default values
    client := Client{
        Server: server,
    }
    // mutate client and add all optional params
    for _, o := range opts {
        if err := o(&client); err != nil {
            return nil, err
        }
    }
    // ensure the server URL always has a trailing slash
    if !strings.HasSuffix(client.Server, "/") {
        client.Server += "/"
    }
    // create httpClient, if not already present
    if client.Client == nil {
        client.Client = &http.Client{}
    }
    return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)
		return nil
	}
}

// The interface specification for the client above.
type ClientInterface interface {
{{range . -}}
{{$hasParams := .RequiresParamObject -}}
{{$pathParams := .PathParams -}}
{{$opid := .OperationId -}}
    // {{$opid}} request{{if .HasBody}} with any body{{end}}
    {{$opid}}{{if .HasBody}}WithBody{{end}}(ctx context.Context{{genParamArgs $pathParams}}{{if $hasParams}}, params *{{$opid}}Params{{end}}{{if .HasBody}}, contentType string, body io.Reader{{end}}, reqEditors... RequestEditorFn) (*http.Response, error)
{{range .Bodies}}
    {{$opid}}{{.Suffix}}(ctx context.Context{{genParamArgs $pathParams}}{{if $hasParams}}, params *{{$opid}}Params{{end}}, body {{$opid}}{{.NameTag}}RequestBody, reqEditors... RequestEditorFn) (*http.Response, error)
{{end}}{{/* range .Bodies */}}
{{end}}{{/* range . $opid := .OperationId */}}
}


{{/* Generate client methods */}}
{{range . -}}
{{$hasParams := .RequiresParamObject -}}
{{$pathParams := .PathParams -}}
{{$opid := .OperationId -}}

func (c *Client) {{$opid}}{{if .HasBody}}WithBody{{end}}(ctx context.Context{{genParamArgs $pathParams}}{{if $hasParams}}, params *{{$opid}}Params{{end}}{{if .HasBody}}, contentType string, body io.Reader{{end}}, reqEditors... RequestEditorFn) (*http.Response, error) {
    req, err := New{{$opid}}Request{{if .HasBody}}WithBody{{end}}(c.Server{{genParamNames .PathParams}}{{if $hasParams}}, params{{end}}{{if .HasBody}}, contentType, body{{end}})
    if err != nil {
        return nil, err
    }
    req = req.WithContext(ctx)
    if err := c.applyEditors(ctx, req, reqEditors); err != nil {
        return nil, err
    }
    return c.Client.Do(req)
}

{{range .Bodies}}
func (c *Client) {{$opid}}{{.Suffix}}(ctx context.Context{{genParamArgs $pathParams}}{{if $hasParams}}, params *{{$opid}}Params{{end}}, body {{$opid}}{{.NameTag}}RequestBody, reqEditors... RequestEditorFn) (*http.Response, error) {
    req, err := New{{$opid}}{{.Suffix}}Request(c.Server{{genParamNames $pathParams}}{{if $hasParams}}, params{{end}}, body)
    if err != nil {
        return nil, err
    }
    req = req.WithContext(ctx)
    if err := c.applyEditors(ctx, req, reqEditors); err != nil {
        return nil, err
    }
    return c.Client.Do(req)
}
{{end}}{{/* range .Bodies */}}
{{end}}

{{/* Generate request builders */}}
{{range .}}
{{$hasParams := .RequiresParamObject -}}
{{$pathParams := .PathParams -}}
{{$bodyRequired := .BodyRequired -}}
{{$opid := .OperationId -}}

{{range .Bodies}}
// New{{$opid}}Request{{.Suffix}} calls the generic {{$opid}} builder with {{.ContentType}} body
func New{{$opid}}Request{{.Suffix}}(server string{{genParamArgs $pathParams}}{{if $hasParams}}, params *{{$opid}}Params{{end}}, body {{$opid}}{{.NameTag}}RequestBody) (*http.Request, error) {
    var bodyReader io.Reader
    buf, err := json.Marshal(body)
    if err != nil {
        return nil, err
    }
    bodyReader = bytes.NewReader(buf)
    return New{{$opid}}RequestWithBody(server{{genParamNames $pathParams}}{{if $hasParams}}, params{{end}}, "{{.ContentType}}", bodyReader)
}
{{end}}

// New{{$opid}}Request{{if .HasBody}}WithBody{{end}} generates requests for {{$opid}}{{if .HasBody}} with any type of body{{end}}
func New{{$opid}}Request{{if .HasBody}}WithBody{{end}}(server string{{genParamArgs $pathParams}}{{if $hasParams}}, params *{{$opid}}Params{{end}}{{if .HasBody}}, contentType string, body io.Reader{{end}}) (*http.Request, error) {
    var err error
{{range $paramIdx, $param := .PathParams}}
    var pathParam{{$paramIdx}} string
    {{if .IsPassThrough}}
    pathParam{{$paramIdx}} = {{.GoVariableName}}
    {{end}}
    {{if .IsJson}}
    var pathParamBuf{{$paramIdx}} []byte
    pathParamBuf{{$paramIdx}}, err = json.Marshal({{.GoVariableName}})
    if err != nil {
        return nil, err
    }
    pathParam{{$paramIdx}} = string(pathParamBuf{{$paramIdx}})
    {{end}}
    {{if .IsStyled}}
    pathParam{{$paramIdx}}, err = runtime.StyleParamWithLocation("{{.Style}}", {{.Explode}}, "{{.ParamName}}", runtime.ParamLocationPath, {{.GoVariableName}})
    if err != nil {
        return nil, err
    }
    {{end}}
{{end}}
    serverURL, err := url.Parse(server)
    if err != nil {
        return nil, err
    }

    operationPath := fmt.Sprintf("{{genParamFmtString .Path}}"{{range $paramIdx, $param := .PathParams}}, pathParam{{$paramIdx}}{{end}})
    if operationPath[0] == '/' {
        operationPath = "." + operationPath
    }

    queryURL, err := serverURL.Parse(operationPath)
    if err != nil {
        return nil, err
    }

{{if .QueryParams}}
    queryValues := queryURL.Query()
{{range $paramIdx, $param := .QueryParams}}
    {{if not .Required}} if params.{{.GoName}} != nil { {{end}}
    {{if .IsPassThrough}}
    queryValues.Add("{{.ParamName}}", {{if not .Required}}*{{end}}params.{{.GoName}})
    {{end}}
    {{if .IsJson}}
    if queryParamBuf, err := json.Marshal({{if not .Required}}*{{end}}params.{{.GoName}}); err != nil {
        return nil, err
    } else {
        queryValues.Add("{{.ParamName}}", string(queryParamBuf))
    }

    {{end}}
    {{if .IsStyled}}
    if queryFrag, err := runtime.StyleParamWithLocation("{{.Style}}", {{.Explode}}, "{{.ParamName}}", runtime.ParamLocationQuery, {{if not .Required}}*{{end}}params.{{.GoName}}); err != nil {
        return nil, err
    } else if parsed, err := url.ParseQuery(queryFrag); err != nil {
       return nil, err
    } else {
       for k, v := range parsed {
           for _, v2 := range v {
               queryValues.Add(k, v2)
           }
       }
    }
    {{end}}
    {{if not .Required}}}{{end}}
{{end}}
    queryURL.RawQuery = queryValues.Encode()
{{end}}{{/* if .QueryParams */}}
    req, err := http.NewRequest("{{.Method}}", queryURL.String(), {{if .HasBody}}body{{else}}nil{{end}})
    if err != nil {
        return nil, err
    }

    {{if .HasBody}}req.Header.Add("Content-Type", contentType){{end}}
{{range $paramIdx, $param := .HeaderParams}}
    {{if not .Required}} if params.{{.GoName}} != nil { {{end}}
    var headerParam{{$paramIdx}} string
    {{if .IsPassThrough}}
    headerParam{{$paramIdx}} = {{if not .Required}}*{{end}}params.{{.GoName}}
    {{end}}
    {{if .IsJson}}
    var headerParamBuf{{$paramIdx}} []byte
    headerParamBuf{{$paramIdx}}, err = json.Marshal({{if not .Required}}*{{end}}params.{{.GoName}})
    if err != nil {
        return nil, err
    }
    headerParam{{$paramIdx}} = string(headerParamBuf{{$paramIdx}})
    {{end}}
    {{if .IsStyled}}
    headerParam{{$paramIdx}}, err = runtime.StyleParamWithLocation("{{.Style}}", {{.Explode}}, "{{.ParamName}}", runtime.ParamLocationHeader, {{if not .Required}}*{{end}}params.{{.GoName}})
    if err != nil {
        return nil, err
    }
    {{end}}
    req.Header.Set("{{.ParamName}}", headerParam{{$paramIdx}})
    {{if not .Required}}}{{end}}
{{end}}

{{range $paramIdx, $param := .CookieParams}}
    {{if not .Required}} if params.{{.GoName}} != nil { {{end}}
    var cookieParam{{$paramIdx}} string
    {{if .IsPassThrough}}
    cookieParam{{$paramIdx}} = {{if not .Required}}*{{end}}params.{{.GoName}}
    {{end}}
    {{if .IsJson}}
    var cookieParamBuf{{$paramIdx}} []byte
    cookieParamBuf{{$paramIdx}}, err = json.Marshal({{if not .Required}}*{{end}}params.{{.GoName}})
    if err != nil {
        return nil, err
    }
    cookieParam{{$paramIdx}} = url.QueryEscape(string(cookieParamBuf{{$paramIdx}}))
    {{end}}
    {{if .IsStyled}}
    cookieParam{{$paramIdx}}, err = runtime.StyleParamWithLocation("simple", {{.Explode}}, "{{.ParamName}}", runtime.ParamLocationCookie, {{if not .Required}}*{{end}}params.{{.GoName}})
    if err != nil {
        return nil, err
    }
    {{end}}
    cookie{{$paramIdx}} := &http.Cookie{
        Name:"{{.ParamName}}",
        Value:cookieParam{{$paramIdx}},
    }
    req.AddCookie(cookie{{$paramIdx}})
    {{if not .Required}}}{{end}}
{{end}}
    return req, nil
}

{{end}}{{/* Range */}}

func (c *Client) applyEditors(ctx context.Context, req *http.Request, additionalEditors []RequestEditorFn) error {
    for _, r := range c.RequestEditors {
        if err := r(ctx, req); err != nil {
            return err
        }
    }
    for _, r := range additionalEditors {
        if err := r(ctx, req); err != nil {
            return err
        }
    }
    return nil
}
`,
	"constants.tmpl": `{{- if gt (len .SecuritySchemeProviderNames) 0 }}
const (
{{range $ProviderName := .SecuritySchemeProviderNames}}
    {{- $ProviderName | ucFirst}}Scopes = "{{$ProviderName}}.Scopes"
{{end}}
)
{{end}}
{{if gt (len .EnumDefinitions) 0 }}
{{range $Enum := .EnumDefinitions}}
// Defines values for {{$Enum.TypeName}}.
const (
{{range $index, $value := $Enum.Schema.EnumValues}}
  {{$index}} {{$Enum.TypeName}} = {{$Enum.ValueWrapper}}{{$value}}{{$Enum.ValueWrapper}}
{{end}}
)
{{end}}
{{end}}
`,
	"imports.tmpl": `// Package {{.PackageName}} provides primitives to interact with the openapi HTTP API.
//
// Code generated by {{.ModuleName}} version {{.Version}} DO NOT EDIT.
package {{.PackageName}}

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	openapi_types "github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	{{- range .ExternalImports}}
	{{ . }}
	{{- end}}
)
`,
	"inline.tmpl": `// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{
{{range .SpecParts}}
    "{{.}}",{{end}}
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
    zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
    if err != nil {
        return nil, fmt.Errorf("error base64 decoding spec: %s", err)
    }
    zr, err := gzip.NewReader(bytes.NewReader(zipped))
    if err != nil {
        return nil, fmt.Errorf("error decompressing spec: %s", err)
    }
    var buf bytes.Buffer
    _, err = buf.ReadFrom(zr)
    if err != nil {
        return nil, fmt.Errorf("error decompressing spec: %s", err)
    }

    return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
    var res = make(map[string]func() ([]byte, error))
    if len(pathToFile) > 0 {
        res[pathToFile] = rawSpec
    }
    {{ if .ImportMapping }}
    pathPrefix := path.Dir(pathToFile)
    {{ end }}
    {{ range $key, $value := .ImportMapping }}
    for rawPath, rawFunc := range {{ $value.Name }}.PathToRawSpec(path.Join(pathPrefix, "{{ $key }}")) {
        if _, ok := res[rawPath]; ok {
            // it is not possible to compare functions in golang, so always overwrite the old value
        }
        res[rawPath] = rawFunc
    }
    {{- end }}
    return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
    var resolvePath = PathToRawSpec("")

    loader := openapi3.NewLoader()
    loader.IsExternalRefsAllowed = true
    loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
        var pathToFile = url.String()
        pathToFile = path.Clean(pathToFile)
        getSpec, ok := resolvePath[pathToFile]
        if !ok {
            err1 := fmt.Errorf("path not found: %s", pathToFile)
            return nil, err1
        }
        return getSpec()
    }
    var specData []byte
    specData, err = rawSpec()
    if err != nil {
        return
    }
    swagger, err = loader.LoadFromData(specData)
    if err != nil {
        return
    }
    return
}
`,
	"param-types.tmpl": `{{range .}}{{$opid := .OperationId}}
{{range .TypeDefinitions}}
// {{.TypeName}} defines parameters for {{$opid}}.
type {{.TypeName}} {{if and (opts.AliasTypes) (.CanAlias)}}={{end}} {{.Schema.TypeDecl}}
{{end}}
{{end}}
`,
	"register.tmpl": `

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface) {
    RegisterHandlersWithBaseURL(router, si, "")
}

// Registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string) {
{{if .}}
    wrapper := ServerInterfaceWrapper{
        Handler: si,
    }
{{end}}
{{range .}}router.{{.Method}}(baseURL + "{{.Path | swaggerUriToEchoUri}}", wrapper.{{.OperationId}})
{{end}}
}
`,
	"request-bodies.tmpl": `{{range .}}{{$opid := .OperationId}}
{{range .Bodies}}
{{with .TypeDef $opid}}
// {{.TypeName}} defines body for {{$opid}} for application/json ContentType.
type {{.TypeName}} {{if and (opts.AliasTypes) (.CanAlias)}}={{end}} {{.Schema.TypeDecl}}
{{end}}
{{end}}
{{end}}
`,
	"server-interface.tmpl": `// ServerInterface represents all server handlers.
type ServerInterface interface {
{{range .}}{{.SummaryAsComment }}
// ({{.Method}} {{.Path}})
{{.OperationId}}(ctx echo.Context{{genParamArgs .PathParams}}{{if .RequiresParamObject}}, params {{.OperationId}}Params{{end}}) error
{{end}}
}
`,
	"typedef.tmpl": `{{range .Types}}
{{ with .Schema.Description }}{{ . }}{{ else }}// {{.TypeName}} defines model for {{.JsonName}}.{{ end }}
type {{.TypeName}} {{if and (opts.AliasTypes) (.CanAlias)}}={{end}} {{.Schema.TypeDecl}}
{{end}}
`,
	"wrappers.tmpl": `// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
    Handler ServerInterface
}

{{range .}}{{$opid := .OperationId}}// {{$opid}} converts echo context to params.
func (w *ServerInterfaceWrapper) {{.OperationId}} (ctx echo.Context) error {
    var err error
{{range .PathParams}}// ------------- Path parameter "{{.ParamName}}" -------------
    var {{$varName := .GoVariableName}}{{$varName}} {{.TypeDef}}
{{if .IsPassThrough}}
    {{$varName}} = ctx.Param("{{.ParamName}}")
{{end}}
{{if .IsJson}}
    err = json.Unmarshal([]byte(ctx.Param("{{.ParamName}}")), &{{$varName}})
    if err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Error unmarshaling parameter '{{.ParamName}}' as JSON")
    }
{{end}}
{{if .IsStyled}}
    err = runtime.BindStyledParameterWithLocation("{{.Style}}",{{.Explode}}, "{{.ParamName}}", runtime.ParamLocationPath, ctx.Param("{{.ParamName}}"), &{{$varName}})
    if err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter {{.ParamName}}: %s", err))
    }
{{end}}
{{end}}

{{range .SecurityDefinitions}}
    ctx.Set({{.ProviderName | sanitizeGoIdentity | ucFirst}}Scopes, {{toStringArray .Scopes}})
{{end}}

{{if .RequiresParamObject}}
    // Parameter object where we will unmarshal all parameters from the context
    var params {{.OperationId}}Params
{{range $paramIdx, $param := .QueryParams}}// ------------- {{if .Required}}Required{{else}}Optional{{end}} query parameter "{{.ParamName}}" -------------
    {{if .IsStyled}}
    err = runtime.BindQueryParameter("{{.Style}}", {{.Explode}}, {{.Required}}, "{{.ParamName}}", ctx.QueryParams(), &params.{{.GoName}})
    if err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter {{.ParamName}}: %s", err))
    }
    {{else}}
    if paramValue := ctx.QueryParam("{{.ParamName}}"); paramValue != "" {
    {{if .IsPassThrough}}
    params.{{.GoName}} = {{if not .Required}}&{{end}}paramValue
    {{end}}
    {{if .IsJson}}
    var value {{.TypeDef}}
    err = json.Unmarshal([]byte(paramValue), &value)
    if err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Error unmarshaling parameter '{{.ParamName}}' as JSON")
    }
    params.{{.GoName}} = {{if not .Required}}&{{end}}value
    {{end}}
    }{{if .Required}} else {
        return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Query argument {{.ParamName}} is required, but not found"))
    }{{end}}
    {{end}}
{{end}}

{{if .HeaderParams}}
    headers := ctx.Request().Header
{{range .HeaderParams}}// ------------- {{if .Required}}Required{{else}}Optional{{end}} header parameter "{{.ParamName}}" -------------
    if valueList, found := headers[http.CanonicalHeaderKey("{{.ParamName}}")]; found {
        var {{.GoName}} {{.TypeDef}}
        n := len(valueList)
        if n != 1 {
            return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Expected one value for {{.ParamName}}, got %d", n))
        }
{{if .IsPassThrough}}
        params.{{.GoName}} = {{if not .Required}}&{{end}}valueList[0]
{{end}}
{{if .IsJson}}
        err = json.Unmarshal([]byte(valueList[0]), &{{.GoName}})
        if err != nil {
            return echo.NewHTTPError(http.StatusBadRequest, "Error unmarshaling parameter '{{.ParamName}}' as JSON")
        }
{{end}}
{{if .IsStyled}}
        err = runtime.BindStyledParameterWithLocation("{{.Style}}",{{.Explode}}, "{{.ParamName}}", runtime.ParamLocationHeader, valueList[0], &{{.GoName}})
        if err != nil {
            return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter {{.ParamName}}: %s", err))
        }
{{end}}
        params.{{.GoName}} = {{if not .Required}}&{{end}}{{.GoName}}
        } {{if .Required}}else {
            return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Header parameter {{.ParamName}} is required, but not found"))
        }{{end}}
{{end}}
{{end}}

{{range .CookieParams}}
    if cookie, err := ctx.Cookie("{{.ParamName}}"); err == nil {
    {{if .IsPassThrough}}
    params.{{.GoName}} = {{if not .Required}}&{{end}}cookie.Value
    {{end}}
    {{if .IsJson}}
    var value {{.TypeDef}}
    var decoded string
    decoded, err := url.QueryUnescape(cookie.Value)
    if err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Error unescaping cookie parameter '{{.ParamName}}'")
    }
    err = json.Unmarshal([]byte(decoded), &value)
    if err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Error unmarshaling parameter '{{.ParamName}}' as JSON")
    }
    params.{{.GoName}} = {{if not .Required}}&{{end}}value
    {{end}}
    {{if .IsStyled}}
    var value {{.TypeDef}}
    err = runtime.BindStyledParameterWithLocation("simple",{{.Explode}}, "{{.ParamName}}", runtime.ParamLocationCookie, cookie.Value, &value)
    if err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter {{.ParamName}}: %s", err))
    }
    params.{{.GoName}} = {{if not .Required}}&{{end}}value
    {{end}}
    }{{if .Required}} else {
        return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Query argument {{.ParamName}} is required, but not found"))
    }{{end}}

{{end}}{{/* .CookieParams */}}

{{end}}{{/* .RequiresParamObject */}}
    // Invoke the callback with all the unmarshalled arguments
    err = w.Handler.{{.OperationId}}(ctx{{genParamNames .PathParams}}{{if .RequiresParamObject}}, params{{end}})
    return err
}
{{end}}
`,
}

// Parse parses declared templates.
func Parse(t *template.Template) (*template.Template, error) {
	for name, s := range templates {
		var tmpl *template.Template
		if t == nil {
			t = template.New(name)
		}
		if name == t.Name() {
			tmpl = t
		} else {
			tmpl = t.New(name)
		}
		if _, err := tmpl.Parse(s); err != nil {
			return nil, err
		}
	}
	return t, nil
}

