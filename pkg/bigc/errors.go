package bigc

import "fmt"

type RequestError struct {
	StatusCode int
	Err        error
	Body       []byte
}

func (r *RequestError) Error() string {
	return fmt.Sprintf("status %d: err: %v\nbody: %v", r.StatusCode, r.Err, string(r.Body))
}

func (r *RequestError) Status() int {
	return r.StatusCode
}

type NotImplementedError struct {
	Message string
}

func (m *NotImplementedError) Error() string {
	if m.Message != "" {
		return m.Message
	}
	return "Functionality not implemented."
}

type ProductNotFoundError struct {
	Sku     string
	Source  string
	Message string
}

func (m *ProductNotFoundError) Error() string {
	msg := "Product not found"
	if m.Source != "" {
		msg += " in " + m.Source
	}
	if m.Sku != "" {
		msg += "\tSku: " + m.Sku
	}
	return msg
}
