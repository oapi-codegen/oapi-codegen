package constants

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	ContentTypesJSON    = []string{"application/json", "text/x-json", "application/problem+json"}
	ContentTypesHalJSON = []string{"application/hal+json"}
	ContentTypesYAML    = []string{"application/yaml", "application/x-yaml", "text/yaml", "text/x-yaml"}
	ContentTypesXML     = []string{"application/xml", "text/xml", "application/problems+xml"}

	ResponseTypeSuffix = "Response"

	TitleCaser = cases.Title(language.English)
)
