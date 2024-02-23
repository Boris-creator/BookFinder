package oreilly

import (
	"bookfinder/search"
	dbutils "bookfinder/utils/db"
	fileutils "bookfinder/utils/file"
	httputils "bookfinder/utils/http"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"net/url"
	//"path/filepath"
	"strings"
	"sync"
)

type book struct {
	Title       string
	Authors     []string
	Description string
	Isbn        string
	CoverUrl    string `json:"cover_url"`
}

type bookModel struct {
	Title       string `db:"title"`
	Description string `db:"description"`
	Isbn        string `db:"isbn"`
	Source      int    `db:"source"`
}
type imageModel struct {
	Name   string `db:"name"`
	Image  []byte `db:"image"`
	BookId int    `db:"bookId"`
	Source int    `db:"source"`
	Hash   string `db:"hash"`
}

func SearchAndSave(query string) error {
	results, err := searchBooks(query)
	if err != nil {
		return err
	}

	var saveConfig dbutils.Config[book, bookModel]
	err = dbutils.BulkInsert[book, bookModel]("books", results.Results, saveConfig.Prepare(func(b book) bookModel {
		return bookModel{
			b.Title,
			b.Description,
			b.Isbn,
			int(search.OReilly),
		}
	}))
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	wg.Add(len(results.Results))

	coverImages := make([]imageModel, 0, len(results.Results))
	mu := &sync.Mutex{}
	for _, result := range results.Results {
		coverUrl, _ := url.Parse(result.CoverUrl)
		chunks := strings.Split(coverUrl.Path, "/")
		go func(result book) {
			defer wg.Done()
			//_ = fileutils.DownloadFile(result.CoverUrl, filepath.Join("covers", fmt.Sprintf("%s.jpg", chunks[len(chunks)-2])))
			bytes, err := fileutils.FetchFile(result.CoverUrl)
			checkSum := md5.Sum(bytes)
			if err == nil {
				mu.Lock()
				coverImages = append(coverImages, imageModel{
					Name:   fmt.Sprintf("%s.jpg", chunks[len(chunks)-2]),
					Image:  bytes,
					Hash:   string(checkSum[:]),
					Source: int(search.OReilly),
				})
				mu.Unlock()
			}
		}(result)
	}

	wg.Wait()

	err = dbutils.BulkInsert[imageModel, imageModel]("images", coverImages)

	return err
}

func searchBooks(query string) (search.SearchResults[book], error) {
	var results search.SearchResults[book]
	res, err := httputils.Fetch(func() (*http.Response, error) {
		return http.Get(fmt.Sprintf("https://learning.oreilly.com/api/v2/search/?query=%s&formats=book", url.QueryEscape(query)))
	})()
	if err != nil {
		return results, err
	}
	resBody, _ := io.ReadAll(res)
	defer res.Close()
	json.Unmarshal(resBody, &results)

	return results, nil
}
