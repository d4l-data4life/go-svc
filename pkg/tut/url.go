package tut

import (
	"fmt"
)

// CreateURLBuilder returns a function that takes a URL as a parameter and returns
// the URL prefixed with "http://<HOST>:<PORT>".
func CreateURLBuilder(host string, port string) func(url string) string {
	return func(url string) string {
		baseURL := fmt.Sprintf("http://%s:%s", host, port)

		return fmt.Sprintf("%s%s", baseURL, url)
	}
}

// QueryValues returns a url.Values map from string values
func QueryValues(key string, values ...string) map[string][]string {
	query := make(map[string][]string)

	for _, value := range values {
		if q, ok := query[key]; ok {
			query[key] = append(q, value)
			continue
		}

		query[key] = []string{value}
	}

	return query
}
