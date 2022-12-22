package stream

import (
	"errors"

	"git.broccolimicro.io/Broccoli/pr.git/chp"
	"git.broccolimicro.io/Broccoli/pr.git/chp/timing"
	"git.broccolimicro.io/Broccoli/pr.git/chp/bd"
)

func Buffer[T interface{}](g chp.Globals, L Receiver[T], R Sender[T]) {
	p := g.Init(L, R)
	defer g.Done()

	d0R := p.Find("d0R")
	d0 := p.Find("d0")
	e0 := p.Find("e0")

	for {
		x, tl := L.Recv()
		tr := R.Send(x, tl+d0R)
		g.Cycle(e0, tl, tr+d0)
	}
}

func Copy[T interface{}](g chp.Globals, L Receiver[T], R []Sender[T]) {
	p := g.Init(L, R)
	defer g.Done()

	d0R := p.Find("d0R")
	d0 := p.Find("d0")
	e0 := p.Find("e0")*float64(len(R))

	for {
		x, tl := L.Recv()
		tr := timing.Set()
		for _, r := range R {
			tr.Max(r.Send(x, tl+d0R))
		}
		g.Cycle(e0, tl, tr.Get()+d0)
	}
}

func Split[T interface{}](g chp.Globals, C chp.Receiver[int], L Receiver[T], R []Sender[T]) {
	p := g.Init(C, L, R)
	defer g.Done()

	d0R := p.Find("d0R")
	d0 := p.Find("d0")
	e0 := p.Find("e0")

	for {
		c, tc := C.Recv()
		if c < 0 || c >= len(R) {
			panic(errors.New("split control channel out of bounds"))
		}

		var tl, tr float64
		var x bd.Token[bool, T]

		tr = tc
		for !x.C {
			tc = tr
			x, tl = L.Recv(tr)
			tr = R[c].Send(x, tl+d0R)
			
			g.Cycle(e0, tc, tr+d0)
			tr = 0
		}
	}
}

func Merge[T interface{}](g chp.Globals, C chp.Receiver[int], L []Receiver[T], R Sender[T]) {
	p := g.Init(C, L, R)
	defer g.Done()

	d0R := p.Find("d0R")
	d0 := p.Find("d0")
	e0 := p.Find("e0")

	for {
		c, tc := C.Recv()
		if c < 0 || c >= len(L) {
			panic(errors.New("merge control channel out of bounds"))
		}

		var tl, tr float64
		var x bd.Token[bool, T]

		tr = tc
		for !x.C {
			tc = tr
			x, tl = L[c].Recv(tr)
			tr = R.Send(x, tl+d0R)

			g.Cycle(e0, tc, tr+d0)
			tr = 0
		}
	}
}

func SerialToParallel[T interface{}](g chp.Globals, A Receiver[T], S []Sender[T]) {
	p := g.Init(A, S)
	defer g.Done()

	d0R := p.Find("d0R")
	d0 := p.Find("d0")
	e0 := p.Find("e0")

	i := 0
	for {
		a, ta := A.Recv()
		ts := ta
		if i < len(S) {
			ts = S[i].Send(a, ta+d0R)
		}

		if a.C {
			i = 0
		} else {
			i += 1
		}

		g.Cycle(e0, ta, ts+d0)
	}
}

func ParallelToSerial[T interface{}](g chp.Globals, A []Receiver[T], S Sender[T]) {
	p := g.Init(A, S)
	defer g.Done()

	d0R := p.Find("d0R")
	d0 := p.Find("d0")
	e0 := p.Find("e0")

	i := 0
	for {
		a, ta := A[i].Recv()
		ts := S.SendToken(a.C || i == len(A)-1, a.D, ta+d0R)

		if a.C {
			i = 0
		} else {
			i += 1
		}
		
		g.Cycle(e0, ta, ts+d0)
	}
}


