package templates

import "embed"

//go:embed files/*
var TemplateFS embed.FS
