package main

import (
	"log"

	"github.com/fumorin/gdl-manager/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
