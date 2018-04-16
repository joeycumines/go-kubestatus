package kubestatus

import (
	"errors"
	"net/http"
	"fmt"
	"github.com/gin-gonic/gin/json"
	"net/url"
	"strings"
)

// Client provides an interface to the server, for nested readiness checks, for example, providing short-circuiting
// logic, where the first failure status is returned
type Client struct {
	// Addresses indicates a list of addresses to Get
	Addresses []string

	// All, if set to true, will try to Get all addresses regardless of any errors
	All bool

	// UUIDs will be passed in via the query parameter
	UUIDs []string
}

func statusOK(status int) bool {
	if status < 200 {
		return false
	}
	if status >= 300 {
		return false
	}
	return true
}

// Get hits the endpoint on all clients, and returns any statuses (if valid json responses are returned and can be
// deserialized), a non-nil error will be returned if any clients return a status not in the 200 range.
func (c Client) Get(endpoint string) ([]*Status, error) {
	var (
		statuses = make([]*Status, len(c.Addresses))
		err      error
	)

	for i, address := range c.Addresses {
		var (
			httpResp *http.Response
			httpErr  error
			URL      *url.URL
		)

		URL, httpErr = url.Parse(address)

		if httpErr == nil {
			URL.Path += endpoint

			if len(c.UUIDs) != 0 {
				query := URL.Query()
				query.Set("uuids", strings.Join(c.UUIDs, ","))
				URL.RawQuery = query.Encode()
			}

			httpResp, httpErr = http.Get(URL.String())
		}

		if httpErr == nil {
			if !statusOK(httpResp.StatusCode) {
				httpErr = errors.New(httpResp.Status)
			}

			status := new(Status)
			decoder := json.NewDecoder(httpResp.Body)

			if err := decoder.Decode(status); err == nil {
				statuses[i] = status

				if httpErr != nil && status.Message != "" {
					httpErr = fmt.Errorf("%s: %s", httpErr.Error(), status.Message)
				}
			}

			httpResp.Body.Close()
		}

		if httpErr != nil {
			if err == nil {
				err = httpErr
			}

			if !c.All {
				break
			}
		}
	}

	return statuses, err
}

// Health hits `/healthz` returns a status slice of equal length to the addresses, with returned statuses for each
// (or nil), and the first error encountered (if any)
func (c Client) Health() ([]*Status, error) {
	return c.Get("/healthz")
}

// Readiness hits `/readiness` returns a status slice of equal length to the addresses, with returned statuses for each
// (or nil), and the first error encountered (if any)
func (c Client) Readiness() ([]*Status, error) {
	return c.Get("/readiness")
}
