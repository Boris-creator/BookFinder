package apress

import (
	"bookfinder/search"
	"bookfinder/store"
	dbutils "bookfinder/utils/db"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type bookInfo struct {
	Title       string
	Description string
	Isbn        string
	Pages       string
}

func Search(query string) error {
	books, err := searchBooks(query)
	if err != nil {
		return err
	}
	var saveConfig dbutils.Config[bookInfo, store.BookModel]
	err = dbutils.BulkInsert[bookInfo, store.BookModel]("books", books, saveConfig.Prepare(func(value bookInfo) store.BookModel {
		return store.BookModel{
			Title:       value.Title,
			Isbn:        value.Isbn,
			Description: value.Description,
			Source:      int(search.APress),
		}
	}))
	return err
}

var baseUrl = "https://link.springer.com"
var booksPerPage = 20

func parseBooksList(document *goquery.Document) []string {
	var links []string
	document.Find("#results-list li a.title").Each(func(i int, s *goquery.Selection) {
		links = append(links, s.AttrOr("href", ""))
	})
	return links
}
func parseBookCard(document *goquery.Document) bookInfo {
	var info bookInfo
	info.Title = strings.TrimSpace(document.Find("[data-test='book-title']").Text())

	descriptionBlock := document.Find("[data-title='About this book'] .c-book-section")
	if description, err := descriptionBlock.Html(); err == nil {
		info.Description = strings.TrimSpace(description)
	} else {
		info.Description = descriptionBlock.Text()
	}

	bibliographyBlocks := document.Find(".c-bibliographic-information__list-item")
	bibliography := make(map[string]string, bibliographyBlocks.Length())
	bibliographyBlocks.Each(func(i int, block *goquery.Selection) {
		field := block.Find("p span").First().Text()
		bibliography[field] = strings.TrimSpace(block.Find(".c-bibliographic-information__value").Text())
	})
	for key, value := range bibliography {
		if strings.Contains(key, "eBook ISBN") {
			info.Isbn = strings.ReplaceAll(value, "-", "")
		}
		if strings.Contains(key, "Number of Pages") {
			info.Pages = value
		}
	}
	return info
}

func searchBooks(query string) ([]bookInfo, error) {
	var results []bookInfo
	jar, err := cookiejar.New(nil)
	if err != nil {
		return results, err
	}

	client := &http.Client{
		Jar: jar,
	}
	getUrl := func(page int) string {
		pageUrl := fmt.Sprintf("/search/page/%d?query=%s&package=41786&facet-content-type=%%22Book%%22", page, url.QueryEscape(query))
		resUrl, _ := url.JoinPath(baseUrl, pageUrl)
		return resUrl
	}
	res, err := client.Get(getUrl(1))
	if err != nil {
		return results, err
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return results, err
	}
	pagesTotal, _ := strconv.Atoi(doc.Find("form.pagination .number-of-pages").First().Text())
	chLinks := make(chan string)
	chBooks := make(chan bookInfo, pagesTotal*booksPerPage)
	defer close(chLinks)

	go func() {
		for _, link := range parseBooksList(doc) {
			chLinks <- link
		}
	}()

	var wg sync.WaitGroup
	if pagesTotal > 1 {
		wg.Add(pagesTotal - 1)
	}

	for page := 2; page <= pagesTotal; page++ {
		go func(p int) {
			defer wg.Done()
			res, _ := client.Get(getUrl(p))
			doc, _ := goquery.NewDocumentFromReader(res.Body)
			for _, link := range parseBooksList(doc) {
				chLinks <- link
			}
		}(page)
	}

	go func() {
		for link := range chLinks {
			wg.Add(1)
			go func(link string) {
				defer wg.Done()
				bookUrl, _ := url.JoinPath(baseUrl, link)
				res, _ := client.Get(bookUrl)
				doc, _ := goquery.NewDocumentFromReader(res.Body)
				chBooks <- parseBookCard(doc)
			}(link)
		}
	}()

	wg.Wait()
	close(chBooks)

	for book := range chBooks {
		results = append(results, book)
	}

	return results, nil
}
