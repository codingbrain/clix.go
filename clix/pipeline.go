// Common models used by utilities

package clix

type Receptor interface {
	Consume(item interface{}) error
	End() error
}

type Emitter interface {
	EmitTo(receptor Receptor)
}
