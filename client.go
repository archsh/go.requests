package requests

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"
)

var builtinClient = &http.Client{}

func SetupClient(c http.CookieJar, keepalive, timeout time.Duration, proxy string, skipVerifySSL bool) error {
	//cli := http.Client{Timeout: timeout,}
	defaultTransport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: keepalive,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 10,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: skipVerifySSL},
	}
	if proxy != "" {
		if pu, e := url.Parse(proxy); nil != e {
			return e
		} else {
			defaultTransport.Proxy = http.ProxyURL(pu)
		}
	}
	if nil != c {
		builtinClient.Jar = c
	}
	builtinClient.Transport = defaultTransport
	return nil
}

func init() {
	_ = SetupClient(nil, 30*time.Second, 30*time.Second, "", true)
}
