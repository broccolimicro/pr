package bd

import "pr/chp"

type Token[ctype, dtype interface{}] struct {
	C ctype
	D dtype
}

type Sender[ctype, dtype interface{}] interface {
	chp.Sender[Token[ctype, dtype]]

	SendToken(c ctype, d dtype, args ...float64) float64
}

type Receiver[ctype, dtype interface{}] interface {
	chp.Receiver[Token[ctype, dtype]]
}

/******************************
*           Channel           *
******************************/

type sender[ctype, dtype interface{}] struct {
	chp.Sender[Token[ctype, dtype]]
}

type receiver[ctype, dtype interface{}] struct {
	chp.Receiver[Token[ctype, dtype]]
}

func Chan[ctype, dtype interface{}](name string, slack int64) (Sender[ctype, dtype], Receiver[ctype, dtype]) {
	s, r := chp.Chan[Token[ctype, dtype]](name, slack)
	return &sender[ctype, dtype]{s}, &receiver[ctype, dtype]{r}
}

func ChanArr[ctype, dtype interface{}](name string, n int, slack int64) ([]Sender[ctype, dtype], []Receiver[ctype, dtype]) {
	s, r := chp.ChanArr[Token[ctype, dtype]](name, n, slack)
	S := make([]Sender[ctype, dtype], n)
	R := make([]Receiver[ctype, dtype], n)
	for i := 0; i < n; i++ {
		S[i] = &sender[ctype, dtype]{s[i]}
		R[i] = &receiver[ctype, dtype]{r[i]}
	}
	return S, R
}

func (self *sender[ctype, dtype]) SendToken(c ctype, d dtype, args ...float64) float64 {
	return self.Sender.Send(Token[ctype, dtype]{c, d}, args...)
}
