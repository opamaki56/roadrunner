package psr7

import (
	"net/http"
	"fmt"
	"encoding/json"
	"github.com/spiral/roadrunner"
	"github.com/sirupsen/logrus"
)

type Request struct {
	Protocol string            `json:"protocol"`
	Uri      string            `json:"uri"`
	Method   string            `json:"method"`
	Headers  http.Header       `json:"headers"`
	Cookies  map[string]string `json:"cookies"`
	RawQuery string            `json:"rawQuery"`
	Uploads  fileData          `json:"fileUploads"`

	// buffers
	postData postData
}

func ParseRequest(r *http.Request) (req *Request, err error) {
	req = &Request{
		Protocol: r.Proto,
		Uri:      fmt.Sprintf("%s%s", r.Host, r.URL.String()),
		Method:   r.Method,
		Headers:  r.Header,
		Cookies:  make(map[string]string),
		RawQuery: r.URL.RawQuery,
	}

	for _, c := range r.Cookies() {
		req.Cookies[c.Name] = c.Value
	}

	if req.HasBody() {
		r.ParseMultipartForm(32 << 20)

		if req.postData, err = parseData(r); err != nil {
			return nil, err
		}

		if req.Uploads, err = parseFiles(r); err != nil {
			return nil, err
		}

		if req.Uploads != nil {
			logrus.Debug("opening files")
		}
	}

	return req, nil
}

func (r *Request) Payload() *roadrunner.Payload {
	ctx, err := json.Marshal(r)
	if err != nil {
		panic(err) //todo: change it
	}

	// todo: non parseble payloads
	body, err := json.Marshal(r.postData)
	if err != nil {
		panic(err) //todo: change it
	}

	return &roadrunner.Payload{Context: ctx, Body: body}
}

func (r *Request) Close() {
	if r.Uploads != nil {

	}
}

// HasBody returns true if request might include POST data or file uploads.
func (r *Request) HasBody() bool {
	return r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH"
}

// parse incoming data request into JSON (including multipart form data)
func parseData(r *http.Request) (postData, error) {
	data := make(postData)
	for k, v := range r.MultipartForm.Value {
		data.push(k, v)
	}

	return data, nil
}

// parse incoming data request into JSON (including multipart form data)
func parseFiles(r *http.Request) (fileData, error) {
	data := make(fileData)
	for k, v := range r.MultipartForm.File {
		data.push(k, v)
	}

	return data, nil
}
