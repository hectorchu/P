package proc

import (
	"bytes"
	"io"
	"sync"
)

type pipe struct {
	mu  sync.Mutex
	buf bytes.Buffer
	ch  chan bool
	err error
}

func newPipe() (io.Reader, *pipe) {
	p := &pipe{ch: make(chan bool, 1)}
	return p, p
}

func (p *pipe) notify() {
	select {
	case p.ch <- true:
	default:
	}
}

func (p *pipe) Read(b []byte) (n int, err error) {
	for {
		<-p.ch
		p.mu.Lock()
		if n, err = p.buf.Read(b); err != nil {
			if err = p.err; err == nil {
				p.mu.Unlock()
				continue
			}
		} else if p.err == nil {
			p.notify()
		}
		p.mu.Unlock()
		return
	}
}

func (p *pipe) Write(b []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.err != nil {
		return 0, p.err
	}
	p.notify()
	return p.buf.Write(b)
}

func (p *pipe) CloseWithError(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	close(p.ch)
	if err == nil {
		err = io.EOF
	}
	p.err = err
}
