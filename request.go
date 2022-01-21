package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strconv"
	"time"
)

var headers = map[string]string{
	"Content-Type": "application/json",
	"Connection":   "Keep-Alive",
	"User-Agent":   "android-ok-http-client/xl-acc-sdk/version-3.1.2.185150",
}

type Request struct {
	Client  *http.Client
	Cookies cookiejar.Jar
}

func NewRequest() *Request {
	jar, _ := cookiejar.New(nil)
	return &Request{Client: &http.Client{
		Jar:     jar,
		Timeout: 5 * time.Second,
	}}
}

func (x *Request) Post(url string, body []byte) ([]byte, error) {
	var r io.Reader
	if body != nil {
		//fmt.Printf("Body:%s\n", body)
		r = bytes.NewReader(body)
	}
	req, _ := http.NewRequest(http.MethodPost, url, r)
	return x.Do(req)
}

func (x *Request) Get(url string) ([]byte, error) {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	return x.Do(req)
}

func (x *Request) Do(req *http.Request) ([]byte, error) {
	//fmt.Printf("%s -> %s\n", req.Method, req.URL.String())
	setHeaders(req, headers)
	for count := 0; count < 3; count++ {
		resp, err := x.Client.Do(req)
		if err != nil {
			if os.IsTimeout(err) {
				continue
			}
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, errors.New("http code error " + strconv.Itoa(resp.StatusCode))
		}
		b, err := io.ReadAll(resp.Body)
		ForceClose(resp.Body)
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return nil, context.DeadlineExceeded
}

func setHeaders(req *http.Request, headers map[string]string) {
	for key, value := range headers {
		req.Header.Set(key, value)
	}
}
