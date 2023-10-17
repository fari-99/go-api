package generate_pdf

import (
	"bytes"
	"fmt"
	"html/template"
)

//pdf request pdf struct
type RequestPdf struct {
	body string
}

//new request to pdf function
func NewRequestPdf(body string) *RequestPdf {
	return &RequestPdf{
		body: body,
	}
}

//parsing template function
func (r *RequestPdf) ParseTemplate(templateFileName string, data interface{}) error {

	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		return fmt.Errorf("failed to parse files, err := %s", err.Error())
	}

	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return fmt.Errorf("failed to execute parsed files, err := %s", err.Error())
	}

	r.body = buf.String()
	return nil
}
