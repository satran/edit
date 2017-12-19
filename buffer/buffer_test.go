package buffer

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"
)

func TestRowSplit(t *testing.T) {
	tests := []struct{ start, length, pos, xstart, xlength, ystart, ylength int64 }{
		{0, 6, 3, 0, 3, 3, 3},
	}
	for i, test := range tests {
		r := row{
			start:  test.start,
			length: test.length,
		}
		x, y := r.split(test.pos)
		if exp, got := test.xstart, x.start; exp != got {
			t.Errorf("%d failed xstart:  exp: %d, got: %d", i, exp, got)
		}
		if exp, got := test.ystart, y.start; exp != got {
			t.Errorf("%d failed ystart:  exp: %d, got: %d", i, exp, got)
		}
		if exp, got := test.xlength, x.length; exp != got {
			t.Errorf("%d failed xlength:  exp: %d, got: %d", i, exp, got)
		}
		if exp, got := test.ylength, y.length; exp != got {
			t.Errorf("%d failed ylength:  exp: %d, got: %d", i, exp, got)
		}
	}
}

func TestBuffer(t *testing.T) {
	tests := []struct {
		initial, text, exp string
		pos                int64
	}{
		{"hello world", " new", "hello new world", 5},
		{"the quick brownjumps over the lazy dog", " fox ", "the quick brown fox jumps over the lazy dog", 15},
	}
	for i, test := range tests {
		r := bytes.NewReader([]byte(test.initial))
		w, err := ioutil.TempFile("/dev/shm", "")
		if err != nil {
			t.Fatalf("%d: %s", i, err)
		}
		defer w.Close()
		b, err := New(r, w, int64(len([]byte(test.initial))))
		if err != nil {
			t.Fatalf("%d: %s", i, err)
		}
		_, err = b.Seek(test.pos, io.SeekStart)
		if err != nil {
			t.Fatalf("%d: %s", i, err)
		}
		_, err = b.Write([]byte(test.text))
		if err != nil {
			t.Fatalf("%d: %s", i, err)
		}
		cont, err := ioutil.ReadAll(b)
		if err != nil {
			t.Fatalf("%d: %s", i, err)
		}
		if string(cont) != test.exp {
			t.Fatalf("%d: exp: %s\ngot:%s", i, test.exp, string(cont))
		}
	}
}
