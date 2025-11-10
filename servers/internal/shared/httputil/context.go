package httputil

import (
	"context"
	"net/http"
)

type HttpRequestContext struct {
	Writer  http.ResponseWriter
	Request *http.Request
}

func NewHttpUtilContext(
	w http.ResponseWriter,
	r *http.Request,
) *HttpRequestContext {
	return &HttpRequestContext{Writer: w, Request: r}
}

func (c *HttpRequestContext) Ctx() context.Context {
	return c.Request.Context()
}
