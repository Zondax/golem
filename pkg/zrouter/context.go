package zrouter

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type Context interface {
	Request() *http.Request
	BindJSON(obj interface{}) error
	JSON(code int, obj interface{}) error
	Data(code int, contentType string, data []byte)
	Header(key, value string)
	Param(key string) string
	Query(key string) string
	DefaultQuery(key, defaultValue string) string
}

type chiContextAdapter struct {
	ctx http.ResponseWriter
	req *http.Request
}

func (c *chiContextAdapter) Request() *http.Request {
	return c.req
}

func (c *chiContextAdapter) JSON(code int, obj interface{}) error {
	c.ctx.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(c.ctx)
	c.ctx.WriteHeader(code)
	return encoder.Encode(obj)
}

func (c *chiContextAdapter) BindJSON(obj interface{}) error {
	return json.NewDecoder(c.req.Body).Decode(obj)
}

func (c *chiContextAdapter) Data(code int, contentType string, data []byte) {
	c.ctx.Header().Set("Content-Type", contentType)
	c.ctx.WriteHeader(code)
	_, _ = c.ctx.Write(data)
}

func (c *chiContextAdapter) Header(key, value string) {
	c.ctx.Header().Set(key, value)
}

func (c *chiContextAdapter) Param(key string) string {
	return chi.URLParam(c.req, key)
}

func (c *chiContextAdapter) Query(key string) string {
	values := c.req.URL.Query()
	return values.Get(key)
}

func (c *chiContextAdapter) DefaultQuery(key, defaultValue string) string {
	value := c.Query(key)
	if value == "" {
		return defaultValue
	}
	return value
}
