package lsbf

import (
	"testing"
	"math"

	"git.broccolimicro.io/Broccoli/pr.git/chp"
	"git.broccolimicro.io/Broccoli/pr.git/chp/bd/stream"
	"git.broccolimicro.io/Broccoli/pr.git/chp/param"
	
	"github.com/stretchr/testify/assert"
)

func TestIntegrationBuffer(t *testing.T) {
	profile := param.String(1, "example.prof")
	out := param.String(2, "test/stream/buffer")
	base := param.Int64(3, int64(16))
	min := param.Int64(4, math.MinInt64)
	max := param.Int64(5, math.MaxInt64)

	g, err := chp.New(out, profile)
	assert.NoError(t, err)
	defer g.Done()

	Ls, Lr := Chan("L", base, 0)
	Rs, Rr := Chan("L", base, 0)

	go chp.SourceN[int64](100, chp.RandomInt64(min, max), g.Sub("src"), Ls)
	go chp.Sink[int64](g.Sub("sink"), Rr)
	go stream.Buffer(g.Sub("dut"), Lr.Raw(), Rs.Raw())
}

func TestIntegrationCopy(t *testing.T) {
	profile := param.String(1, "example.prof")
	out := param.String(2, "test/stream/copy")
	copies := param.Int(3, 2)
	base := param.Int64(4, int64(16))
	min := param.Int64(5, math.MinInt64)
	max := param.Int64(6, math.MaxInt64)

	g, err := chp.New(out, profile)
	assert.NoError(t, err)
	defer g.Done()

	Ls, Lr := Chan("L", base, 0)
	Rs, Rr := ChanArr("R", copies, base, 0)

	for i := 0; i < copies; i++ {
		go chp.Sink[int64](g.Sub("sink.%d", i), Rr[i])
	}
	go chp.SourceN[int64](100, chp.RandomInt64(min, max), g.Sub("src"), Ls)
	go stream.Copy(g.Sub("dut"), Lr.Raw(), RawSenders(Rs))
}

func TestIntegrationSplit(t *testing.T) {
	profile := param.String(1, "example.prof")
	out := param.String(2, "test/stream/split")
	choices := param.Int(3, 2)
	base := param.Int64(4, int64(16))
	min := param.Int64(5, math.MinInt64)
	max := param.Int64(6, math.MaxInt64)

	g, err := chp.New(out, profile)
	assert.NoError(t, err)
	defer g.Done()

	Cs, Cr := chp.Chan[int]("C", 0)
	Ls, Lr := Chan("L", base, 0)
	Rs, Rr := ChanArr("R", choices, base, 0)

	go chp.SourceN(100, chp.RandomInt(0, choices), g.Sub("src_C"), Cs)
	go chp.Source[int64](chp.RandomInt64(min, max), g.Sub("src_L"), Ls)
	for i := 0; i < choices; i++ {
		go chp.Sink[int64](g.Sub("sink.%d", i), Rr[i])
	}
	go stream.Split(g.Sub("dut"), Cr, Lr.Raw(), RawSenders(Rs))
}

func TestIntegrationMerge(t *testing.T) {
	profile := param.String(1, "example.prof")
	out := param.String(2, "test/stream/merge")
	choices := param.Int(3, 2)
	base := param.Int64(4, int64(16))
	min := param.Int64(5, math.MinInt64)
	max := param.Int64(6, math.MaxInt64)

	g, err := chp.New(out, profile)
	assert.NoError(t, err)
	defer g.Done()
	
	Cs, Cr := chp.Chan[int]("C", 0)
	Ls, Lr := ChanArr("L", choices, base, 0)
	Rs, Rr := Chan("R", base, 0)

	go chp.SourceN(100, chp.RandomInt(0, choices), g.Sub("src_C"), Cs)
	for i := 0; i < choices; i++ {
		go chp.Source[int64](chp.RandomInt64(min, max), g.Sub("src_L.%d", i), Ls[i])
	}
	go chp.Sink[int64](g.Sub("sink"), Rr)
	go stream.Merge(g.Sub("dut"), Cr, RawReceivers(Lr), Rs.Raw())
}

