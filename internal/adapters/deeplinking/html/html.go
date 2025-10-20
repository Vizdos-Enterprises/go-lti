package deeplinking_html

import (
	_ "embed"
	"html/template"
)

//go:embed return_to_lms.html
var returnToLMSHTML []byte

var ReturnToLMSHTML = template.Must(template.New("return").Parse(string(returnToLMSHTML)))
