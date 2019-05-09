package templates

import "text/template"

var templates = map[string]string{"client.tmpl": `// A client which conforms to the OpenAPI3 specification for this service. The
// server should be fully qualified with shema and server, ie,
// https://deepmap.com.
type Client struct {
    Server string
    Client http.Client
}

{{range .}}{{$opid := .OperationId}}
{{if .HasBody}}
// Request for {{$opid}} with JSON body
func (c *Client) {{$opid}}(ctx context.Context{{genParamArgs .PathParams}}{{if .RequiresParamObject}}, params *{{$opid}}Params{{end}}, body {{if not .GetBodyDefinition.Required}}*{{end}}{{if .GetBodyDefinition.CustomType}}{{$opid}}RequestBody{{else}}{{.GetBodyDefinition.TypeDef}}{{end}}) (*http.Response, error){
    req, err := New{{$opid}}Request(c.Server{{genParamNames .PathParams}}{{if .RequiresParamObject}}, params{{end}}, body)
    if err != nil {
        return nil, err
    }
    req = req.WithContext(ctx)
    return c.Client.Do(req)
}
{{end}}{{/* end of .HasBody */}}
{{if .GenerateGenericForm}}
// Request for {{$opid}}{{if .HasAnyBody}} with arbitrary body{{end}}
func (c *Client) {{$opid}}{{if .HasAnyBody}}WithBody{{end}}(ctx context.Context{{genParamArgs .PathParams}}{{if .RequiresParamObject}}, params *{{$opid}}Params{{end}}{{if .HasAnyBody}}, contentType string, body io.Reader{{end}}) (*http.Response, error){
    req, err := New{{$opid}}Request{{if .HasAnyBody}}WithBody{{end}}(c.Server{{genParamNames .PathParams}}{{if .RequiresParamObject}}, params{{end}}{{if .HasAnyBody}}, contentType, body{{end}})
    if err != nil {
        return nil, err
    }
    req = req.WithContext(ctx)
    return c.Client.Do(req)
}
{{end}}
{{end}}

{{range .}}{{$opid := .OperationId -}}
{{if .HasBody}}
// Request generator for {{$opid}} with JSON body
func New{{$opid}}Request(server string{{genParamArgs .PathParams}}{{if .RequiresParamObject}}, params *{{$opid}}Params{{end}}, body {{if not .GetBodyDefinition.Required}}*{{end}}{{if .GetBodyDefinition.CustomType}}{{$opid}}RequestBody{{else}}{{.GetBodyDefinition.TypeDef}}{{end}}) (*http.Request, error) {
    var bodyReader io.Reader
    {{if not .GetBodyDefinition.Required}}if body != nil { {{end}}
        buf, err := json.Marshal(body)
        if err != nil {
            return nil, err
        }
        bodyReader = bytes.NewReader(buf)
    {{if not .GetBodyDefinition.Required}}}{{end}}
        return New{{$opid}}RequestWithBody(server{{genParamNames .PathParams}}{{if .RequiresParamObject}}, params *{{$opid}}Params{{end}}, "application/json", bodyReader)
}{{end}}{{/* end of .HasBody */}}
// Request generator for {{$opid}}{{if .HasAnyBody}} with non-JSON body{{end}}
func New{{$opid}}Request{{if .HasAnyBody}}WithBody{{end}}(server string{{genParamArgs .PathParams}}{{if .RequiresParamObject}}, params *{{$opid}}Params{{end}}{{if .HasAnyBody}}, contentType string, body io.Reader{{end}}) (*http.Request, error) {
    var err error
{{range $paramIdx, $param := .PathParams}}
    var pathParam{{$paramIdx}} string
    {{if .IsPassThrough}}
    pathParam{{$paramIdx}} = {{.ParamName}}
    {{end}}
    {{if .IsJson}}
    var pathParamBuf{{$paramIdx}} []byte
    pathParamBuf{{$paramIdx}}, err = json.Marshal({{.ParamName}})
    if err != nil {
        return nil, err
    }
    pathParam{{$paramIdx}} = string(pathParamBuf{{$paramIdx}})
    {{end}}
    {{if .IsStyled}}
    pathParam{{$paramIdx}}, err = runtime.StyleParam("{{.Style}}", {{.Explode}}, "{{.ParamName}}", {{.ParamName | camelCase | lcFirst }})
    if err != nil {
        return nil, err
    }
    {{end}}
{{end}}
    queryUrl := fmt.Sprintf("%s{{genParamFmtString .Path}}", server{{range $paramIdx, $param := .PathParams}}, pathParam{{$paramIdx}}{{end}})
{{if .QueryParams}}
    var queryStrings []string
{{range $paramIdx, $param := .QueryParams}}
    var queryParam{{$paramIdx}} string
    {{if not .Required}} if params.{{.GoName}} != nil { {{end}}
    {{if .IsPassThrough}}
    queryParam{{$paramIdx}} = "{{.ParamName}}=" + {{if not .Required}}*{{end}}params.{{.GoName}}
    {{end}}
    {{if .IsJson}}
    var queryParamBuf{{$paramIdx}} []byte
    queryParamBuf{{$paramIdx}}, err = json.Marshal({{if not .Required}}*{{end}}params.{{.GoName}})
    if err != nil {
        return nil, err
    }
    queryParam{{$paramIdx}} = "{{.ParamName}}=" + string(queryParamBuf{{$paramIdx}})

    {{end}}
    {{if .IsStyled}}
    queryParam{{$paramIdx}}, err = runtime.StyleParam("{{.Style}}", {{.Explode}}, "{{.ParamName}}", {{if not .Required}}*{{end}}params.{{.GoName}})
    if err != nil {
        return nil, err
    }
    {{end}}
    queryStrings = append(queryStrings, queryParam{{$paramIdx}})
    {{if not .Required}}}{{end}}
{{end}}
    if len(queryStrings) != 0 {
        queryUrl += "?" + strings.Join(queryStrings, "&")
    }
{{end}}{{/* if .QueryParams */}}
    req, err := http.NewRequest("{{.Method}}", queryUrl, {{if .HasAnyBody}}body{{else}}nil{{end}})
    if err != nil {
        return nil, err
    }

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
    headerParam{{$paramIdx}}, err = runtime.StyleParam("{{.Style}}", {{.Explode}}, "{{.ParamName}}", {{if not .Required}}*{{end}}params.{{.GoName}})
    if err != nil {
        return nil, err
    }
    {{end}}
    req.Header.Add("{{.ParamName}}", headerParam{{$paramIdx}})
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
    cookieParam{{$paramIdx}}, err = runtime.StyleParam("simple", {{.Explode}}, "{{.ParamName}}", {{if not .Required}}*{{end}}params.{{.GoName}})
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
    {{if .HasAnyBody}}req.Header.Add("Content-Type", contentType){{end}}
    return req, nil
}

{{end}}{{/* Range */}}
`,
	"imports.tmpl": `// This is an autogenerated file, any edits which you make here will be lost!
package {{.PackageName}}

import (
{{range .Imports}} "{{.}}"
{{end}})
`,
	"inline.tmpl": `// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{
{{range .}}
    "{{.}}",{{end}}
}

// Returns the Swagger specification corresponding to the generated code
// in this file.
func GetSwagger() (*openapi3.Swagger, error) {
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

    swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData(buf.Bytes())
    if err != nil {
        return nil, fmt.Errorf("error loading Swagger: %s", err)
    }
    return swagger, nil
}
`,
	"parameters.tmpl": `{{range .Types}}
// Type definition for component parameter "{{.JsonTypeName}}"
type {{.TypeName}} {{.TypeDef}}
{{end}}
`,
	"register.tmpl": `func RegisterHandlers(router runtime.EchoRouter, si ServerInterface) {
{{if .}}
    wrapper := ServerInterfaceWrapper{
        Handler: si,
    }
{{end}}
{{range .}}router.{{.Method}}("{{.Path | swaggerUriToEchoUri}}", wrapper.{{.OperationId}})
{{end}}
}
`,
	"request-bodies.tmpl": `{{range .Types}}
// Type definition for component requestBodies "{{.JsonTypeName}}"
type {{.TypeName}} {{.TypeDef}}
{{end}}
`,
	"responses.tmpl": `{{range .Types}}
// Type definition for component response "{{.JsonTypeName}}"
type {{.TypeName}} {{.TypeDef}}
{{end}}
`,
	"schemas.tmpl": `{{range .Types}}
// Type definition for component schema "{{.JsonTypeName}}"
type {{.TypeName}} {{.TypeDef}}
{{end}}
`,
	"server-interface.tmpl": `{{range .}}

{{if .Params}}
// Parameters object for {{.OperationId}}
type {{.OperationId}}Params struct {
{{range .Params}}
    {{.GoName}} {{if not .Required}}*{{end}}{{.TypeDef}} {{.JsonTag}}{{end}}
}
{{end}}

{{if .HasBody}}
{{if .GetBodyDefinition.CustomType}}
// Request body for {{.OperationId}} for application/json ContentType
type {{.OperationId}}RequestBody {{.GetBodyDefinition.TypeDef}}
{{end}}
{{end}}

{{end}}

type ServerInterface interface {
{{range .}}// {{.Summary}} ({{.Method}} {{.Path}})
{{.OperationId}}(ctx echo.Context{{genParamArgs .PathParams}}{{if .RequiresParamObject}}, params {{.OperationId}}Params{{end}}) error
{{end}}
}
`,
	"wrappers.tmpl": `type ServerInterfaceWrapper struct {
    Handler ServerInterface
}

{{range .}}{{$opid := .OperationId}}// Wrapper for {{$opid}}
func (w *ServerInterfaceWrapper) {{.OperationId}} (ctx echo.Context) error {
    var err error
{{range .PathParams}}// ------------- Path parameter "{{.ParamName}}" -------------
    var {{$varName := .GoName | lcFirst}}{{$varName}} {{.TypeDef}}
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
    err = runtime.BindStyledParameter("{{.Style}}",{{.Explode}}, "{{.ParamName}}", ctx.Param("{{.ParamName}}"), &{{$varName}})
    if err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter {{.ParamName}}: %s", err))
    }
{{end}}
{{end}}

{{if .RequiresParamObject}}
    // Parameter object where we will unmarshal all parameters from the
    // context.
    var params {{.OperationId}}Params
{{range $paramIdx, $param := .QueryParams}}// ------------- {{if .Required}}Required{{else}}Optional{{end}} query parameter "{{.ParamName}}" -------------
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
    {{if .IsStyled}}
    err = runtime.BindQueryParameter("{{.Style}}", {{.Explode}}, {{.Required}}, "{{.ParamName}}", ctx.QueryParams(), &params.{{.GoName}})
    if err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter {{.ParamName}}: %s", err))
    }
    {{end}}
{{end}}

{{if .HeaderParams}}
    headers := ctx.Request().Header
{{range .HeaderParams}}// ------------- {{if .Required}}Required{{else}}Optional{{end}} header parameter "{{.ParamName}}" -------------
    if valueList, found := headers["{{.ParamName}}"]; found {
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
        err = runtime.BindStyledParameter("{{.Style}}",{{.Explode}}, "{{.ParamName}}", valueList[0], &{{.GoName}})
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
    err = runtime.BindStyledParameter("simple",{{.Explode}}, "{{.ParamName}}", cookie.Value, &value)
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

