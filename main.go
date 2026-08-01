package main

import (
	"fmt"
	"log"

	"github.com/frogdevops/gator/internal/config"
)

func main() {
	data, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(data)
}
