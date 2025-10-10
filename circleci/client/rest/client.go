package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL     *url.URL
	circleToken string
	client      *http.Client
}

func New(host, endpoint, circleToken string) *Client {
	// Ensure endpoint ends with a slash
	if !strings.HasSuffix(endpoint, "/") {
		endpoint += "/"
	}

	u, _ := url.Parse(host)
	return &Client{
		baseURL:     u.ResolveReference(&url.URL{Path: endpoint}),
		circleToken: circleToken,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) NewRequest(method string, u *url.URL, payload interface{}) (req *http.Request, err error) {
	var r io.Reader
	var bodyBytes []byte

	if payload != nil {
		bodyBytes, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		r = bytes.NewReader(bodyBytes)
	}

	req, err = http.NewRequest(method, c.baseURL.ResolveReference(u).String(), r)
	if err != nil {
		return nil, err
	}

	// Set GetBody to allow request body to be replayed for retries
	if payload != nil {
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(bodyBytes)), nil
		}
	}

	req.Header.Set("Circle-Token", c.circleToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "mrolla/terraform-provider-circleci")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (c *Client) DoRequest(req *http.Request, resp interface{}) (statusCode int, err error) {
	const (
		maxRetries     = 5
		initialBackoff = 1 * time.Second
		maxBackoff     = 32 * time.Second
	)

	var httpResp *http.Response
	var bodyBytes []byte

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Clone the request body if it exists, since it can only be read once
		if req.GetBody != nil && attempt > 0 {
			req.Body, err = req.GetBody()
			if err != nil {
				return 0, err
			}
		}

		httpResp, err = c.client.Do(req)
		if err != nil {
			return 0, err
		}

		// Read the response body
		bodyBytes, err = io.ReadAll(httpResp.Body)
		httpResp.Body.Close()
		if err != nil {
			return httpResp.StatusCode, err
		}

		// Check for rate limiting (429)
		if httpResp.StatusCode == http.StatusTooManyRequests && attempt < maxRetries {
			// Calculate exponential backoff with jitter
			backoff := time.Duration(math.Pow(2, float64(attempt))) * initialBackoff
			if backoff > maxBackoff {
				backoff = maxBackoff
			}

			time.Sleep(backoff)
			continue
		}

		// Not a rate limit error or max retries reached, break out of retry loop
		break
	}

	if httpResp.StatusCode >= 300 {
		httpError := struct {
			Message string `json:"message"`
		}{}
		err = json.Unmarshal(bodyBytes, &httpError)
		if err != nil {
			return httpResp.StatusCode, err
		}
		return httpResp.StatusCode, &HTTPError{Code: httpResp.StatusCode, Message: httpError.Message}
	}

	if resp != nil {
		// REST API v2 does not currently return Content-Type

		// if !strings.Contains(httpResp.Header.Get("Content-Type"), "application/json") {
		// 	return httpResp.StatusCode, errors.New("wrong content type received")
		// }

		err = json.Unmarshal(bodyBytes, resp)
		if err != nil {
			return httpResp.StatusCode, err
		}
	}
	return httpResp.StatusCode, nil
}

type HTTPError struct {
	Code    int
	Message string
}

func (e *HTTPError) Error() string {
	if e.Code == 0 {
		e.Code = http.StatusInternalServerError
	}
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("response %d (%s)", e.Code, http.StatusText(e.Code))
}
