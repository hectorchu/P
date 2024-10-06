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
		case []byte:
			rs = append(rs, bytes.NewReader(r))
		case string:
			rs = append(rs, strings.NewReader(r))
		}
	}
	return Fun(func(w io.Writer) error {
		_, err := io.Copy(w, io.MultiReader(rs...))
		return err
	})
}

func Cmd(name string, arg ...string) *Proc {
	return Cat().Cmd(name, arg...)
}

func Err(err error) *Proc {
	return &Proc{err: err}
}

func Fun(f func(io.Writer) error) *Proc {
	r, w := nio.Pipe(buffer.New(math.MaxInt64))
	p := &Proc{r: r, done: make(chan struct{})}
	go func() {
		p.err = f(w)
		w.CloseWithError(p.err)
		close(p.done)
	}()
	return p
}

func (p *Proc) Cat(s ...any) *Proc {
	return Cat(append([]any{p}, s...)...)
}

func (p *Proc) Cmd(name string, arg ...string) *Proc {
	return Fun(func(w io.Writer) (err error) {
		var buf bytes.Buffer
		c := exec.Command(name, arg...)
		c.Stdin, c.Stdout, c.Stderr = p.r, w, &buf
		if err = c.Start(); err != nil {
			return
		}
		if err = p.Err(); err != nil {
			return
		}
		err = c.Wait()
		if _, ok := err.(*exec.ExitError); !ok {
			return
		}
		if s := bufio.NewScanner(&buf); s.Scan() {
			err = errors.New(s.Text())
		}
		return
	})
}

func (p *Proc) Err() error {
	if p.done != nil {
		<-p.done
	}
	return p.err
}

func (p *Proc) Map(f func(string) *Proc) *Proc {
	ch := make(chan *Proc, 10)
	s := bufio.NewScanner(p)
	go func() {
		for ; s.Scan(); ch <- f(s.Text()) {
		}
		close(ch)
	}()
	return Fun(func(w io.Writer) (err error) {
		for p := <-ch; p != nil; p = <-ch {
			if _, err = io.Copy(w, p); err != nil {
				break
			}
		}
		if s.Err() != nil {
			err = s.Err()
		}
		return
	})
}

func (p *Proc) Nul() *Proc {
	return Fun(func(io.Writer) error {
		_, err := io.Copy(io.Discard, p)
		return err
	})
}

func (p *Proc) Put(name string) *Proc {
	return Fun(func(w io.Writer) (err error) {
		f, err := os.Create(name)
		if err != nil {
			return
		}
		_, err = io.Copy(f, p)
		if err2 := f.Close(); err == nil {
			err = err2
		}
		return
	})
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
