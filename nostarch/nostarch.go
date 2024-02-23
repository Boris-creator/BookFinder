package nostarch

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type bookInfo struct {
	Title       string
	Authors     string
	Description string
	Isbn        string
}

func Search(query string) error {
	books, err := searchBooks(query)
	if err != nil {
		return err
	}
	fmt.Println(books)
	return nil
}

func searchBooks(query string) ([]bookInfo, error) {
	baseUrl := "https://nostarch.com/"
	searchUrl, _ := url.JoinPath(baseUrl, "search", url.PathEscape(query))
	res, err := http.Get(searchUrl)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	bookBlocks := doc.Find("ol.search-results article.node-product")
	books := make([]bookInfo, 0, bookBlocks.Length())
	var wg sync.WaitGroup
	wg.Add(bookBlocks.Length())

	bookBlocks.Each(func(i int, s *goquery.Selection) {
		title := s.Find("header h2").Text()
		authors := s.Find(".field-name-field-author").Text()
		link, hasLink := s.Find("a").Attr("href")
		if !hasLink {
			return
		}
		book := bookInfo{
			Title:   title,
			Authors: authors,
		}
		books = append(books, book)
		go func() {
			defer wg.Done()
			bookPageUrl, _ := url.JoinPath(baseUrl, link)
			res, err := http.Get(bookPageUrl)
			if err != nil {
				return
			}
			doc, err := goquery.NewDocumentFromReader(res.Body)
			if err != nil {
				return
			}
			isbn := doc.Find(".field-name-field-isbn13 .field-items").Text()
			if len(isbn) != 0 {
				books[i].Isbn = isbn
			}
		}()
	})

	wg.Wait()
	return books, nil
}
