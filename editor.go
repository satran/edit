package main

import (
	"bufio"
	"io"
	"log"
	"os"

	termbox "github.com/nsf/termbox-go"
	"github.com/satran/edi/buffer"
)

type Editor struct {
	buffers []*Buffer
	current *Buffer
}

func NewEditor(files ...string) (*Editor, error) {
	e := &Editor{}
	for _, name := range files {
		b, closer, err := buffer.NewWithFile(name)
		if err != nil {
			return nil, err
		}
		buf := &Buffer{b, closer, name}
		e.buffers = append(e.buffers, buf)
		e.current = buf
	}
	return e, nil
}

func (e *Editor) Close() {
	for _, b := range e.buffers {
		b.close()
	}
}

func (e *Editor) Run() error {
	if err := render(e.current); err != nil {
		return err
	}
	if err := e.listen(); err != nil {
		return err
	}
	return nil
}

func render(b *Buffer) error {
	renderHeader(b)
	w, h := termbox.Size()
	r := bufio.NewReader(b.b)
	x, y := 0, 1
	for {
		if x == w {
			x = 0
			y++
		}
		if y == h {
			break
		}
		ch, _, err := r.ReadRune()
		if err != nil && err != io.EOF {
			return err
		}
		if ch == '\n' {
			x = 0
			y++
			continue
		}
		if ch == '\t' {
			for i := 0; i < 8; i++ {
				termbox.SetCell(x, y, ' ', termbox.ColorDefault, termbox.ColorDefault)
				x++
			}
		}
		termbox.SetCell(x, y, ch, termbox.ColorDefault, termbox.ColorDefault)
		x++
		if err == io.EOF {
			break
		}
	}
	termbox.SetCursor(0, 1)
	termbox.Flush()
	return nil
}

func renderHeader(b *Buffer) {
	var (
		x int
		r rune
	)
	w, _ := termbox.Size()
	for x, r = range b.filename {
		termbox.SetCell(x, 0, r, termbox.ColorWhite, termbox.ColorCyan)
	}
	for i := x + 1; i < w; i++ {
		termbox.SetCell(i, 0, ' ', termbox.ColorWhite, termbox.ColorCyan)
	}
}

func (e *Editor) listen() error {
	for {
		ev := termbox.PollEvent()
		switch ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyCtrlQ:
				return nil
			}
		}
	}
}

type Buffer struct {
	b        *buffer.Buffer
	close    func()
	filename string
}

var lg *log.Logger

func init() {
	f, err := os.OpenFile("log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	lg = log.New(f, "", 0)
}
