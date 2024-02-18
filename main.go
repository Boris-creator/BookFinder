package main

import (
	"bookfinder/manning"
	"bookfinder/nostarch"
	"os"

	//"bookfinder/oreilly"
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
	//err := oreilly.SearchAndSave(query)
	res, err := nostarch.Search(query)
	fmt.Println(res, err)
	manningBooks, err := manning.SearchBooks(query)
	fmt.Println(manningBooks, err)
}
