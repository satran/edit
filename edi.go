package main

import (
	"log"
	"os"

	termbox "github.com/nsf/termbox-go"
)

func main() {
	e, err := NewEditor(os.Args[1:]...)
	if err != nil {
		log.Fatal(err)
	}
	defer e.Close()

	if err := termbox.Init(); err != nil {
		log.Fatal(err)
	}
	defer termbox.Close()

	if err := e.Run(); err != nil {
		log.Fatal(err)
	}
}
