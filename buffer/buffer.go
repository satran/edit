package buffer

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

// buffer is a piece table implementation that serves the purpose of editing text

type Buffer struct {
	// the original content
	r io.ReadSeeker
	// all changes are made here
	w io.ReadWriteSeeker
	// the table represents the current contents
	t *table

	length int64
	offset int64
}

func New(r io.ReadSeeker, w io.ReadWriteSeeker, length int64) *Buffer {
	return &Buffer{
		r:      r,
		w:      w,
		t:      newTable(r, length),
		length: length,
	}
}

func NewWithFile(filename string) (*Buffer, func(), error) {
	r, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	stat, err := os.Stat(os.Args[1])
	if err != nil {
		return nil, nil, err
	}

	w, err := ioutil.TempFile("", "edi")
	if err != nil {
		return nil, nil, err
	}

	b := New(r, w, stat.Size())
	return b, func() {
		r.Close()
		w.Close()
	}, nil
}

type table struct {
	start *row
}

func newTable(r io.ReadSeeker, length int64) *table {
	return &table{
		start: &row{
			r:      r,
			start:  0,
			length: length,
		},
	}
}

func (t *table) insert(offset int64, rd io.ReadSeeker, start, length int64) {
	var current int64 = 0
	r := t.start
	if r.next != nil && r.next.start == 0 && r.next.length == 0 {
		r.r = rd
		r.start = start
		r.length = length
		return
	}
	for offset > current+r.length || r.next != nil {
		current += r.length
		r = r.next
	}
	pos := offset - current
	x, y := r.split(pos)
	nr := row{
		r:      rd,
		start:  start,
		length: length,
		next:   &y,
		prev:   r,
	}
	r.start = x.start
	r.length = x.length
	r.next = &nr
}

type row struct {
	r      io.ReadSeeker
	start  int64
	length int64

	next *row
	prev *row
}

func (r *row) split(pos int64) (row, row) {
	x := row{
		r:      r.r,
		start:  r.start,
		length: r.start + pos,
		prev:   r.prev,
	}

	y := row{
		r:      r.r,
		start:  pos,
		length: r.length - pos,
		next:   r.next,
	}
	return x, y
}

func (b *Buffer) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekCurrent:
		b.offset += offset
	case io.SeekStart:
		b.offset = offset
	case io.SeekEnd:
		b.offset += offset
	default:
		return -1, errors.New("unknown position")
	}
	if b.offset > b.length {
		b.offset = b.length
	}
	return b.offset, nil
}

func (b *Buffer) Write(p []byte) (int, error) {
	start, err := b.w.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}
	_, err = b.w.Write(p)
	if err != nil {
		return 0, err
	}
	l := int64(len(p))
	b.t.insert(b.offset, b.w, start, l)
	b.offset += l
	b.length += l
	return 0, nil
}

func (b *Buffer) Read(p []byte) (n int, err error) {
	r := b.t.start
	read := 0
	for {
		toread := len(p) - read
		l := int(r.length)
		if toread > l {
			toread = l
		}
		_, err := r.r.Seek(r.start, io.SeekStart)
		if err != nil {
			return 0, err
		}
		cont := p[read : read+toread]
		n, err := r.r.Read(cont)
		if err != nil {
			return 0, err
		}
		if n != toread {
			return 0, fmt.Errorf("expected to read %d got %d", toread, n)
		}
		read += toread
		if read == len(p) {
			return read, nil
		}
		if r.next == nil {
			break
		}
		r = r.next
	}
	return read, io.EOF
}
