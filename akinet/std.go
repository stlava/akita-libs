package akinet

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/uuid"
)

func FromStdRequest(streamID uuid.UUID, seq int, src *http.Request, body []byte) HTTPRequest {
	return HTTPRequest{
		StreamID:   streamID,
		Seq:        seq,
		Method:     src.Method,
		ProtoMajor: src.ProtoMajor,
		ProtoMinor: src.ProtoMinor,
		URL:        src.URL,
		Host:       src.Host,
		Header:     src.Header,
		Body:       body,
	}
}

func (r HTTPRequest) ToStdRequest() *http.Request {
	result := &http.Request{
		Method:        r.Method,
		URL:           r.URL,
		Proto:         fmt.Sprintf("HTTP/%d.%d", r.ProtoMajor, r.ProtoMinor),
		ProtoMajor:    r.ProtoMajor,
		ProtoMinor:    r.ProtoMinor,
		Host:          r.Host,
		Header:        r.Header,
		ContentLength: int64(len(r.Body)),
		Body:          ioutil.NopCloser(bytes.NewReader(r.Body)),
	}

	for _, c := range r.Cookies {
		result.AddCookie(c)
	}
	return result
}

func FromStdResponse(streamID uuid.UUID, seq int, src *http.Response, body []byte) HTTPResponse {
	return HTTPResponse{
		StreamID:   streamID,
		Seq:        seq,
		StatusCode: src.StatusCode,
		ProtoMajor: src.ProtoMajor,
		ProtoMinor: src.ProtoMinor,
		Header:     src.Header,
		Body:       body,
	}
}

func (r HTTPResponse) ToStdResponse() *http.Response {
	return &http.Response{
		Status:        http.StatusText(r.StatusCode),
		StatusCode:    r.StatusCode,
		Proto:         fmt.Sprintf("HTTP/%d.%d", r.ProtoMajor, r.ProtoMinor),
		ProtoMajor:    r.ProtoMajor,
		ProtoMinor:    r.ProtoMinor,
		Header:        r.Header,
		ContentLength: int64(len(r.Body)),
		Body:          ioutil.NopCloser(bytes.NewReader(r.Body)),
	}
}
