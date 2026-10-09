// Package perky provides independently owned, typed context keys with defaults
// and immutable overrides, either standalone or attached to a Go context.
package perky

import "context"

// Values is the lookup shared by Context and context.Context.
type Values interface {
	Value(any) any
}

type source[T any] struct {
	value  T
	getter func() T
}

// Each key has its own reference identity and a typed default.
type key[T any] struct {
	fallback source[T]
}

// KeyRef names a typed key reference for fields, parameters, and return values.
type KeyRef[T any] = *key[T]

// Key declares a distinct key with a fixed default, including function values.
func Key[T any](value T) KeyRef[T] {
	return &key[T]{fallback: source[T]{value: value}}
}

// LiveKey declares a distinct key whose default getter runs on every read.
func LiveKey[T any](getter func() T) KeyRef[T] {
	if getter == nil {
		panic("perky: nil getter")
	}
	return &key[T]{fallback: source[T]{getter: getter}}
}

// Binding associates a key with a fixed value or live getter.
// Its contents are private and cannot be changed through the public API.
type Binding struct {
	key    any
	source any
}

// Bind creates a fixed override without changing any context.
func (k *key[T]) Bind(value T) Binding {
	return Binding{key: k, source: source[T]{value: value}}
}

// LiveBind creates an override whose getter runs on every read.
func (k *key[T]) LiveBind(getter func() T) Binding {
	if getter == nil {
		panic("perky: nil getter")
	}
	return Binding{key: k, source: source[T]{getter: getter}}
}

// Get reads a typed value from either a Perky context or a Go context.
// When no binding exists, the key's default applies.
func (k *key[T]) Get(ctx Values) T {
	f, _ := ctx.Value(frameKey{}).(*frame)
	s := k.fallback
	for ; f != nil; f = f.parent {
		if value, found := f.values[k]; found {
			// Bind and LiveBind guarantee the source has this key's type.
			s = value.(source[T])
			break
		}
	}
	if s.getter != nil {
		return s.getter()
	}
	return s.value
}

// A single private Go-context key carries the entire Perky frame chain.
type frameKey struct{}

type frame struct {
	parent *frame
	values map[any]any
}

// Context carries only Perky values. Its zero value is an empty context.
// It has no cancellation, deadlines, or unrelated Go context values.
type Context struct {
	frame *frame
}

// New creates an independent context with optional initial bindings.
func New(bindings ...Binding) Context {
	return Context{}.With(bindings...)
}

// With derives a child context. Later bindings for the same key win.
func (c Context) With(bindings ...Binding) Context {
	if len(bindings) == 0 {
		return c
	}
	f := &frame{parent: c.frame, values: make(map[any]any, len(bindings))}
	for _, b := range bindings {
		f.values[b.key] = b.source
	}
	// The map stays private and is never mutated after construction.
	return Context{frame: f}
}

// Value supports typed key lookup by exposing only Perky's private frame key.
// All other keys return nil; this is not a general-purpose Go value store.
func (c Context) Value(key any) any {
	if _, ok := key.(frameKey); ok && c.frame != nil {
		return c.frame
	}
	return nil
}

// Attach derives Perky bindings on a Go context, preserving its cancellation,
// deadlines, and other values.
func Attach(parent context.Context, bindings ...Binding) context.Context {
	return context.WithValue(parent, frameKey{}, Detach(parent).With(bindings...).frame)
}

// Detach extracts Perky values without retaining the native Go context chain.
// It neither copies bindings nor evaluates getters.
func Detach(ctx context.Context) Context {
	f, _ := ctx.Value(frameKey{}).(*frame)
	return Context{frame: f}
}
