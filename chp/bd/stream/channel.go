package stream

import (
	"pr/chp/bd"
)

type Sender[T interface{}] interface {
	bd.Sender[bool, T]

	SendStream(tokens []T, args ...float64) float64
}

type Receiver[T interface{}] interface {
	bd.Receiver[bool, T]

	RecvStream(args ...float64) ([]T, float64)
}

/******************************
*          Channel            *
******************************/

type sender[T interface{}] struct {
	bd.Sender[bool, T]
}

type receiver[T interface{}] struct {
	bd.Receiver[bool, T]
}

func Chan[T interface{}](name string, slack int64) (Sender[T], Receiver[T]) {
	s, r := bd.Chan[bool, T](name, slack)
	return &sender[T]{s}, &receiver[T]{r}
}

func ChanArr[T interface{}](name string, n int, slack int64) ([]Sender[T], []Receiver[T]) {
	s, r := bd.ChanArr[bool, T](name, n, slack)
	S := make([]Sender[T], n)
	R := make([]Receiver[T], n)
	for i := 0; i < n; i++ {
		S[i] = &sender[T]{s[i]}
		R[i] = &receiver[T]{r[i]}
	}
	return S, R
}

func (self *sender[T]) SendStream(tokens []T, args ...float64) float64 {
	var start float64 = 0.0
	if len(args) > 0 {
		start = args[0]
	}

	var step float64 = 0.0
	if len(args) > 1 {
		step = args[1]
	}

	end := start
	for i, token := range tokens {
		end = self.Sender.SendToken(i == len(tokens)-1, token, start)
		start += step
	}

	return end
}

func (self *receiver[T]) RecvStream(args ...float64) ([]T, float64) {
	var start float64 = 0.0
	if len(args) > 0 {
		start = args[0]
	}

	var step float64 = 0.0
	if len(args) > 1 {
		step = args[1]
	}

	var tokens []T

	var t bd.Token[bool, T]
	end := start
	for !t.C {
		t, end = self.Receiver.Recv(start)
		tokens = append(tokens, t.D)
		start += step
	}
	return tokens, end
}

