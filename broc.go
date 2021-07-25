package broc

import (
	"fmt"

	"github.com/nats-io/nats.go"
)

type Handler func(*Context) (interface{}, error)

type Broc struct {
	conn           *nats.Conn
	prefix         string
	handlers       []Handler
	methodHandlers map[string][]Handler
}

func NewBroc(conn *nats.Conn) *Broc {

	b := &Broc{
		conn:           conn,
		prefix:         "",
		handlers:       make([]Handler, 0),
		methodHandlers: make(map[string][]Handler),
	}

	b.Use(b.rootHandler)

	return b
}

func (broc *Broc) rootHandler(ctx *Context) (interface{}, error) {

	fmt.Printf("  -> %s%s\n", broc.prefix, ctx.Get("method").(string))

	// Trying to run next handler
	resp, err := ctx.Next()
	if err != nil {
		fmt.Printf("  <- %s%s - Error: %v\n", broc.prefix, ctx.Get("method").(string), err)
		return nil, err
	}

	return resp, nil
}

func (broc *Broc) prepareContext(method string, m *nats.Msg) *Context {

	handlers, _ := broc.methodHandlers[method]
	ctx := NewContext(broc)
	ctx.msg = m
	ctx.handlers = append(broc.handlers, handlers...)
	ctx.meta["method"] = method

	if m != nil {
		ctx.meta["request"] = m.Data
	}

	return ctx
}

func (broc *Broc) handler(method string, m *nats.Msg) {

	ctx := broc.prepareContext(method, m)

	data, err := ctx.handlers[0](ctx)
	if err != nil {
		m.Nak()
		return
	}

	if data == nil {
		m.RespondMsg(m)
		return
	}

	m.Respond(data.([]byte))
}

func (broc *Broc) register(method string) error {

	channel := fmt.Sprintf("%s%s", broc.prefix, method)

	fmt.Printf("Registering %s\n", channel)

	_, err := broc.conn.Subscribe(channel, func(m *nats.Msg) {
		broc.handler(method, m)
	})
	if err != nil {
		return err
	}

	return nil
}

func (broc *Broc) SetPrefix(prefix string) {
	broc.prefix = prefix
}

func (broc *Broc) Use(handler Handler) {
	broc.handlers = append(broc.handlers, handler)
}

func (broc *Broc) Register(method string, handlers ...Handler) {
	broc.methodHandlers[method] = handlers
}

func (broc *Broc) Apply() error {

	for method, _ := range broc.methodHandlers {

		err := broc.register(method)
		if err != nil {
			return err
		}
	}

	return nil
}
