package proc

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"math"
	"os"
	"os/exec"
	"strings"

	"github.com/djherbis/buffer"
	"github.com/djherbis/nio/v3"
)

type Proc struct {
	r    io.Reader
	done chan struct{}
	err  error
}

func Cat(s ...any) *Proc {
	var rs []io.Reader
	for _, s := range s {
		switch r := s.(type) {
		case io.Reader:
			rs = append(rs, r)
		case string:
			rs = append(rs, strings.NewReader(r))
		}
	}
	if len(rs) == 0 {
		return &Proc{}
	}
	r, w := nio.Pipe(buffer.New(math.MaxInt64))
	p := &Proc{r: r, done: make(chan struct{})}
	go func() {
		_, p.err = io.Copy(w, io.MultiReader(rs...))
		w.CloseWithError(p.err)
		close(p.done)
	}()
	return p
}

func Cmd(name string, arg ...string) *Proc {
	return Cat().Cmd(name, arg...)
}

func Err(err error) *Proc {
	return &Proc{err: err}
}

func (p *Proc) Cat(s ...any) *Proc {
	return Cat(append([]any{p}, s...)...)
}

func (p *Proc) Cmd(name string, arg ...string) *Proc {
	q := &Proc{}
	c := exec.Command(name, arg...)
	c.Stdin = p.r
	r, w := nio.Pipe(buffer.New(math.MaxInt64))
	q.r, c.Stdout = r, w
	var buf bytes.Buffer
	c.Stderr = &buf
	if q.err = c.Start(); q.err != nil {
		w.CloseWithError(q.err)
		return q
	}
	q.done = make(chan struct{})
	go func() {
		if q.err = p.Err(); q.err == nil {
			q.err = c.Wait()
			if _, ok := q.err.(*exec.ExitError); ok {
				if s := bufio.NewScanner(&buf); s.Scan() {
					q.err = errors.New(s.Text())
				}
			}
		}
		w.CloseWithError(q.err)
		close(q.done)
	}()
	return q
}

func (p *Proc) Err() error {
	if p.done != nil {
		<-p.done
	}
	return p.err
}

func (p *Proc) Map(f func(string) *Proc) *Proc {
	r, w := nio.Pipe(buffer.New(math.MaxInt64))
	q := &Proc{r: r, done: make(chan struct{})}
	ch := make(chan *Proc, 10)
	s := bufio.NewScanner(p)
	go func() {
		for ; s.Scan(); ch <- f(s.Text()) {
		}
		close(ch)
	}()
	go func() {
		for p := <-ch; p != nil; p = <-ch {
			if _, q.err = io.Copy(w, p); q.err != nil {
				break
			}
		}
		if s.Err() != nil {
			q.err = s.Err()
		}
		w.CloseWithError(q.err)
		close(q.done)
	}()
	return q
}

func (p *Proc) Nul() *Proc {
	q := &Proc{done: make(chan struct{})}
	go func() {
		q.err = p.Err()
		close(q.done)
	}()
	return q
}

func (p *Proc) Put(name string) *Proc {
	f, err := os.Create(name)
	q := &Proc{err: err}
	if err != nil {
		return q
	}
	q.done = make(chan struct{})
	go func() {
		_, q.err = io.Copy(f, p)
		if err := f.Close(); q.err == nil {
			q.err = err
		}
		close(q.done)
	}()
	return q
}

func (p *Proc) Read(b []byte) (int, error) {
	switch {
	case p.r != nil:
		return p.r.Read(b)
	case p.err != nil:
		return 0, p.err
	default:
		return 0, io.EOF
	}
}
