package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/nguyenbach0423/httpx/server"
)

func Recover() server.MiddlewareFunc {
	return func(next server.HandlerFunc) server.HandlerFunc {
		return func(c *server.Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					if r == http.ErrAbortHandler {
						panic(r)
					}

					switch e := r.(type) {
					case error:
						err = e
					case string:
						err = errors.New(e)
					default:
						err = errors.New(fmt.Sprint(e))
					}
				}
			}()

			return next(c)
		}
	}
}
