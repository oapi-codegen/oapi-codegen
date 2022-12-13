package codegen

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

func tidy(s *openapi3.T, opts Configuration) {
	var b bytes.Buffer
	w := &b
	for _, rule := range opts.Tidy.Functions {
		fmt.Fprintf(w, "\nFunctions tidy rule:\nReplace '%s' with '%s'\n\n", rule.Replace, rule.With)
		tidyPaths(w, s, rule, true)
	}
	for _, rule := range opts.Tidy.Params {
		fmt.Fprintf(w, "\n\nParams tidy rule:\nReplace '%s' with '%s'\n\n", rule.Replace, rule.With)
		tidyPaths(w, s, rule, false)
	}

	if opts.Tidy.Verbose {
		fmt.Println(w.String())
	}
}

func _tidy(w io.Writer, rule TidyRule, s string) string {
	fmt.Fprintf(w, "- %s", s)
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
