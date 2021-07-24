package broc

import "github.com/nats-io/nats.go"

type Context struct {
	msg      *nats.Msg
	handlers []Handler
	meta     map[string]interface{}
	current  int
}

func NewContext() *Context {
	return &Context{
		meta:    make(map[string]interface{}),
		current: 0,
	}
}

func (ctx *Context) Set(propName string, value interface{}) {
	ctx.meta[propName] = value
}

func (ctx *Context) Get(propName string) interface{} {
	value, _ := ctx.meta[propName]
	return value
}

func (ctx *Context) GetMeta() map[string]interface{} {
	return ctx.meta
}

func (ctx *Context) Next() (interface{}, error) {
	ctx.current++
	return ctx.handlers[ctx.current](ctx)
}
