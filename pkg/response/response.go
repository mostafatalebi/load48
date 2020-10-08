package response

type Response struct {
	variables map[string]interface{}
}

func NewResponse() *Response {
	return &Response{}
}

func (r *Response) ProcessVariables(content string, contentType string, variablesMap map[string]string) {

}