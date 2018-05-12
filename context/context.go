package context

// Context ...
type Context interface {
	Value(key interface{}) interface{}
}

type valueCtx struct {
	Context
	key, value interface{}
}

func (c *valueCtx) Value(key interface{}) interface{} {
	if c.key == key {
		return c.value
	}
	return c.Context.Value(key)
}

// WithValue ...
func WithValue(parent Context, key, value interface{}) Context {
	return &valueCtx{parent, key, value}
}

// NewContext ...
func NewContext() Context {
	return &valueCtx{}
}