package main

import (
	_ "embed"
	"fmt"
	"jrest/internal"
	"log"
	"os"
)

//go:embed internal/source/example.json
var example string

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Fatalf("an unexpected error occurred: %v", err)
		}
	}()

	filename := "source.json"
	if len(os.Args) == 2 {
		filename = os.Args[1]
	}

	if filename == "--help" || filename == "-h" {
		fmt.Println("\nA utility for serving the content of a json file.  By default a local file 'source.json'")
		fmt.Println("will be served, but this can be replaced with any file name")
		fmt.Printf("Usage: %s source.json\n", os.Args[0])
		fmt.Println("      --help prints this message")
		fmt.Println("      --example creates an example file source.json")
	} else if filename == "--example" || filename == "-e" {
		err := os.WriteFile("example.json", []byte(example), 0644)
		if err != nil {
			log.Fatalf("unable to write example: %v", err)
		}
		fmt.Println("See example.json")
	} else {
		app := internal.NewApp(filename)
		app.Serve()
	}
}
