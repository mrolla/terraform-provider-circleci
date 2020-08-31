package circleci

import "net/http"

const (
	_testCCIToken = "fake-cci-token"
	_testBaseURL  = "http://fake.url"
)

var (
	_testHTTPClient = http.DefaultClient
)
