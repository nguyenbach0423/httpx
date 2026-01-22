package middleware

import "github.com/nguyenbach0423/httpx/server"

func WithParam(k string, v any) server.MiddlewareFunc {
	return func(next server.HandlerFunc) server.HandlerFunc {
		return func(c *server.Context) {
			c.Params[k] = v
			next(c)
		}
	}
}

func WithParams(params map[string]any) server.MiddlewareFunc {
	return func(next server.HandlerFunc) server.HandlerFunc {
		return func(c *server.Context) {
			for k, v := range params {
				c.Params[k] = v
			}
			next(c)
		}
	}
}
