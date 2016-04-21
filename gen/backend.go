package gen

import "github.com/codingbrain/clix.go/args"

// Backend is code generator
type Backend interface {
	// GenerateCode emits code to specified Writer
	// Backend is able to reconfigure the indent settings of the writer
	GenerateCode(*args.CliDef, *Writer) error
}

// BackendParams is parameters for creating a backend
type BackendParams map[string]interface{}

// BackendFactory creates a backend
type BackendFactory func(BackendParams) (Backend, error)

// BackendFactories are registered named backend factories
var BackendFactories = make(map[string]BackendFactory)

// CreateBackend creates a backend using registered factory
func CreateBackend(name string, params BackendParams) (Backend, error) {
	if f, ok := BackendFactories[name]; ok {
		return f(params)
	}
	return nil, nil
}
