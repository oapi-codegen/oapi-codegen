package templates

import "embed"

// TemplateFS contains all embedded template files.
// The files/* pattern recursively includes all files in subdirectories.
//
//go:embed files/*
var TemplateFS embed.FS
