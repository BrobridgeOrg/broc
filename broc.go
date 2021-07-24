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
	ctx.meta["request"] = ctx.msg.Data
	return ctx.Next()
}

func (broc *Broc) handler(method string, m *nats.Msg) {

	handlers, _ := broc.methodHandlers[method]

	ctx := NewContext()
	ctx.msg = m
	ctx.handlers = append(broc.handlers, handlers...)
	ctx.meta["method"] = method

	data, err := handlers[0](ctx)
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

		channel := fmt.Sprintf("%s%s", broc.prefix, method)

		_, err := broc.conn.Subscribe(channel, func(m *nats.Msg) {
			broc.handler(channel, m)
		})
		if err != nil {
			return err
		}
	}

	return nil
}
