package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/database"
)

// gqlRequest is the wire format for GraphQL-over-HTTP.
type gqlRequest struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
}

// gqlResponse is the GraphQL JSON response envelope.
type gqlResponse struct {
	Data   interface{}  `json:"data,omitempty"`
	Errors []gqlError   `json:"errors,omitempty"`
}

// gqlError follows the GraphQL spec error shape.
type gqlError struct {
	Message string `json:"message"`
}

// Handler is an http.Handler for the GraphQL endpoint.
type Handler struct {
	resolver *Resolver
}

// NewHandler constructs a GraphQL http.Handler backed by store.
func NewHandler(store database.Store) http.Handler {
	return &Handler{resolver: NewResolver(store)}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req gqlRequest

	switch r.Method {
	case http.MethodGet:
		req.Query = r.URL.Query().Get("query")
		req.OperationName = r.URL.Query().Get("operationName")
	case http.MethodPost:
		ct := r.Header.Get("Content-Type")
		if !strings.HasPrefix(ct, "application/json") {
			writeGQLError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json for POST")
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeGQLError(w, http.StatusBadRequest, "invalid JSON request body")
			return
		}
	default:
		writeGQLError(w, http.StatusMethodNotAllowed, "only GET and POST are supported")
		return
	}

	req.Query = strings.TrimSpace(req.Query)
	if req.Query == "" {
		writeGQLError(w, http.StatusBadRequest, "query must not be empty")
		return
	}

	data, errs := h.execute(r.Context(), req)
	resp := gqlResponse{Data: data}
	for _, e := range errs {
		resp.Errors = append(resp.Errors, gqlError{Message: e.Error()})
	}

	status := http.StatusOK
	if len(resp.Errors) > 0 && data == nil {
		status = http.StatusBadRequest
	}
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

// ─── minimal query executor ───────────────────────────────────────────────────

// execute parses a minimal GraphQL query and dispatches to the resolver.
// Supports: introspection (__schema, __type), and all root Query fields.
func (h *Handler) execute(ctx context.Context, req gqlRequest) (interface{}, []error) {
	q := req.Query

	// Introspection: __schema
	if isIntrospectionSchema(q) {
		return map[string]interface{}{
			"__schema": map[string]interface{}{
				"sdl": Schema,
			},
		}, nil
	}

	// Introspection: __typename on Query
	if isTypenameQuery(q) {
		return map[string]interface{}{"__typename": "Query"}, nil
	}

	// Parse the selection set from the query.
	fields, args, err := parseQuery(q)
	if err != nil {
		return nil, []error{fmt.Errorf("syntax error: %w", err)}
	}

	result := make(map[string]interface{})
	var errs []error

	for _, field := range fields {
		fieldArgs := mergeArgs(args[field], req.Variables)
		value, err := h.resolveField(ctx, field, fieldArgs)
		if err != nil {
			errs = append(errs, fmt.Errorf("field %s: %w", field, err))
			result[field] = nil
		} else {
			result[field] = value
		}
	}

	if len(result) == 0 && len(errs) > 0 {
		return nil, errs
	}
	return result, errs
}

func (h *Handler) resolveField(ctx context.Context, field string, args map[string]interface{}) (interface{}, error) {
	switch field {
	case "projects":
		return h.resolver.resolveProjects(ctx, args)
	case "project":
		return h.resolver.resolveProject(ctx, args)
	case "organizations":
		return h.resolver.resolveOrganizations(ctx, args)
	case "events":
		return h.resolver.resolveEvents(ctx, args)
	case "signals":
		return h.resolver.resolveSignals(ctx, args)
	case "reconciliation":
		return h.resolver.resolveReconciliation(ctx, args)
	case "aiSovereignty":
		return h.resolver.resolveAISovereignty(ctx, args)
	default:
		return nil, fmt.Errorf("unknown field %q", field)
	}
}

// ─── query parser ─────────────────────────────────────────────────────────────

// parseQuery extracts top-level field names and their inline arguments from a
// simple GraphQL query. It does not handle fragments, directives, or nested
// selections — those pass through opaquely and field data is returned flat.
//
// Supported syntax examples:
//
//	{ projects { id name } }
//	{ project(id: "abc") { id name stage } }
//	query { projects(sector: "LNG", limit: 10) { id } }
func parseQuery(q string) (fields []string, args map[string]map[string]interface{}, err error) {
	args = make(map[string]map[string]interface{})

	// Strip outer `query` keyword and whitespace.
	q = strings.TrimSpace(q)
	q = strings.TrimPrefix(q, "query")
	q = strings.TrimPrefix(q, "Query")
	q = strings.TrimSpace(q)

	// Strip the outer braces.
	q, err = stripOuterBraces(q)
	if err != nil {
		return nil, nil, err
	}
	q = strings.TrimSpace(q)

	// Scan tokens: field names are bare identifiers; parentheses hold args.
	pos := 0
	for pos < len(q) {
		// Skip whitespace and commas.
		for pos < len(q) && isWS(q[pos]) {
			pos++
		}
		if pos >= len(q) {
			break
		}

		// Read field name.
		start := pos
		for pos < len(q) && isIdent(q[pos]) {
			pos++
		}
		if pos == start {
			// Not an identifier — could be a sub-selection brace or alias colon.
			// Skip until next identifier-start character.
			depth := 0
			for pos < len(q) {
				if q[pos] == '{' {
					depth++
				} else if q[pos] == '}' {
					if depth == 0 {
						break
					}
					depth--
				} else if depth == 0 && isIdent(q[pos]) {
					break
				}
				pos++
			}
			continue
		}

		fieldName := q[start:pos]

		// Skip whitespace.
		for pos < len(q) && isWS(q[pos]) {
			pos++
		}

		// Alias colon: "alias: fieldName ..." — advance past the real field name.
		if pos < len(q) && q[pos] == ':' {
			pos++ // skip ':'
			for pos < len(q) && isWS(q[pos]) {
				pos++
			}
			start2 := pos
			for pos < len(q) && isIdent(q[pos]) {
				pos++
			}
			if pos > start2 {
				fieldName = q[start2:pos]
			}
			for pos < len(q) && isWS(q[pos]) {
				pos++
			}
		}

		// Parse arguments if present.
		var fieldArgs map[string]interface{}
		if pos < len(q) && q[pos] == '(' {
			end, a, parseErr := parseArgs(q, pos)
			if parseErr != nil {
				return nil, nil, parseErr
			}
			fieldArgs = a
			pos = end
		}

		// Skip sub-selection block.
		for pos < len(q) && isWS(q[pos]) {
			pos++
		}
		if pos < len(q) && q[pos] == '{' {
			depth := 1
			pos++
			for pos < len(q) && depth > 0 {
				if q[pos] == '{' {
					depth++
				} else if q[pos] == '}' {
					depth--
				}
				pos++
			}
		}

		fields = append(fields, fieldName)
		if fieldArgs != nil {
			args[fieldName] = fieldArgs
		}
	}

	if len(fields) == 0 {
		return nil, nil, fmt.Errorf("no fields found in query")
	}
	return fields, args, nil
}

// parseArgs parses a GraphQL argument list "(key: value, ...)" starting at
// openParen. Returns the index after the closing ')' and the parsed args map.
func parseArgs(q string, openParen int) (int, map[string]interface{}, error) {
	args := make(map[string]interface{})
	pos := openParen + 1 // skip '('

	for pos < len(q) {
		// Skip whitespace and commas.
		for pos < len(q) && (isWS(q[pos]) || q[pos] == ',') {
			pos++
		}
		if pos >= len(q) {
			return pos, args, fmt.Errorf("unclosed argument list")
		}
		if q[pos] == ')' {
			return pos + 1, args, nil
		}

		// Read key.
		start := pos
		for pos < len(q) && isIdent(q[pos]) {
			pos++
		}
		if pos == start {
			return pos, args, fmt.Errorf("expected argument name at pos %d", pos)
		}
		key := q[start:pos]

		// Skip whitespace + colon.
		for pos < len(q) && isWS(q[pos]) {
			pos++
		}
		if pos >= len(q) || q[pos] != ':' {
			return pos, args, fmt.Errorf("expected ':' after argument name %q", key)
		}
		pos++ // skip ':'
		for pos < len(q) && isWS(q[pos]) {
			pos++
		}

		// Read value.
		val, end, err := parseValue(q, pos)
		if err != nil {
			return end, args, err
		}
		args[key] = val
		pos = end
	}
	return pos, args, fmt.Errorf("unclosed argument list")
}

// parseValue reads a GraphQL literal value (string, int, float, boolean, null,
// variable, or enum). Returns parsed Go value and the position after the value.
func parseValue(q string, pos int) (interface{}, int, error) {
	if pos >= len(q) {
		return nil, pos, fmt.Errorf("expected value at end of input")
	}

	// String: "..."
	if q[pos] == '"' {
		pos++ // skip opening quote
		var sb strings.Builder
		for pos < len(q) {
			ch := q[pos]
			if ch == '\\' && pos+1 < len(q) {
				pos++
				switch q[pos] {
				case '"':
					sb.WriteByte('"')
				case '\\':
					sb.WriteByte('\\')
				case 'n':
					sb.WriteByte('\n')
				case 'r':
					sb.WriteByte('\r')
				case 't':
					sb.WriteByte('\t')
				default:
					sb.WriteByte('\\')
					sb.WriteByte(q[pos])
				}
			} else if ch == '"' {
				pos++ // skip closing quote
				return sb.String(), pos, nil
			} else {
				sb.WriteByte(ch)
			}
			pos++
		}
		return nil, pos, fmt.Errorf("unterminated string")
	}

	// Number (integer or float): [-]digit...
	if q[pos] == '-' || (q[pos] >= '0' && q[pos] <= '9') {
		start := pos
		if q[pos] == '-' {
			pos++
		}
		for pos < len(q) && ((q[pos] >= '0' && q[pos] <= '9') || q[pos] == '.' || q[pos] == 'e' || q[pos] == 'E' || q[pos] == '+' || q[pos] == '-') {
			pos++
		}
		tok := q[start:pos]
		// Try integer first.
		var iv int64
		if _, err := fmt.Sscan(tok, &iv); err == nil && !strings.ContainsAny(tok, ".eE") {
			return int(iv), pos, nil
		}
		var fv float64
		if _, err := fmt.Sscan(tok, &fv); err == nil {
			return fv, pos, nil
		}
		return nil, pos, fmt.Errorf("invalid number %q", tok)
	}

	// Boolean / null / enum identifier.
	if isIdentStart(q[pos]) {
		start := pos
		for pos < len(q) && isIdent(q[pos]) {
			pos++
		}
		tok := q[start:pos]
		switch tok {
		case "true":
			return true, pos, nil
		case "false":
			return false, pos, nil
		case "null":
			return nil, pos, nil
		default:
			// Enum or variable reference treated as string.
			return tok, pos, nil
		}
	}

	// Variable: $varName
	if q[pos] == '$' {
		pos++
		start := pos
		for pos < len(q) && isIdent(q[pos]) {
			pos++
		}
		return "$" + q[start:pos], pos, nil
	}

	return nil, pos, fmt.Errorf("unexpected character %q at pos %d", q[pos], pos)
}

// ─── introspection helpers ────────────────────────────────────────────────────

func isIntrospectionSchema(q string) bool {
	q = strings.ToLower(strings.TrimSpace(q))
	q = strings.TrimPrefix(q, "query")
	q = strings.TrimSpace(q)
	return strings.Contains(q, "__schema")
}

func isTypenameQuery(q string) bool {
	q = strings.ToLower(strings.TrimSpace(q))
	return strings.Contains(q, "__typename")
}

// ─── utility ─────────────────────────────────────────────────────────────────

func stripOuterBraces(q string) (string, error) {
	q = strings.TrimSpace(q)
	if len(q) == 0 || q[0] != '{' {
		return "", fmt.Errorf("query must begin with '{'")
	}
	// Find matching closing brace.
	depth := 0
	for i, ch := range q {
		if ch == '{' {
			depth++
		} else if ch == '}' {
			depth--
			if depth == 0 {
				return q[1:i], nil
			}
		}
	}
	return "", fmt.Errorf("unmatched '{' in query")
}

func isWS(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == ','
}

func isIdentStart(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_'
}

func isIdent(b byte) bool {
	return isIdentStart(b) || (b >= '0' && b <= '9')
}

func mergeArgs(fieldArgs map[string]interface{}, variables map[string]interface{}) map[string]interface{} {
	if fieldArgs == nil && variables == nil {
		return nil
	}
	out := make(map[string]interface{})
	for k, v := range variables {
		out[k] = v
	}
	for k, v := range fieldArgs {
		// Resolve variable references.
		if s, ok := v.(string); ok && strings.HasPrefix(s, "$") {
			varName := strings.TrimPrefix(s, "$")
			if resolved, exists := variables[varName]; exists {
				out[k] = resolved
				continue
			}
		}
		out[k] = v
	}
	return out
}

func writeGQLError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(gqlResponse{
		Errors: []gqlError{{Message: msg}},
	})
}
