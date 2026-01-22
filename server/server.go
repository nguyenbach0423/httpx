package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type Server struct {
	httpserver  *http.Server
	middlewares []MiddlewareFunc
	routes      []Route
}

func New(opts ...func(server *Server)) *Server {
	s := &Server{
		middlewares: []MiddlewareFunc{},
		routes:      []Route{},
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func WithHealthCheck() func(*Server) {
	return func(s *Server) {
		s.Get("/health/live", func(c *Context) error {
			return c.JSON(http.StatusOK, map[string]string{
				"msg": "ok",
			})
		})
	}
}

func (s *Server) ListenAndServe(port string) error {
	s.httpserver = &http.Server{Addr: ":" + port, Handler: s}

	if err := s.httpserver.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpserver != nil {
		return s.httpserver.Shutdown(ctx)
	}

	return nil
}

func (s *Server) Use(mw ...MiddlewareFunc) {
	s.middlewares = append(s.middlewares, mw...)
}

func (s *Server) Get(path string, h HandlerFunc, mw ...MiddlewareFunc) {
	s.routes = append(s.routes, Route{http.MethodGet, cleanPath(path), applyMiddleware(h, mw...)})
}

func (s *Server) Post(path string, h HandlerFunc, mw ...MiddlewareFunc) {
	s.routes = append(s.routes, Route{http.MethodPost, cleanPath(path), applyMiddleware(h, mw...)})
}

func (s *Server) Put(path string, h HandlerFunc, mw ...MiddlewareFunc) {
	s.routes = append(s.routes, Route{http.MethodPut, cleanPath(path), applyMiddleware(h, mw...)})
}

func (s *Server) Patch(path string, h HandlerFunc, mw ...MiddlewareFunc) {
	s.routes = append(s.routes, Route{http.MethodPatch, cleanPath(path), applyMiddleware(h, mw...)})
}

func (s *Server) Delete(path string, h HandlerFunc, mw ...MiddlewareFunc) {
	s.routes = append(s.routes, Route{http.MethodDelete, cleanPath(path), applyMiddleware(h, mw...)})
}

func (s *Server) Group(prefix string, mw ...MiddlewareFunc) *Group {
	return &Group{
		prefix:      cleanPath(prefix),
		middlewares: append([]MiddlewareFunc{}, mw...),
		server:      s,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := &Context{Request: r, Response: &Response{writer: w}}

	h := s.findHandler(c)

	if h != nil {
		if err := applyMiddleware(h, s.middlewares...)(c); err != nil {
			c.Response.writer.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (s *Server) findHandler(c *Context) HandlerFunc {
	methodMatched := true

	requestPathParts := strings.Split(strings.Trim(c.Request.URL.Path, "/"), "/")

	for _, route := range s.routes {
		routePathParts := strings.Split(strings.Trim(route.path, "/"), "/")

		if len(routePathParts) != len(requestPathParts) {
			continue
		}

		pathMatched := true

		pathParams := make(map[string]string)
		for i, routePathPart := range routePathParts {
			if strings.HasPrefix(routePathPart, ":") {
				pathParams[routePathPart[1:]] = requestPathParts[i]
				continue
			}

			if routePathPart != requestPathParts[i] {
				pathMatched = false
				break
			}
		}

		if !pathMatched {
			continue
		}

		if route.method == c.Request.Method {
			for key, value := range pathParams {
				c.pathValues[key] = value
			}
			return route.handler
		}

		methodMatched = false
	}

	if !methodMatched {
		c.Response.writer.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}

	c.Response.writer.WriteHeader(http.StatusNotFound)
	return nil
}

type Group struct {
	prefix      string
	middlewares []MiddlewareFunc
	server      *Server
}

func (g *Group) Use(mw ...MiddlewareFunc) {
	g.middlewares = append(g.middlewares, mw...)
}

func (g *Group) Get(path string, h HandlerFunc, mw ...MiddlewareFunc) {
	g.server.Get(g.prefix+cleanPath(path), h, append(append([]MiddlewareFunc{}, g.middlewares...), mw...)...)
}

func (g *Group) Post(path string, h HandlerFunc, mw ...MiddlewareFunc) {
	g.server.Post(g.prefix+cleanPath(path), h, append(append([]MiddlewareFunc{}, g.middlewares...), mw...)...)
}

func (g *Group) Put(path string, h HandlerFunc, mw ...MiddlewareFunc) {
	g.server.Put(g.prefix+cleanPath(path), h, append(append([]MiddlewareFunc{}, g.middlewares...), mw...)...)
}

func (g *Group) Patch(path string, h HandlerFunc, mw ...MiddlewareFunc) {
	g.server.Patch(g.prefix+cleanPath(path), h, append(append([]MiddlewareFunc{}, g.middlewares...), mw...)...)
}

func (g *Group) Delete(path string, h HandlerFunc, mw ...MiddlewareFunc) {
	g.server.Delete(g.prefix+cleanPath(path), h, append(append([]MiddlewareFunc{}, g.middlewares...), mw...)...)
}

func (g *Group) Group(prefix string, mw ...MiddlewareFunc) *Group {
	return g.server.Group(g.prefix+cleanPath(prefix), append(append([]MiddlewareFunc{}, g.middlewares...), mw...)...)
}

type Context struct {
	Request  *http.Request
	Response *Response

	pathValues map[string]string

	params map[string]any
}

type Response struct {
	writer http.ResponseWriter
}

func (c *Context) Param(key string) any {
	return c.params[key]
}

func (c *Context) SetParam(key string, value any) {
	c.params[key] = value
}

func (c *Context) PathValue(key string) string {
	return c.pathValues[key]
}

func (c *Context) QueryParam(key string) string {
	return c.Request.URL.Query().Get(key)
}

func (c *Context) RequestBody(v any) error {
	return json.NewDecoder(c.Request.Body).Decode(v)
}

func (c *Context) JSON(status int, v any) error {
	w := c.Response.writer
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

type HandlerFunc func(*Context) error

type MiddlewareFunc func(HandlerFunc) HandlerFunc

type Route struct {
	method  string
	path    string
	handler HandlerFunc
}

func cleanPath(path string) string {
	path = strings.TrimSpace(path)

	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}

	path = strings.TrimSuffix(path, "/")

	if path != "" && !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return path
}

func applyMiddleware(h HandlerFunc, mw ...MiddlewareFunc) HandlerFunc {
	for i := len(mw) - 1; i >= 0; i-- {
		h = mw[i](h)
	}
	return h
}
