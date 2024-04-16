package coro

import (
	"errors"
	"fmt"
)

var ErrCanceled = errors.New("coroutine canceled")

type msg[T any] struct {
	panic any
	val   T
}

func New[In, Out any](f func(In, func(Out) In) Out) (resume func(In) (Out, bool), cancel func()) {
	cin := make(chan msg[In])
	cout := make(chan msg[Out])
	running := true
	resume = func(in In) (out Out, ok bool) {
		if !running {
			return
		}
		cin <- msg[In]{val: in}
		m := <-cout
		if m.panic != nil {
			panic(m.panic)
		}
		return m.val, running
	}
	cancel = func() {
		e := fmt.Errorf("%w", ErrCanceled)
		cin <- msg[In]{panic: e}
		m := <-cout
		if m.panic != nil && m.panic != e {
			panic(m.panic)
		}
	}
	yield := func(out Out) In {
		cout <- msg[Out]{val: out}
		m := <-cin
		if m.panic != nil {
			panic(m.panic)
		}
		return m.val
	}
	go func() {
		defer func() {
			if running {
				running = false
				cout <- msg[Out]{panic: recover()}
			}
		}()
		var out Out
		m := <-cin
		if m.panic == nil {
			out = f(m.val, yield)
		}
		running = false
		cout <- msg[Out]{val: out}
	}()
	return resume, cancel
}
