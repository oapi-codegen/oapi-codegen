package legacyjsonpath

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Translate converts target, written in the previous JSONPath dialect, to an
// RFC 9535 JSONPath expression. See the package documentation for the rules
// and their limits.
func Translate(target string) (string, error) {
	p := &parser{in: target}
	root := "$"
	switch {
	case target == "":
		return "", fmt.Errorf("empty target")
	case strings.HasPrefix(target, "$"):
		p.pos = 1
	default:
		// The previous dialect allowed an implicit root: paths.* means
		// $.paths.*.
		p.implicitRoot = true
	}
	segs, err := p.path(false)
	if err != nil {
		return "", p.wrap(err)
	}
	if !p.eof() {
		return "", p.wrap(p.errorf("unexpected %q", p.rest()))
	}
	out, err := emitPath(root, segs)
	if err != nil {
		return "", fmt.Errorf("cannot translate %q: %w", target, err)
	}
	return out, nil
}

// segment is one step of a parsed path.
type segment interface{ isSegment() }

// nameSeg selects members by name: .name, ['a'] or ['a','b'].
type nameSeg struct {
	names  []string
	dotted bool
}

// wildcardSeg selects every child: .* or [*].
type wildcardSeg struct{ dotted bool }

// subscriptSeg selects array elements by index, slice or union: [0], [1:3].
type subscriptSeg struct {
	raw    string
	single bool // a single index, which RFC 9535 counts as singular
}

// descendantSeg is recursive descent followed by its selector: ..name,
// ..*, ..[0], ..['a'] or ..[?(P)].
type descendantSeg struct{ inner segment }

// filterSeg is a filter, [?(P)].
type filterSeg struct {
	expr expr
	pos  int
}

func (nameSeg) isSegment()       {}
func (wildcardSeg) isSegment()   {}
func (subscriptSeg) isSegment()  {}
func (descendantSeg) isSegment() {}
func (filterSeg) isSegment()     {}

// expr is a node of a filter expression.
type expr interface{ isExpr() }

type orExpr struct{ terms []expr }
type andExpr struct{ terms []expr }
type notExpr struct{ x expr }
type parenExpr struct{ x expr }

// existExpr is a query used as a test: it is true when the query selects
// at least one node.
type existExpr struct{ q query }

// cmpExpr compares two terms.
type cmpExpr struct {
	op       string
	lhs, rhs term
}

// matchExpr is the previous dialect's regular expression match, X =~ /re/.
type matchExpr struct {
	q  query
	re string
}

func (orExpr) isExpr()    {}
func (andExpr) isExpr()   {}
func (notExpr) isExpr()   {}
func (parenExpr) isExpr() {}
func (existExpr) isExpr() {}
func (cmpExpr) isExpr()   {}
func (matchExpr) isExpr() {}

// term is either a query or a literal.
type term interface{ isTerm() }

// query is a path relative to @ or $ inside a filter.
type query struct {
	root string // "@" or "$"
	segs []segment
}

// literal is an RFC 9535 literal, already formatted.
type literal struct{ text string }

func (query) isTerm()   {}
func (literal) isTerm() {}

type parser struct {
	in           string
	pos          int
	implicitRoot bool
}

func (p *parser) eof() bool         { return p.pos >= len(p.in) }
func (p *parser) peek() byte        { return p.in[p.pos] }
func (p *parser) rest() string      { return p.in[p.pos:] }
func (p *parser) has(s string) bool { return strings.HasPrefix(p.in[p.pos:], s) }

func (p *parser) skipSpace() {
	for !p.eof() && (p.peek() == ' ' || p.peek() == '\t') {
		p.pos++
	}
}

type parseError struct {
	pos int
	msg string
}

func (e *parseError) Error() string { return e.msg }

func (p *parser) errorf(format string, args ...any) error {
	return &parseError{pos: p.pos, msg: fmt.Sprintf(format, args...)}
}

func (p *parser) wrap(err error) error {
	if pe, ok := err.(*parseError); ok {
		return fmt.Errorf("invalid target %q at position %d: %s", p.in, pe.pos, pe.msg)
	}
	return fmt.Errorf("invalid target %q: %w", p.in, err)
}

