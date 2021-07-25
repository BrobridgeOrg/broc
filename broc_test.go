package broc

import (
	"bytes"
	"testing"
)

var testBroc *Broc

func TestBrocInit(t *testing.T) {
	testBroc = NewBroc(nil)

	if len(testBroc.handlers) != 1 {
		t.Fail()
	}
}

func TestPrefix(t *testing.T) {

	if len(testBroc.prefix) != 0 {
		t.Fail()
	}

	testBroc.SetPrefix("test.")

	if testBroc.prefix != "test." {
		t.Log(testBroc.prefix)
		t.Fail()
	}
}

func TestUse(t *testing.T) {

	count := len(testBroc.handlers)

	testBroc.Use(func(ctx *Context) (interface{}, error) {
		return ctx.Next()
	})

	if len(testBroc.handlers) != count+1 {
		t.Fail()
	}

}

func TestHandlerStack(t *testing.T) {

	method := "method1"

	testBroc.Register(method, func(ctx *Context) (interface{}, error) {
		return []byte("result"), nil
	})

	ctx := testBroc.prepareContext(method, nil)

	if len(ctx.handlers) != len(testBroc.handlers)+1 {
		t.Fail()
	}

	data, err := ctx.handlers[0](ctx)
	if err != nil {
		t.Error(err)
	}

	if data == nil {
		t.Errorf("result is nil")
	}

	if bytes.Compare(data.([]byte), []byte("result")) != 0 {
		t.Fail()
	}
}
