package requests

import (
	"io"
	"net/http"
)

func NewRequest(method, url string, body io.Reader) (*Request, error) {
	if r, e := http.NewRequest(method, url, body); nil != e {
		return &Request{
			_req: r,
		}, nil
	} else {
		return nil, e
	}
}

func Do(req *Request) (*Response, error) {
	if resp, e := builtinClient.Do(req._req); nil != e {
		return nil, e
	} else {
		return &Response{_resp: resp}, nil
	}
}

func Get(url string) (*Response, error) {
	if q, e := NewRequest("GET", url, nil); nil != e {
		return nil, e
	} else {
		return Do(q)
	}
}

func Post(url string, contentType string, body io.Reader) (*Response, error) {
	if q, e := NewRequest("POST", url, body); nil != e {
		return nil, e
	} else {
		q._req.Header.Add("Content-Type", contentType)
		return Do(q)
	}
}