// isNameTerminator reports whether c ends a dot-notation name. These are the
// characters the previous dialect stopped at.
func isNameTerminator(c byte) bool {
	switch c {
	case '.', '[', ')', ' ', '&', '|', '=', '!', '>', '<', '~':
		return true
	}
	return false
}

// path parses segments until the end of the input or, inside a filter, the
// end of the query.
func (p *parser) path(inFilter bool) ([]segment, error) {
	var segs []segment
	first := true
	for !p.eof() {
		c := p.peek()
		switch {
		case p.has(".."):
			p.pos += 2
			seg, err := p.descendant()
			if err != nil {
				return nil, err
			}
			segs = append(segs, seg)
		case c == '.':
			p.pos++
			seg, err := p.dotted()
			if err != nil {
				return nil, err
			}
			segs = append(segs, seg)
		case c == '[':
			seg, err := p.bracket()
			if err != nil {
				return nil, err
			}
			segs = append(segs, seg)
		case first && p.implicitRoot && !inFilter:
			// An undotted name at the start of the target.
			seg, err := p.dotted()
			if err != nil {
				return nil, err
			}
			segs = append(segs, seg)
		case inFilter:
			return segs, nil
		default:
			return nil, p.errorf("unexpected %q", string(c))
		}
		first = false
	}
	return segs, nil
}

// dotted parses the name after a dot.
func (p *parser) dotted() (segment, error) {
	start := p.pos
	for !p.eof() && !isNameTerminator(p.peek()) {
		p.pos++
	}
	name := p.in[start:p.pos]
	if name == "" {
		return nil, p.errorf("child name missing")
	}
	if !p.eof() && p.peek() == '~' {
		return nil, p.errorf("the ~ property-name operator has no RFC 9535 equivalent")
	}
	if name == "*" {
		return wildcardSeg{dotted: true}, nil
	}
	return nameSeg{names: []string{unescapeName(name)}, dotted: true}, nil
}

// descendant parses what follows "..".
func (p *parser) descendant() (segment, error) {
	if p.eof() {
		return nil, p.errorf("child name or array access or filter missing after recursive descent")
	}
	var (
		inner segment
		err   error
	)
	if p.peek() == '[' {
		inner, err = p.bracket()
	} else {
		inner, err = p.dotted()
	}
	if err != nil {
		return nil, err
	}
	return descendantSeg{inner: inner}, nil
}

// bracket parses a bracketed selector: names, a filter, or a subscript.
func (p *parser) bracket() (segment, error) {
	open := p.pos
	p.pos++ // [
	p.skipSpace()
	if p.eof() {
		return nil, p.errorf("unmatched [")
	}
	switch c := p.peek(); {
	case c == '\'' || c == '"':
		var names []string
		for {
			p.skipSpace()
			if p.eof() || (p.peek() != '\'' && p.peek() != '"') {
				return nil, p.errorf("missing quoted name")
			}
			name, err := p.quotedName()
			if err != nil {
				return nil, err
			}
			names = append(names, name)
			p.skipSpace()
			if !p.eof() && p.peek() == ',' {
				p.pos++
				continue
			}
			break
		}
		if p.eof() || p.peek() != ']' {
			return nil, p.errorf(`missing "]" or ","`)
		}
		p.pos++
		if !p.eof() && p.peek() == '~' {
			return nil, p.errorf("the ~ property-name operator has no RFC 9535 equivalent")
		}
		return nameSeg{names: names}, nil
	case p.has("?("):
		p.pos += 2
		e, err := p.orExpr()
		if err != nil {
			return nil, err
		}
		p.skipSpace()
		if !p.has(")]") {
			return nil, p.errorf("missing end of filter")
		}
		p.pos += 2
		return filterSeg{expr: e, pos: open}, nil
	}
	end := strings.IndexByte(p.rest(), ']')
	if end < 0 {
		return nil, p.errorf("unmatched [")
	}
	raw := strings.TrimSpace(p.in[p.pos : p.pos+end])
	p.pos += end + 1
	if !p.eof() && p.peek() == '~' {
		return nil, p.errorf("the ~ property-name operator has no RFC 9535 equivalent")
	}
	if raw == "*" {
		return wildcardSeg{}, nil
	}
	single, err := validateSubscript(raw)
	if err != nil {
		return nil, &parseError{pos: open, msg: fmt.Sprintf("invalid array index [%s]: %v", raw, err)}
	}
	return subscriptSeg{raw: raw, single: single}, nil
}

