package middleware

import "github.com/nguyenbach0423/httpx/server"

func WithParam(k string, v any) server.MiddlewareFunc {
	return func(next server.HandlerFunc) server.HandlerFunc {
		return func(c *server.Context) error {
			c.SetParam(k, v)
			return next(c)
		}
	}
}

func WithParams(params map[string]any) server.MiddlewareFunc {
	return func(next server.HandlerFunc) server.HandlerFunc {
		return func(c *server.Context) error {
			for k, v := range params {
				c.SetParam(k, v)
			}
			return next(c)
		}
	}
}
