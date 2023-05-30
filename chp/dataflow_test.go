package chp

import (
	"testing"
	"github.com/stretchr/testify/assert"

	"git.broccolimicro.io/Broccoli/pr.git/chp/param"
)

func TestIntegrationConnect(t *testing.T) {
	profile := param.String(1, "example.prof")
	out := param.String(2, "test/chp/buffer")
	min := param.Int64(3, int64(0))
	max := param.Int64(4, int64(2))

	g, err := New(out, profile)
	assert.NoError(t, err)
	defer g.Done()
	
	Ls, Lr := Chan[int64]("L", 0)
	Vs, Vr := Chan[int64]("V", 2)
	Rs, Rr := Chan[int64]("R", 0)

	go SourceN(100, RandomInt64(min, max), g.Sub("src"), Ls, Vs)
	go SinkAndCheck(AreEqual[int64], g.Sub("sink"), Vr, Rr)
	go Connect(g.Sub("dut"), Lr, Rs)
}

func TestIntegrationBuffer(t *testing.T) {
	profile := param.String(1, "example.prof")
	out := param.String(2, "test/chp/buffer")
	min := param.Int64(3, int64(0))
	max := param.Int64(4, int64(2))

	g, err := New(out, profile)
	assert.NoError(t, err)
	defer g.Done()
	
	Ls, Lr := Chan[int64]("L", 0)
	Vs, Vr := Chan[int64]("V", 2)
	Rs, Rr := Chan[int64]("R", 0)

	go SourceN(100, RandomInt64(min, max), g.Sub("src"), Ls, Vs)
	go SinkAndCheck(AreEqual[int64], g.Sub("sink"), Vr, Rr)
	go Buffer(g.Sub("dut"), Lr, Rs)
}

func TestIntegrationCopy(t *testing.T) {
	profile := param.String(1, "example.prof")
	out := param.String(2, "test/chp/copy")
	copies := param.Int(3, 2)
	min := param.Int64(4, int64(0))
	max := param.Int64(5, int64(2))

	g, err := New(out, profile)
	assert.NoError(t, err)
	defer g.Done()

	Ls, Lr := Chan[int64]("L", 0)
	Rs, Rr := ChanArr[int64]("R", copies, 0)

	go SourceN[int64](100, RandomInt64(min, max), g.Sub("src"), Ls)
	for i := 0; i < copies; i++ {
		go Sink[int64](g.Sub("sink.%d", i), Rr[i])
	}
	go Copy[int64](g.Sub("dut"), Lr, Rs)
}

func TestIntegrationSplit(t *testing.T) {
	profile := param.String(1, "example.prof")
	out := param.String(2, "test/chp/split")
	choices := param.Int(3, 2)
	min := param.Int64(4, int64(0))
	max := param.Int64(5, int64(2))

	g, err := New(out, profile)
	assert.NoError(t, err)
	defer g.Done()

	Cs, Cr := Chan[int]("C", 0)
	Ls, Lr := Chan[int64]("L", 0)
	Rs, Rr := ChanArr[int64]("R", choices, 0)

	go SourceN[int](100, RandomInt(0, choices), g.Sub("src_C"), Cs)
	go Source[int64](RandomInt64(min, max), g.Sub("src"), Ls)
	for i := 0; i < choices; i++ {
		go Sink[int64](g.Sub("sink.%d", i), Rr[i])
	}
	go Split[int64](g.Sub("dut"), Cr, Lr, Rs)
}

func TestIntegrationMerge(t *testing.T) {
	profile := param.String(1, "example.prof")
	out := param.String(2, "test/chp/merge")
	choices := param.Int(3, 2)
	min := param.Int64(4, int64(0))
	max := param.Int64(5, int64(2))

	g, err := New(out, profile)
	assert.NoError(t, err)
	defer g.Done()

	Cs, Cr := Chan[int]("C", 0)
	Ls, Lr := ChanArr[int64]("L", choices, 0)
	Rs, Rr := Chan[int64]("R", 0)

	go SourceN[int](100, RandomInt(0, choices), g.Sub("src_C"), Cs)
	for i := 0; i < choices; i++ {
		go Source[int64](RandomInt64(min, max), g.Sub("src.%d", i), Ls[i])
	}
	go Sink[int64](g.Sub("sink"), Rr)
	go Merge[int64](g.Sub("dut"), Cr, Lr, Rs)
}

