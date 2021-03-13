package gateway

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"path"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
)

const (
	APIPath        = "/api"
	AggregatesPath = "meters/aggregates"
	StatusPath     = "system_status/soe"
	LoginPath      = "login/Basic"

	UserName   = "customer"
	ForceSMOff = false
)

type Aggregates struct {
	Values map[string]interface{}
}

func (a *Aggregates) UnmarshalJSON(b []byte) error {
	a.Values = make(map[string]interface{})

	objects := make(map[string]map[string]interface{})
	if err := json.Unmarshal(b, &objects); err != nil {
		return err
	}

	// flatten the keyspace
	for object, blob := range objects {
		for k, v := range blob {
			a.Values[strings.Join([]string{object, k}, "_")] = v
		}
	}
	return nil
}

func (a Aggregates) Path() string {
	return path.Join(APIPath, AggregatesPath)
}

type SOE struct {
	Percentage float64
}

func (s SOE) Path() string {
	return path.Join(APIPath, StatusPath)
}

type Login struct {
	Username   string
	Password   string
	Email      string
	ForceSmOff bool `json: force_sm_off`
}

func NewLogin(email, password string) *Login {
	return &Login{
		Username:   UserName,
		Password:   password,
		Email:      email,
		ForceSmOff: ForceSMOff,
	}
}

func (l Login) Path() string {
	return path.Join(APIPath, LoginPath)
}

func (l *Login) Request(u *url.URL) (*http.Request, error) {
	b := new(bytes.Buffer)

	e := json.NewEncoder(b)
	if err := e.Encode(l); err != nil {
		return nil, err
	}

	return http.NewRequest("POST", u.String()+l.Path(), b)
}

type Client struct {
	c *http.Client
	u *url.URL
}

func NewClient(u *url.URL, email, password string) (*Client, error) {
	j, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, err
	}
	c := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		},
		Jar: j,
	}

	l := NewLogin(email, password)
	b, err := l.Request(u)
	if err != nil {
		return nil, err
	}

	r, err := c.Do(b)
	if err != nil {
		return nil, err
	}

	if r.StatusCode != http.StatusOK {
		buf, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, errors.New("login error")
		}

		return nil, errors.New(string(buf))
	}

	return &Client{c, u}, nil
}

func (c *Client) get(a interface{}) error {
	p := a.(interface {
		Path() string
	})

	resp, err := c.c.Get(c.u.String() + p.Path())
	if err != nil {
		return err
	}

	defer func() {
		if resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if resp.StatusCode != http.StatusOK {
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.New("aggregates error")
		}

		return errors.New(string(buf))
	}

	d := json.NewDecoder(resp.Body)
	if err := d.Decode(a); err != nil {
		return err
	}

	return nil
}

func (c *Client) Aggregates() (*Aggregates, error) {
	a := new(Aggregates)
	if err := c.get(a); err != nil {
		return nil, err
	}

	return a, nil
}

func (c *Client) SOE() (*SOE, error) {
	s := new(SOE)
	if err := c.get(s); err != nil {
		return nil, err
	}

	return s, nil
}
