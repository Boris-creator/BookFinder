package main

import (
	"bookfinder/apress"
	"bookfinder/manning"
	"bookfinder/nostarch"
	"os"
	"sync"

	"bookfinder/oreilly"
	_ "bookfinder/store"
	"fmt"
)

func main() {
	var query string
	if args := os.Args; len(args) > 1 {
		query = args[1]
	}
	if len(query) == 0 {
		return
	}

	var wg sync.WaitGroup
	wg.Add(4)
	go func() {
		defer wg.Done()
		err := apress.Search(query)
		fmt.Println(err)
	}()
	go func() {
		defer wg.Done()
		err := oreilly.SearchAndSave(query)
		fmt.Println(err)
	}()
	go func() {
		defer wg.Done()
		err := nostarch.Search(query)
		fmt.Println(err)
	}()
	go func() {
		defer wg.Done()
		err := manning.Search(query)
		fmt.Println(err)
	}()
	wg.Wait()
}
