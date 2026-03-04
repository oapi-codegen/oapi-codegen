package codegen

import (
	"fmt"
	"strconv"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
)

const serverURLPrefix = "ServerUrl"
const serverURLSuffixIterations = 10

// ServerObjectDefinition defines the definition of an OpenAPI Server object (https://spec.openapis.org/oas/v3.0.3#server-object) as it is provided to code generation in `oapi-codegen`
type ServerObjectDefinition struct {
	// GoName is the name of the variable for this Server URL
	GoName string

	// OAPISchema is the underlying OpenAPI representation of the Server
	OAPISchema *openapi3.Server
}

func GenerateServerURLs(t *template.Template, spec *openapi3.T) (string, error) {
	names := make(map[string]*openapi3.Server)

	for _, server := range spec.Servers {
		suffix := server.Description
		if suffix == "" {
			suffix = nameNormalizer(server.URL)
		}
		name := serverURLPrefix + UppercaseFirstCharacter(suffix)
		name = nameNormalizer(name)

		// if this is the only type with this name, store it
		if _, conflict := names[name]; !conflict {
			names[name] = server
			continue
		}

		// otherwise, try appending a number to the name
		saved := false
		// NOTE that we start at 1 on purpose, as
		//
		//  ... ServerURLDevelopmentServer
		//  ... ServerURLDevelopmentServer1`
		//
		// reads better than:
		//
		//  ... ServerURLDevelopmentServer
		//  ... ServerURLDevelopmentServer0
		for i := 1; i < 1+serverURLSuffixIterations; i++ {
			suffixed := name + strconv.Itoa(i)
			// and then store it if there's no conflict
			if _, suffixConflict := names[suffixed]; !suffixConflict {
				names[suffixed] = server
				saved = true
				break
			}
		}

		if saved {
			continue
		}

		// otherwise, error
		return "", fmt.Errorf("failed to create a unique name for the Server URL (%#v) with description (%#v) after %d iterations", server.URL, server.Description, serverURLSuffixIterations)
	}

	keys := SortedMapKeys(names)
	servers := make([]ServerObjectDefinition, len(keys))
	i := 0
	for _, k := range keys {
		servers[i] = ServerObjectDefinition{
			GoName:     k,
			OAPISchema: names[k],
		}
		i++
	}

	return GenerateTemplates([]string{"server-urls.tmpl"}, t, servers)
}
