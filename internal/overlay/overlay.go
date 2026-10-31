// Package overlay applies OpenAPI Overlay 1.0.0 documents
// (https://spec.openapis.org/overlay/v1.0.0.html) to OpenAPI documents.
//
// Documents are handled as the values encoding/json decodes into an any:
// map[string]any, []any, string, json.Number, bool and nil. Decode them with
// json.Decoder.UseNumber so that numbers keep their exact text.
//
// Action targets are RFC 9535 JSONPath expressions, evaluated with
// github.com/theory/jsonpath. Options.LegacyJSONPath accepts the dialect that
// oapi-codegen used before instead, by translating each target with package
// legacyjsonpath.
package overlay

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/oasdiff/yaml"
)

// Overlay is a parsed Overlay document.
type Overlay struct {
	// Version is the value of the overlay field, such as "1.0.0".
	Version string
	Info    Info
	// Extends is parsed and validated, but not used: the overlay is applied
	// to whichever document it is given.
	Extends string
	Actions []Action
}

// Info is the overlay's info object.
type Info struct {
	Title   string
	Version string
}

// Action is one entry of the overlay's actions.
type Action struct {
	// Target is the JSONPath expression that selects the nodes the action
	// applies to.
	Target      string
	Description string
	// Update is the value merged into each target. It is only used when
	// HasUpdate is true, which tells an absent update from update: null.
	Update    any
	HasUpdate bool
	// Remove removes each target from its parent.
	Remove bool

	// errs are problems found while parsing, reported by Validate.
	errs []error
}

// ParseFile reads and parses the overlay document at path, which may be YAML
// or JSON.
func ParseFile(path string) (*Overlay, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read overlay: %w", err)
	}
	return Parse(data)
}

// Parse parses an overlay document, which may be YAML or JSON. It reports
// syntax errors only; call Validate to check the document's content.
func Parse(data []byte) (*Overlay, error) {
	// YAMLToJSON turns mapping keys that YAML reads as numbers or booleans,
	// such as response codes written as 200:, into strings.
	jsonData, err := yaml.YAMLToJSON(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse overlay: %w", err)
	}
	raw, err := DecodeJSON(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse overlay: %w", err)
	}
	doc, ok := raw.(map[string]any)
	if !ok {
		return nil, errors.New("failed to parse overlay: the document is not a mapping")
	}

	o := &Overlay{}
	var errs []error
	o.Version, errs = stringField(doc, "overlay", "overlay version", errs)
	o.Extends, errs = stringField(doc, "extends", "overlay extends", errs)
	if info, ok := doc["info"].(map[string]any); ok {
		o.Info.Title, errs = scalarField(info, "title", "overlay info title", errs)
		o.Info.Version, errs = scalarField(info, "version", "overlay info version", errs)
	} else if v, present := doc["info"]; present && v != nil {
		errs = append(errs, errors.New("overlay info must be a mapping"))
	}

	switch actions := doc["actions"].(type) {
	case nil:
	case []any:
		for i, a := range actions {
			o.Actions = append(o.Actions, parseAction(i, a))
		}
	default:
		errs = append(errs, errors.New("overlay actions must be a sequence"))
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("failed to parse overlay: %w", errors.Join(errs...))
	}
	return o, nil
}

func parseAction(i int, v any) Action {
	m, ok := v.(map[string]any)
	if !ok {
		return Action{errs: []error{fmt.Errorf("overlay action at index %d must be a mapping", i)}}
	}
	var a Action
	what := fmt.Sprintf("overlay action at index %d", i)
	a.Target, a.errs = stringField(m, "target", what+" target", a.errs)
	a.Description, a.errs = scalarField(m, "description", what+" description", a.errs)
	a.Update, a.HasUpdate = m["update"]
	if _, has := m["copy"]; has {
		// Overlay 1.1.0 drafts add a copy action; the previous implementation
		// accepted the version without applying it in oapi-codegen's flow.
		a.errs = append(a.errs, fmt.Errorf("%s uses copy, which is not supported", what))
	}
	switch r := m["remove"].(type) {
	case nil:
	case bool:
		a.Remove = r
	default:
		a.errs = append(a.errs, fmt.Errorf("%s remove must be a boolean", what))
	}
	return a
}

// stringField returns the string at m[key], or an error when it is set to
// anything else.
func stringField(m map[string]any, key, what string, errs []error) (string, []error) {
	switch v := m[key].(type) {
	case nil:
		return "", errs
	case string:
		return v, errs
	default:
		return "", append(errs, fmt.Errorf("%s must be a string", what))
	}
}

// scalarField is stringField that also takes numbers and booleans, as the
// previous YAML-based parser did for these fields.
func scalarField(m map[string]any, key, what string, errs []error) (string, []error) {
	switch v := m[key].(type) {
	case json.Number:
		return v.String(), errs
	case bool:
		return fmt.Sprint(v), errs
	default:
		return stringField(m, key, what, errs)
	}
}

// supportedVersion accepts Overlay 1.0.x, and 1.1.x for documents written for
// the previous implementation, which took "1.1.0" to mean RFC 9535 targets:
// what every overlay gets here.
var supportedVersion = regexp.MustCompile(`^1\.[01]\.\d+$`)

// Validate checks the overlay against the Overlay 1.0.0 specification and
// returns every problem it finds, joined into one error.
func (o *Overlay) Validate() error {
	var errs []error
	switch {
	case o.Version == "":
		errs = append(errs, errors.New("overlay version must be defined"))
	case !supportedVersion.MatchString(o.Version):
		errs = append(errs, fmt.Errorf("overlay version %q is not supported: it must be 1.0.x or 1.1.x", o.Version))
	}
	if o.Info.Title == "" {
		errs = append(errs, errors.New("overlay info title must be defined"))
	}
	if o.Info.Version == "" {
		errs = append(errs, errors.New("overlay info version must be defined"))
	}
	if o.Extends != "" {
		if _, err := url.Parse(o.Extends); err != nil {
			errs = append(errs, errors.New("overlay extends must be a valid URL"))
		}
	}
	if len(o.Actions) == 0 {
		errs = append(errs, errors.New("overlay must define at least one action"))
	}
	for i, a := range o.Actions {
		errs = append(errs, a.errs...)
		if len(a.errs) > 0 {
			continue
		}
		if a.Target == "" {
			errs = append(errs, fmt.Errorf("overlay action at index %d target must be defined", i))
		}
		if a.Remove && a.HasUpdate {
			errs = append(errs, fmt.Errorf("overlay action at index %d should not both set remove and define update", i))
		}
	}
	return errors.Join(errs...)
}

// DecodeJSON decodes JSON into the values Apply works on, keeping numbers
// as json.Number.
func DecodeJSON(data []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	if dec.More() {
		return nil, errors.New("unexpected data after the document")
	}
	return v, nil
}

// describe names an action in errors.
func (a Action) describe(i int) string {
	s := fmt.Sprintf("action at index %d (target %q", i, a.Target)
	if a.Description != "" {
		s += fmt.Sprintf(", description %q", a.Description)
	}
	return s + ")"
}

// UnmatchedTargetsError is returned by Apply in strict mode when the targets
// of some actions select nothing.
type UnmatchedTargetsError struct {
	// Actions describes each action whose target matched nothing.
	Actions []string
}

func (e *UnmatchedTargetsError) Error() string {
	return "the targets of these overlay actions did not match anything in the document:\n- " + strings.Join(e.Actions, "\n- ")
}
