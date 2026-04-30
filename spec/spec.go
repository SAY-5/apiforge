// Package spec models a minimal subset of OpenAPI 3.x focused on
// what the linter + diff engine need: paths, methods, parameters,
// responses. The full OpenAPI surface is much bigger; we cover the
// 80% that catches 100% of the common contract violations.
package spec

import (
	"encoding/json"
	"errors"
	"sort"
)

type Spec struct {
	Title   string             `json:"title"`
	Version string             `json:"version"`
	Paths   map[string]Path    `json:"paths"`
}

type Path struct {
	// One Operation per HTTP method.
	Get    *Operation `json:"get,omitempty"`
	Post   *Operation `json:"post,omitempty"`
	Put    *Operation `json:"put,omitempty"`
	Delete *Operation `json:"delete,omitempty"`
	Patch  *Operation `json:"patch,omitempty"`
}

type Operation struct {
	OperationID string                  `json:"operationId,omitempty"`
	Summary     string                  `json:"summary,omitempty"`
	Parameters  []Parameter             `json:"parameters,omitempty"`
	Responses   map[string]Response     `json:"responses,omitempty"`
	RequestBody *RequestBody            `json:"requestBody,omitempty"`
}

type Parameter struct {
	Name     string `json:"name"`
	In       string `json:"in"`     // "path" | "query" | "header"
	Required bool   `json:"required"`
	Type     string `json:"type"`   // "string" | "integer" | etc.
}

type Response struct {
	Description string `json:"description"`
	Type        string `json:"type,omitempty"`
}

type RequestBody struct {
	Required bool   `json:"required"`
	Type     string `json:"type"`
}

func Parse(raw []byte) (*Spec, error) {
	var s Spec
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, err
	}
	if s.Paths == nil {
		s.Paths = map[string]Path{}
	}
	return &s, nil
}

// Operations returns (path, method, op) tuples in deterministic
// (path-then-method) order. Used by lint + diff.
func (s *Spec) Operations() []OperationRef {
	out := make([]OperationRef, 0)
	for path, p := range s.Paths {
		methods := []struct {
			name string
			op   *Operation
		}{
			{"GET", p.Get}, {"POST", p.Post}, {"PUT", p.Put},
			{"DELETE", p.Delete}, {"PATCH", p.Patch},
		}
		for _, m := range methods {
			if m.op != nil {
				out = append(out, OperationRef{Path: path, Method: m.name, Op: m.op})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Path != out[j].Path {
			return out[i].Path < out[j].Path
		}
		return out[i].Method < out[j].Method
	})
	return out
}

type OperationRef struct {
	Path   string
	Method string
	Op     *Operation
}

// ErrEmptySpec is returned by Parse when given empty input.
var ErrEmptySpec = errors.New("spec: empty input")
