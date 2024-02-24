package main

import (
	"bookfinder/manning"
	"bookfinder/nostarch"
	"os"

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
	oreillyErr := oreilly.SearchAndSave(query)
	fmt.Println(oreillyErr)
	err := nostarch.Search(query)
	fmt.Println(err)
	err = manning.Search(query)
	fmt.Println(err)
}
