package chp

import (
	"errors"

	"wyrm/chp/timing"
)

func Buffer[T interface{}](g Globals, L Receiver[T], R Sender[T]) {
	p := g.Init(L, R)
	defer g.Done()

	d0L := p.Find("d0L")
	d0R := p.Find("d0R")
	d0 := p.Find("d0")
	e0 := p.Find("e0")

	for {
		x, tl := L.Recv(d0L)
		tr := R.Send(x, tl+d0R)

		g.Cycle(e0, tl, tr+d0)
	}
}

func Copy[T interface{}](g Globals, L Receiver[T], R []Sender[T]) {
	p := g.Init(L, R)
	defer g.Done()

	d0L := p.Find("d0L")
	d0R := p.Find("d0R")
	d0 := p.Find("d0")
	e0 := p.Find("e0")*float64(len(R))

	for {
		x, tl := L.Recv(d0L)
		tr := timing.Set(tl)
		for _, r := range R { 
			tr.Max(r.Send(x, tl+d0R))
		}

		g.Cycle(e0, tl, tr.Get()+d0)
	}
}

func Split[T interface{}](g Globals, C Receiver[int], L Receiver[T], R []Sender[T]) {
	p := g.Init(C, L, R)
	defer g.Done()

	d0C := p.Find("d0C")
	d0L := p.Find("d0L")
	d0R := p.Find("d0R")
	d0 := p.Find("d0")
	e0 := p.Find("e0")

	for {
		c, tc := C.Recv(d0C)
		if c < 0 || c >= len(R) {
			panic(errors.New("split control channel out of bounds"))
		}
		x, tl := L.Recv(d0L)
		t := timing.Max(tc, tl)
		tr := R[c].Send(x, t+d0R)

		g.Cycle(e0, t, tr+d0)
	}
}

func Merge[T interface{}](g Globals, C Receiver[int], L []Receiver[T], R Sender[T]) {
	p := g.Init(C, L, R)
	defer g.Done()

	d0C := p.Find("d0C")
	d0L := p.Find("d0L")
	d0R := p.Find("d0R")
	d0 := p.Find("d0")
	e0 := p.Find("e0")

	for {
		c, tc := C.Recv(d0C)
		if c < 0 || c >= len(L) {
			panic(errors.New("merge control channel out of bounds"))
		}
		x, tl := L[c].Recv(d0L)
		t := timing.Max(tc, tl)
		tr := R.Send(x, t+d0R)

		g.Cycle(e0, t, tr+d0)
	}
}

func Source[T interface{}](fn func(i int64) T, g Globals, R ...Sender[T]) {
	p := g.Init(R)
	defer g.Done()

	d0 := p.Find("d0")
	e0 := p.Find("e0")*float64(len(R))

	for i := int64(0); ; i++ {
		value := fn(i)
		t := timing.Set()
		for j := 0; j < len(R); j++ {
			t.Max(R[j].Send(value))
		}
		g.Cycle(e0, t.Get(), t.Get()+d0)
	}
}

func SourceN[T interface{}](n int64, fn func(i int64) T, g Globals, R ...Sender[T]) {
	p := g.Init(R)
	defer g.Done()

	d0 := p.Find("d0")
	e0 := p.Find("e0")*float64(len(R))

	for i := int64(0); i < n; i++ {
		value := fn(i)
		t := timing.Set()
		for j := 0; j < len(R); j++ {
			t.Max(R[j].Send(value))
		}
		g.Cycle(e0, t.Get(), t.Get()+d0)
	}
}

func Sink[T interface{}](g Globals, L Receiver[T]) {
	p := g.Init(L)
	defer g.Done()

	d0 := p.Find("d0")
	e0 := p.Find("e0")

	for {
		_, tl := L.Recv()
		g.Cycle(e0, tl, tl+d0)
	}
}

func SinkN[T interface{}](n int64, g Globals, L Receiver[T]) {
	p := g.Init(L)
	defer g.Done()

	d0 := p.Find("d0")
	e0 := p.Find("e0")

	for i := int64(0); i < n; i++ {
		_, tl := L.Recv()
		g.Cycle(e0, tl, tl+d0)
	}
}

func SinkAndCheck[T interface{}](valid func(token int64, values []T) error, g Globals, L ...Receiver[T]) {
	p := g.Init(L)
	defer g.Done()

	d0 := p.Find("d0")
	e0 := p.Find("e0")

	var tl float64
	values := make([]T, len(L))
	for i := int64(0); ; i++ {
		t := timing.Set()
		for j := 0; j < len(L); j++ {
			values[j], tl = L[j].Recv()
			t.Max(tl)
		}
		err := valid(i, values)
		if err != nil {
			panic(err)
		}
		g.Cycle(e0, t.Get(), t.Get()+d0)
	}
}

func SinkAndCheckN[T interface{}](n int64, valid func(token int64, values []T) error, g Globals, L ...Receiver[T]) {
	p := g.Init(L)
	defer g.Done()

	d0 := p.Find("d0")
	e0 := p.Find("e0")

	var tl float64
	values := make([]T, len(L))
	for i := int64(0); i < n; i++ {
		t := timing.Set()
		for j := 0; j < len(L); j++ {
			values[j], tl = L[j].Recv()
			t.Max(tl)
		}
		err := valid(i, values)
		if err != nil {
			panic(err)
		}
		g.Cycle(e0, t.Get(), t.Get()+d0)
	}
}