// validateSubscript checks a union of indices and slices, and reports whether
// it is a single index.
func validateSubscript(raw string) (bool, error) {
	parts := strings.Split(raw, ",")
	for _, part := range parts {
		bounds := strings.Split(part, ":")
		if len(bounds) > 3 {
			return false, fmt.Errorf("too many colons")
		}
		for _, b := range bounds {
			b = strings.TrimSpace(b)
			if b == "" {
				if len(bounds) == 1 {
					return false, fmt.Errorf("index missing")
				}
				continue
			}
			if _, err := strconv.Atoi(b); err != nil {
				return false, fmt.Errorf("non-integer index %q", b)
			}
		}
	}
	return len(parts) == 1 && !strings.Contains(raw, ":"), nil
}

// quotedName parses a quoted name in brackets, where the previous dialect
// allowed only \<quote> and \\ as escapes.
func (p *parser) quotedName() (string, error) {
	quote := p.peek()
	p.pos++
	var b strings.Builder
	for {
		if p.eof() {
			return "", p.errorf("unmatched %c", quote)
		}
		c := p.peek()
		switch c {
		case quote:
			p.pos++
			return b.String(), nil
		case '\\':
			if p.pos+1 < len(p.in) && (p.in[p.pos+1] == quote || p.in[p.pos+1] == '\\') {
				b.WriteByte(p.in[p.pos+1])
				p.pos += 2
				continue
			}
			return "", p.errorf("unsupported escape sequence inside %c%c", quote, quote)
		default:
			b.WriteByte(c)
			p.pos++
		}
	}
}

