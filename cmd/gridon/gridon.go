package main

import (
	"fmt"
	"log"

	"gitlab.com/tsuchinaga/gridon"
)

func main() {
	fmt.Println("こんにちわーるど")
	service, err := gridon.NewService()
	if err != nil {
		log.Fatalln(err)
	}
	log.Fatalln(service.Start())
}
