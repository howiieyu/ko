package ko

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Context struct {
	Writer http.ResponseWriter
	Req *http.Request

	// request info
	Path string
	Method string
	Params map[string]string

	engine *Engine
	// response info
	StatusCode int

	// middlewares
	handlers []HandlerFunc
	index int
}


func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req: req,

		Path: req.URL.Path,
		Method: req.Method,
		Params: make(map[string]string),

		index: -1,
	}
}

func (c *Context) Next() {
	c.index++
	
	s := len(c.handlers)
	
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)

	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) JSON(code int, data interface{}) {
	c.Status(code)
	c.SetHeader("Content-Type", "application/json")
	encoder := json.NewEncoder(c.Writer)

	if err := encoder.Encode(data); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.JSON(code, H{"message": err})
}

func (c *Context) HTML(code int, name string, data any) {
	c.Status(code)
	c.SetHeader("Content-Type", "text/html")
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Fail(500, err.Error())
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

func (c *Context) PostFrom(key string) string {
	return c.Req.PostFormValue(key)
}

func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

type H map[string]interface{}
