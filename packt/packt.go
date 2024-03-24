package packt

import (
	"fmt"
	"net/url"
)

func searchBooks(query string) {

	var params = url.URL{
		Scheme: "https",
		Host:   "vivzzxfqg1-dsn.algolia.net",
		Path:   "/1/indexes/*/",
	}
	rq := url.Values{}
	rq.Add("x-algolia-agent", "Algolia for JavaScript (4.13.0); Browser; JS Helper (3.13.3); react (17.0.1); react-instantsearch (6.40.1)")
	params.RawQuery = rq.Encode()
	fmt.Println(params.String())
}
