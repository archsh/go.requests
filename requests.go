package requests

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"strings"

	yaml "gopkg.in/yaml.v2"
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

func PostForm(url string, data url.Values) (resp *Response, err error) {
	return Post(url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

func PostJSON(url string, o interface{}) (resp *Response, err error) {
	if bs, e := json.Marshal(o); nil != e {
		return nil, e
	} else {
		return Post(url, "application/json", bytes.NewReader(bs))
	}
}

func PostYAML(url string, o interface{}) (resp *Response, err error) {
	if bs, e := yaml.Marshal(o); nil != e {
		return nil, e
	} else {
		return Post(url, "application/yaml", bytes.NewReader(bs))
	}
}

func PostXML(url string, o interface{}) (resp *Response, err error) {
	if bs, e := xml.Marshal(o); nil != e {
		return nil, e
	} else {
		return Post(url, "application/xml", bytes.NewReader(bs))
	}
}
