package chp

import (
	"time"
	"testing"
	"github.com/stretchr/testify/assert"

	"git.broccolimicro.io/Broccoli/pr.git/chp/param"
)

func TestUnitSendRecv(t *testing.T) {
	profile := param.String(1, "example.prof")
	out := param.String(2, "test/chp/channel")

	g, err := New(out, profile)
	assert.NoError(t, err)
	defer g.Done()
	g.RandomTiming(time.Millisecond)
	g.SetDebug(true)
	
	Cs, Cr := Chan[int64]("L", 0)

	Cs.SetGlobals(g)
	Cr.SetGlobals(g)

	for i := 0; i < 10; i++ {
		go func(g Globals) {
			g.Init()
			defer g.Done()

			for j := 0; j < 100; j++ {
				Cs.Send(0)
			}
		}(g.Sub("src[%d]", i))
		go func(g Globals) {
			g.Init()
			defer g.Done()

			for j := 0; j < 100; j++ {
				Cr.Recv()
			}
		}(g.Sub("sink[%d]", i))
	}
}

