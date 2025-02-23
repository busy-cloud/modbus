package internal

import (
	"fmt"
	"time"
)

type Token struct {
	C        chan struct{}
	err      error
	request  []byte
	response []byte
}

func NewToken() *Token {
	return &Token{
		C: make(chan struct{}),
	}
}

func (t *Token) Put(data []byte, err error) {
	t.response = data
	t.err = err
	t.C <- struct{}{}
}

func (t *Token) WaitBk() ([]byte, error) {
	<-t.C
	return t.response, t.err
}

func (t *Token) Wait() ([]byte, error) {
	return t.WaitTimeout(time.Second * 5)
}

func (t *Token) Close() {
	close(t.C)
}

func (t *Token) WaitTimeout(dur time.Duration) ([]byte, error) {
	select {
	case <-t.C:
	case <-time.After(dur):
		t.err = fmt.Errorf("token timeout")
		t.Close()
	}
	return t.response, t.err
}
