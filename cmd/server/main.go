package main

import (
	"log"
	"os"
)

func main() {
	if err := runServer(); err != nil {
		log.Printf("something went wrong: %v", err)
		os.Exit(1)
	}
}