// unescapeName removes backslashes from a dot-notation name the way the
// previous dialect did: \\ is a backslash, any other backslash is dropped.
func unescapeName(raw string) string {
	if !strings.Contains(raw, `\`) {
		return raw
	}
	var b strings.Builder
	escaped := false
	for _, r := range raw {
		if r == '\\' {
			if escaped {
				b.WriteRune(r)
			}
			escaped = !escaped
			continue
		}
		escaped = false
		b.WriteRune(r)
	}
	return b.String()
}

func (p *parser) orExpr() (expr, error) {
	first, err := p.andExpr()
	if err != nil {
		return nil, err
	}
	terms := []expr{first}
	for {
		p.skipSpace()
		if !p.has("||") {
			break
		}
		p.pos += 2
		next, err := p.andExpr()
		if err != nil {
			return nil, err
		}
		terms = append(terms, next)
	}
	if len(terms) == 1 {
		return first, nil
	}
	return orExpr{terms: terms}, nil
}

func (p *parser) andExpr() (expr, error) {
	first, err := p.basicExpr()
	if err != nil {
		return nil, err
	}
	terms := []expr{first}
	for {
		p.skipSpace()
		if !p.has("&&") {
			break
		}
		p.pos += 2
		next, err := p.basicExpr()
		if err != nil {
			return nil, err
		}
		terms = append(terms, next)
	}
	if len(terms) == 1 {
		return first, nil
	}
	return andExpr{terms: terms}, nil
}

var comparators = []string{"==", "!=", ">=", "<=", ">", "<"}

func (p *parser) basicExpr() (expr, error) {
	p.skipSpace()
	if p.eof() {
		return nil, p.errorf("missing end of filter")
	}
	switch {
	case p.has("!="):
		return nil, p.errorf("missing first operand for binary operator !=")
	case p.peek() == '!':
		p.pos++
		x, err := p.basicExpr()
		if err != nil {
			return nil, err
		}
		return notExpr{x: x}, nil
	case p.peek() == '(':
		p.pos++
		x, err := p.orExpr()
		if err != nil {
			return nil, err
		}
		p.skipSpace()
		if p.eof() || p.peek() != ')' {
			return nil, p.errorf("missing )")
		}
		p.pos++
		return parenExpr{x: x}, nil
	}
	lhs, err := p.term()
	if err != nil {
		return nil, err
	}
	p.skipSpace()
	if p.has("=~") {
		q, ok := lhs.(query)
		if !ok {
			return nil, p.errorf("literal cannot be matched using =~")
		}
		p.pos += 2
		p.skipSpace()
		re, err := p.regex()
		if err != nil {
			return nil, err
		}
		return matchExpr{q: q, re: re}, nil
	}
	for _, op := range comparators {
		if p.has(op) {
			p.pos += len(op)
			rhs, err := p.term()
			if err != nil {
				return nil, err
			}
			return cmpExpr{op: op, lhs: lhs, rhs: rhs}, nil
		}
	}
	q, ok := lhs.(query)
	if !ok {
		return nil, p.errorf("a literal on its own is not a filter")
	}
	return existExpr{q: q}, nil
}

func (p *parser) term() (term, error) {
	p.skipSpace()
	if p.eof() {
		return nil, p.errorf("missing filter term")
	}
	switch c := p.peek(); {
	case c == '@' || c == '$':
		p.pos++
		segs, err := p.path(true)
		if err != nil {
			return nil, err
		}
		return query{root: string(c), segs: segs}, nil
	case c == '\'' || c == '"':
		return p.stringLiteral()
	case c == '-' || c == '.' || (c >= '0' && c <= '9'):
		return p.numberLiteral()
	}
	for _, kw := range []string{"true", "false", "null"} {
		if p.has(kw) {
			p.pos += len(kw)
			return literal{text: kw}, nil
		}
	}
	return nil, p.errorf("invalid filter term")
}

// stringLiteral parses a filter string literal. The previous dialect had no
// escapes in these: everything up to the closing quote is the value.
func (p *parser) stringLiteral() (term, error) {
	quote := p.peek()
	start := p.pos
	end := strings.IndexByte(p.in[start+1:], quote)
	if end < 0 {
		return nil, p.errorf("unmatched string delimiter %c", quote)
	}
	value := p.in[start+1 : start+1+end]
	p.pos = start + end + 2
	return literal{text: quoteString(value)}, nil
}

func (p *parser) numberLiteral() (term, error) {
	start := p.pos
	p.pos++
	isFloat := p.in[start] == '.'
	for !p.eof() {
		c := p.peek()
		if c == '.' || c == 'e' || c == 'E' || c == '-' || c == '+' {
			isFloat = true
		} else if c < '0' || c > '9' {
			break
		}
		p.pos++
	}
	text := p.in[start:p.pos]
	if isFloat {
		f, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return nil, &parseError{pos: start, msg: fmt.Sprintf("invalid float literal %q", text)}
		}
		return literal{text: strconv.FormatFloat(f, 'g', -1, 64)}, nil
	}
	n, err := strconv.Atoi(text)
	if err != nil {
		return nil, &parseError{pos: start, msg: fmt.Sprintf("invalid integer literal %q", text)}
	}
	return literal{text: strconv.Itoa(n)}, nil
}

// regex parses /re/, where \/ stands for a slash.
func (p *parser) regex() (string, error) {
	if p.eof() || p.peek() != '/' {
		return "", p.errorf("regular expression does not start with /")
	}
	start := p.pos
	p.pos++
	escaped := false
	for {
		if p.eof() {
			return "", &parseError{pos: start, msg: "unmatched regular expression delimiter /"}
		}
		c := p.peek()
		if c == '/' && !escaped {
			break
		}
		escaped = c == '\\' && !escaped
		p.pos++
	}
	re := strings.ReplaceAll(p.in[start+1:p.pos], `\/`, `/`)
	p.pos++ // closing /
	if !p.eof() && isLetter(p.peek()) {
		return "", p.errorf("regular expression flags are not supported; use an inline group such as (?i) instead")
	}
	if _, err := regexp.Compile(re); err != nil {
		return "", &parseError{pos: start, msg: fmt.Sprintf("invalid regular expression: %v", err)}
	}
	return re, nil
}

func isLetter(c byte) bool { return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') }

// quoteString formats s as an RFC 9535 single-quoted string literal.
func quoteString(s string) string {
	var b strings.Builder
	b.WriteByte('\'')
	for _, r := range s {
		switch {
		case r == '\'':
			b.WriteString(`\'`)
		case r == '\\':
			b.WriteString(`\\`)
		case r == '\b':
			b.WriteString(`\b`)
		case r == '\f':
			b.WriteString(`\f`)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\r':
			b.WriteString(`\r`)
		case r == '\t':
			b.WriteString(`\t`)
		case r < 0x20:
			fmt.Fprintf(&b, `\u%04x`, r)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('\'')
	return b.String()
}

var shorthandName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func emitName(name string, dotted bool) string {
	if dotted && shorthandName.MatchString(name) {
		return "." + name
	}
	return "[" + quoteString(name) + "]"
}

func emitSegment(seg segment) string {
	switch s := seg.(type) {
	case nameSeg:
		if len(s.names) == 1 {
			return emitName(s.names[0], s.dotted)
		}
		quoted := make([]string, len(s.names))
		for i, n := range s.names {
			quoted[i] = quoteString(n)
		}
		return "[" + strings.Join(quoted, ",") + "]"
	case wildcardSeg:
		if s.dotted {
			return ".*"
		}
		return "[*]"
	case subscriptSeg:
		return "[" + s.raw + "]"
	}
	panic(fmt.Sprintf("unexpected segment %T", seg))
}

// emitPath emits root followed by segs in RFC 9535 syntax.
func emitPath(root string, segs []segment) (string, error) {
	parts := make([]string, 0, len(segs))
	for i, seg := range segs {
		switch s := seg.(type) {
		case filterSeg:
			cond, err := emitExpr(s.expr)
			if err != nil {
				return "", err
			}
			var prev segment
			if i > 0 {
				prev = segs[i-1]
			}
			switch pv := prev.(type) {
			case wildcardSeg:
				// The previous dialect tested each child the wildcard
				// selected; an RFC 9535 filter at the wildcard's place does
				// the same.
				parts[len(parts)-1] = "[?" + cond + "]"
			case descendantSeg:
				if _, ok := pv.inner.(wildcardSeg); !ok {
					return "", filterOnNamedNode(root, segs[:i], s.pos)
				}
				parts[len(parts)-1] = "..[?" + cond + "]"
			default:
				return "", filterOnNamedNode(root, segs[:i], s.pos)
			}
		case descendantSeg:
			if f, ok := s.inner.(filterSeg); ok {
				cond, err := emitExpr(f.expr)
				if err != nil {
					return "", err
				}
				parts = append(parts, "..[?"+cond+"]")
				continue
			}
			inner := emitSegment(s.inner)
			if strings.HasPrefix(inner, ".") {
				parts = append(parts, "."+inner)
			} else {
				parts = append(parts, ".."+inner)
			}
		default:
			parts = append(parts, emitSegment(seg))
		}
	}
	return root + strings.Join(parts, ""), nil
}

// filterOnNamedNode is the error for a filter that follows something other
// than a wildcard or recursive descent.
func filterOnNamedNode(root string, before []segment, pos int) error {
	node := root
	if len(before) > 0 {
		if s, err := emitPath(root, before); err == nil {
			node = s
		}
	}
	return fmt.Errorf("the filter at position %d applies to %s itself: the previous dialect tested a mapping by its own content, and RFC 9535 cannot test a node by its own key; "+
		"if %s is a sequence, write [*] before the filter, which both dialects read as filtering its elements, "+
		"otherwise move the condition to its parent, as in $.parent[?@.name.field] instead of $.parent.name[?(@.field)]", pos, node, node)
}

func emitExpr(e expr) (string, error) {
	switch x := e.(type) {
	case orExpr:
		return emitJoined(x.terms, " || ")
	case andExpr:
		return emitJoined(x.terms, " && ")
	case parenExpr:
		inner, err := emitExpr(x.x)
		if err != nil {
			return "", err
		}
		return "(" + inner + ")", nil
	case notExpr:
		inner, err := emitExpr(x.x)
		if err != nil {
			return "", err
		}
		// RFC 9535 only negates a test expression or a parenthesised
		// expression.
		switch x.x.(type) {
		case parenExpr, existExpr, matchExpr:
			if !strings.HasPrefix(inner, "!") {
				return "!" + inner, nil
			}
		}
		return "!(" + inner + ")", nil
	case existExpr:
		return emitQuery(x.q)
	case cmpExpr:
		return emitComparison(x)
	case matchExpr:
		return emitMatch(x)
	}
	panic(fmt.Sprintf("unexpected expression %T", e))
}

func emitJoined(terms []expr, sep string) (string, error) {
	out := make([]string, len(terms))
	for i, t := range terms {
		s, err := emitExpr(t)
		if err != nil {
			return "", err
		}
		out[i] = s
	}
	return strings.Join(out, sep), nil
}

func emitQuery(q query) (string, error) {
	return emitPath(q.root, q.segs)
}

func emitTerm(t term) (string, error) {
	switch v := t.(type) {
	case literal:
		return v.text, nil
	case query:
		return emitQuery(v)
	}
	panic(fmt.Sprintf("unexpected term %T", t))
}

// singular reports whether segs select at most one node, as RFC 9535
// requires of queries in comparisons and function arguments.
func singular(segs []segment) bool {
	for _, seg := range segs {
		switch s := seg.(type) {
		case nameSeg:
			if len(s.names) != 1 {
				return false
			}
		case subscriptSeg:
			if !s.single {
				return false
			}
		default:
			return false
		}
	}
	return true
}

// splitAtWildcard splits a query that is singular except for one wildcard
// into the query up to the wildcard and the rest after it.
func splitAtWildcard(q query) (prefix, suffix query, ok bool) {
	for i, seg := range q.segs {
		if _, isWildcard := seg.(wildcardSeg); isWildcard {
			if singular(q.segs[:i]) && singular(q.segs[i+1:]) {
				return query{root: q.root, segs: q.segs[:i]}, query{root: "@", segs: q.segs[i+1:]}, true
			}
			return query{}, query{}, false
		}
	}
	return query{}, query{}, false
}

// emitComparison emits a comparison. A query that selects several nodes
// through a wildcard becomes a nested filter over them: X[*] == v becomes
// X[?@ == v].
func emitComparison(c cmpExpr) (string, error) {
	lq, lIsQuery := c.lhs.(query)
	rq, rIsQuery := c.rhs.(query)
	lMulti := lIsQuery && !singular(lq.segs)
	rMulti := rIsQuery && !singular(rq.segs)

	if !lMulti && !rMulti {
		l, err := emitTerm(c.lhs)
		if err != nil {
			return "", err
		}
		r, err := emitTerm(c.rhs)
		if err != nil {
			return "", err
		}
		return l + " " + c.op + " " + r, nil
	}
	if lMulti && rMulti {
		return "", fmt.Errorf("cannot compare two queries that each select several nodes")
	}
	multi, other := lq, c.rhs
	if rMulti {
		multi, other = rq, c.lhs
	}
	if oq, ok := other.(query); ok && oq.root == "@" {
		return "", fmt.Errorf("cannot compare a query that selects several nodes with a query relative to @")
	}
	prefix, suffix, ok := splitAtWildcard(multi)
	if !ok {
		return "", fmt.Errorf("a comparison can only use a query that selects several nodes through a single wildcard, such as @.tags[*] == 'x'")
	}
	p, err := emitQuery(prefix)
	if err != nil {
		return "", err
	}
	s, err := emitQuery(suffix)
	if err != nil {
		return "", err
	}
	o, err := emitTerm(other)
	if err != nil {
		return "", err
	}
	if rMulti {
		return p + "[?" + o + " " + c.op + " " + s + "]", nil
	}
	return p + "[?" + s + " " + c.op + " " + o + "]", nil
}

// emitMatch emits X =~ /re/ as search(X, 're'), nesting it in a filter when X
// selects several nodes through a wildcard.
func emitMatch(m matchExpr) (string, error) {
	re := quoteString(m.re)
	if singular(m.q.segs) {
		q, err := emitQuery(m.q)
		if err != nil {
			return "", err
		}
		return "search(" + q + ", " + re + ")", nil
	}
	prefix, suffix, ok := splitAtWildcard(m.q)
	if !ok {
		return "", fmt.Errorf("=~ can only use a query that selects several nodes through a single wildcard, such as @.tags[*] =~ /x/")
	}
	p, err := emitQuery(prefix)
	if err != nil {
		return "", err
	}
	s, err := emitQuery(suffix)
	if err != nil {
		return "", err
	}
	return p + "[?search(" + s + ", " + re + ")]", nil
}
