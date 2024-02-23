package manning

import (
	"bookfinder/search"
	httputils "bookfinder/utils/http"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type BookProductInfo struct {
	ProductManningId int    `json:"productManningId"`
	ProductTitle     string `json:"productTitle"`
	Description      string `json:"description"`
}

type BookAdditionInfo struct {
	Isbn        string `json:"isbn"`
	Id          int    `json:"id"`
	Description string `json:"description"`
}

type SearchResults struct {
	ProductPagesResponse search.SearchResults[BookProductInfo]
}

func Search(query string) error {
	books, err := searchBooks(query)
	if err != nil {
		return err
	}
	fmt.Println(books)
	return nil
}

func searchBooks(query string) (search.SearchResults[BookProductInfo], error) {
	var results SearchResults
	var resultsAdditionInfo []BookAdditionInfo
	searchData, err := httputils.FetchWithTimeout(func() (*http.Response, error) {
		return http.Get(
			fmt.Sprintf("https://www.manning.com/nsearch/shallowSearch?q=%s&category=all&dontReturnText=true&returnElementInfo=true&lemma=%s", query, query),
		)
	})()
	if err != nil {
		return results.ProductPagesResponse, err
	}
	resBody, err := io.ReadAll(searchData)
	searchData.Close()
	if err != nil {
		return results.ProductPagesResponse, err
	}
	json.Unmarshal(resBody, &results)

	booksFound := &results.ProductPagesResponse.Results
	booksIds := make([]string, 0, len(*booksFound))
	booksByIds := make(map[int]*BookProductInfo, len(*booksFound))
	for i, book := range *booksFound {
		booksIds = append(booksIds, fmt.Sprintf("%d", book.ProductManningId))
		booksByIds[book.ProductManningId] = &(*booksFound)[i]
	}

	queryParams := url.Values{
		"productIds": []string{strings.Join(booksIds, ",")},
	}
	response, _ := httputils.Fetch(func() (*http.Response, error) {
		return http.PostForm(
			"https://www.manning.com/search/additionalProductInformation",
			queryParams,
		)
	})()
	resBody, err = io.ReadAll(response)
	defer response.Close()
	if err != nil {
		return results.ProductPagesResponse, err
	}
	json.Unmarshal(resBody, &resultsAdditionInfo)

	for _, info := range resultsAdditionInfo {
		book, ok := booksByIds[info.Id]
		if ok {
			book.Description = info.Description
		}
	}

	return results.ProductPagesResponse, nil
}
