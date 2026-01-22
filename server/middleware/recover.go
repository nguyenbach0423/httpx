package middleware

import (
	"fmt"
	"net/http"

	"github.com/nguyenbach0423/httpx/server"
)

func Recover() server.MiddlewareFunc {
	return func(next server.HandlerFunc) server.HandlerFunc {
		return func(c *server.Context) {
			defer func() {
				if r := recover(); r != nil {
					if r == http.ErrAbortHandler {
						panic(r)
					}

					var err error

					switch v := r.(type) {
					case error:
						err = v
					default:
						err = fmt.Errorf("%v", v)
					}

					c.AbortWithError(http.StatusInternalServerError, err)
				}
			}()

			next(c)
		}
	}
}
