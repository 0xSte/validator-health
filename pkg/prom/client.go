package prom

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/common/config"
)

type Client struct {
	c            api.Client
	roundTripper http.RoundTripper
	address      *string
	auth         bool
	timeout      time.Duration
}

type ErrInvalidPromClientConfig struct {
	Field string
}

func (e *ErrInvalidPromClientConfig) Error() string {
	return fmt.Sprintf("invalid config for prom client: %s", e.Field)
}

func New(options ...func(c *Client)) (*Client, error) {
	client := &Client{
		timeout: time.Second * 10,
	}
	for _, option := range options {
		option(client)
	}

	if !client.auth {
		return nil, &ErrInvalidPromClientConfig{Field: "basicAuth"}
	}
	if client.address == nil {
		return nil, &ErrInvalidPromClientConfig{Field: "address"}
	}
	c, err := api.NewClient(api.Config{
		Address:      *client.address,
		RoundTripper: client.roundTripper,
	})
	if err != nil {
		return client, err
	}
	client.c = c
	return client, nil
}

func WithAddress(address string) func(reconfigure *Client) {
	return func(c *Client) {
		c.address = &address
	}
}

func WithBasicAuth(user, pass string) func(reconfigure *Client) {
	return func(c *Client) {
		if user != "" && pass != "" {
			c.auth = true
		}
		c.roundTripper = &roundTripper{
			userAgent: "github.com/0xste/validator-stats",
			username:  user,
			password:  config.Secret(pass),
			rt:        api.DefaultRoundTripper,
		}
	}
}

type roundTripper struct {
	userAgent string
	username  string
	password  config.Secret
	rt        http.RoundTripper
}

func (rt *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(req.Header.Get("Authorization")) != 0 {
		return rt.rt.RoundTrip(req)
	}
	req = cloneRequest(req)
	req.SetBasicAuth(rt.username, strings.TrimSpace(string(rt.password)))
	if req.UserAgent() == "" {
		// The specification of http.RoundTripper says that it shouldn't mutate
		// the request so make a copy of req.Header since this is all that is
		// modified.
		r2 := new(http.Request)
		*r2 = *req
		r2.Header = make(http.Header)
		for k, s := range req.Header {
			r2.Header[k] = s
		}
		r2.Header.Set("User-Agent", rt.userAgent)
		req = r2
	}

	return rt.rt.RoundTrip(req)
}

// cloneRequest returns a clone of the provided *http.Request.
// The clone is a shallow copy of the struct and its Header map.
func cloneRequest(r *http.Request) *http.Request {
	// Shallow copy of the struct.
	r2 := new(http.Request)
	*r2 = *r
	// Deep copy of the Header.
	r2.Header = make(http.Header)
	for k, s := range r.Header {
		r2.Header[k] = s
	}
	return r2
}
