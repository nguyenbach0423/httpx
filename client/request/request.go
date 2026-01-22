package request

type Request struct {
	Method      string
	URL         string
	Headers     map[string]string
	QueryParams map[string][]string
	Body        []byte
}

func New(method, url string, opts ...func(*Request)) *Request {
	r := &Request{
		Method:      method,
		URL:         url,
		Headers:     make(map[string]string),
		QueryParams: make(map[string][]string),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

func WithHeader(key, value string) func(*Request) {
	return func(r *Request) {
		if r.Headers == nil {
			r.Headers = make(map[string]string)
		}

		r.Headers[key] = value
	}
}

func WithHeaders(headers map[string]string) func(*Request) {
	return func(r *Request) {
		if r.Headers == nil {
			r.Headers = make(map[string]string)
		}

		for k, v := range headers {
			r.Headers[k] = v
		}
	}
}

func WithQueryParam(key string, value []string) func(*Request) {
	return func(r *Request) {
		if r.QueryParams == nil {
			r.QueryParams = make(map[string][]string)
		}

		r.QueryParams[key] = append(r.QueryParams[key], value...)
	}
}

func WithQueryParams(queryParams map[string][]string) func(*Request) {
	return func(r *Request) {
		if r.QueryParams == nil {
			r.QueryParams = make(map[string][]string)
		}

		for k, v := range queryParams {
			r.QueryParams[k] = append(r.QueryParams[k], v...)
		}
	}
}

func WithBody(body []byte) func(*Request) {
	return func(r *Request) {
		r.Body = body
	}
}
