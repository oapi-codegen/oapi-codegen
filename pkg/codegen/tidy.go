package codegen

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

var tidyMemConfig Configuration

func tidy(s *openapi3.T, opts Configuration) {
	var b bytes.Buffer
	w := &b
	tidyMemConfig = opts
	if len(opts.OutputOptions.IncludeTags) > 0 {
		fmt.Fprintf(w, "\n------------------------\n")
		fmt.Fprintf(w, "Tidying for %+v", opts.OutputOptions.IncludeTags)
		fmt.Fprintf(w, "\n------------------------\n")
	}
	for _, rule := range opts.Tidy.Functions {
		fmt.Fprintf(w, "\nFunctions tidy rule:\nReplace '%s' with '%s'\n\n", rule.Replace, rule.With)
		tidyPaths(w, s, rule, true)
	}
	for _, rule := range opts.Tidy.Params {
		fmt.Fprintf(w, "\n\nParams tidy rule:\nReplace '%s' with '%s'\n\n", rule.Replace, rule.With)
		tidyPaths(w, s, rule, false)
	}
	for _, rule := range opts.Tidy.Schemas {
		fmt.Fprintf(w, "\n\nSchema tidy rule:\nReplace '%s' with '%s'\n\n", rule.Replace, rule.With)
		tidySchemas(w, s, rule)
	}

	if opts.Tidy.Verbose {
		fmt.Println(w.String())
	}
}

func _tidy(w io.Writer, rule TidyRule, s string) string {
	fmt.Fprintf(w, "- %s", s)

	if rule.Match && s == rule.Replace {
		s = rule.With
	}
	if rule.All {
		s = strings.ReplaceAll(s, rule.Replace, rule.With)
	}
	if rule.Prefix && strings.HasPrefix(s, rule.Replace) {
		s = strings.Replace(s, rule.Replace, rule.With, 1)
	}
	if rule.Suffix && strings.HasSuffix(s, rule.Replace) {
		s = strings.TrimSuffix(s, rule.Replace) + rule.With
	}

	fmt.Fprintf(w, " -> %s\n", s)
	return s
}

func tidyFieldName(s string) string {
	var b bytes.Buffer
	w := &b
	for _, rule := range tidyMemConfig.Tidy.Schemas {
		s = _tidy(w, rule, s)
	}
	if tidyMemConfig.Tidy.Verbose {
		fmt.Println(w.String())
	}
	return s
}

func tidyPaths(w io.Writer, s *openapi3.T, rule TidyRule, tidyFns bool) {
	paths := s.Paths
	for key, path := range s.Paths {
		nkey := key
		tidyOperation(w, &nkey, path.Get, rule, tidyFns)
		tidyOperation(w, &nkey, path.Patch, rule, tidyFns)
		tidyOperation(w, &nkey, path.Post, rule, tidyFns)
		tidyOperation(w, &nkey, path.Put, rule, tidyFns)
		tidyOperation(w, &nkey, path.Delete, rule, tidyFns)
		tidyOperation(w, &nkey, path.Head, rule, tidyFns)
		tidyOperation(w, &nkey, path.Options, rule, tidyFns)
		tidyOperation(w, &nkey, path.Trace, rule, tidyFns)
		delete(paths, key)
		paths[nkey] = path
	}
	s.Paths = paths
}

func tidyOperation(w io.Writer, key *string, o *openapi3.Operation, rule TidyRule, tidyFns bool) {
	if o == nil {
		return
	}
	if tidyFns {
		o.OperationID = _tidy(w, rule, o.OperationID)
		return
	}
	beforeAndAfter := map[string]string{}
	for _, param := range o.Parameters {
		if param.Value == nil {
			continue
		}
		v := param.Value.Name
		beforeAndAfter[v] = _tidy(w, rule, param.Value.Name)
		param.Value.Name = beforeAndAfter[v]
	}
	for k, v := range beforeAndAfter {
		ns := *key
		*key = strings.ReplaceAll(ns, fmt.Sprintf("{%s}", k), fmt.Sprintf("{%s}", v))
	}
}

func tidySchemas(w io.Writer, s *openapi3.T, rule TidyRule) {
	tidySchemasInPaths(w, s, rule)
	tidySchemasInComp(w, s, rule)
}

func tidySchemasInPaths(w io.Writer, s *openapi3.T, rule TidyRule) {
	for k, path := range s.Paths {
		_ = k
		tidySchemaInOperation(w, path.Get, rule)
		tidySchemaInOperation(w, path.Patch, rule)
		tidySchemaInOperation(w, path.Post, rule)
		tidySchemaInOperation(w, path.Put, rule)
		tidySchemaInOperation(w, path.Delete, rule)
		tidySchemaInOperation(w, path.Head, rule)
		tidySchemaInOperation(w, path.Options, rule)
		tidySchemaInOperation(w, path.Trace, rule)
	}
}

func tidySchemasInComp(w io.Writer, s *openapi3.T, rule TidyRule) {
	sc := s.Components.Schemas
	for k, v := range sc {
		newK := _tidy(w, rule, k)
		delete(s.Components.Schemas, k)
		if v.Value != nil {
			properties := v.Value.Properties
			if v.Value != nil {
				for _, pv := range properties {
					pv.Ref = tidySchemaRef(w, pv.Ref, rule)
					if pv.Value.Items != nil {
						pv.Value.Items.Ref = tidySchemaRef(w, pv.Value.Items.Ref, rule)
					}
				}
			}
		}
		s.Components.Schemas[newK] = v
	}
}

func tidySchemaInOperation(w io.Writer, o *openapi3.Operation, rule TidyRule) {
	if o == nil {
		return
	}
	for _, param := range o.Parameters {
		param.Ref = tidySchemaRef(w, param.Ref, rule)
	}
	if o.RequestBody != nil {
		o.RequestBody.Ref = tidySchemaRef(w, o.RequestBody.Ref, rule)
	}

	for k, param := range o.Responses {
		param.Ref = tidySchemaRef(w, param.Ref, rule)
		if param.Value != nil {
			for _, cv := range param.Value.Content {
				if cv.Schema != nil {
					cv.Schema.Ref = tidySchemaRef(w, cv.Schema.Ref, rule)
				}
			}
		}
		o.Responses[k] = param
	}
}

func tidySchemaRef(w io.Writer, ref string, rule TidyRule) string {
	if ref == "" || !strings.HasPrefix(ref, "#/components/schemas/") {
		return ref
	}
	t := strings.TrimPrefix(ref, "#/components/schemas/")
	return "#/components/schemas/" + _tidy(w, rule, t)
}
