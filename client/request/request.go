package request

type Request struct {
	Method      string
	URL         string
	Headers     map[string]string
	QueryParams map[string][]string
	Body        []byte
}
