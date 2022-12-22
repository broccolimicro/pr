package chp

import (
	"sync"
	"fmt"
	"reflect"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"pr/chp/timing"
)

type Globals interface {
	Sub(name string, args ...any) Globals
	Init(args ...interface{}) timing.Profile
	Done()
	Cycle(fJ, start, end float64)

	Name() string
	Dir() string
	Curr() float64

	SetDebug(debug bool)
	Debug() bool
}

type cycle struct {
	start float64
	end float64
	fJ float64
}

type globals struct {
	name string
	parent *globals
	children []*globals
	
	wg *sync.WaitGroup
	ports []interface{}
	curr float64

	// timing profile
	t timing.ProfileSet

	// cycle logger
	log *os.File
	dir string

	debug bool
}

func New(args ...string) (Globals, error) {
	dir := "run"
	if len(args) > 0 {
		dir = args[0]
	}

	var err error
	var t timing.ProfileSet
	if len(args) > 1 && args[1] != "" {
		prof := args[1]
		t, err = timing.LoadProfileSet(prof)
		if err != nil {
			return nil, err
		}
	} else {
		t = timing.NewProfileSet()
	}

	name := "top"
	if len(args) > 2 {
		name = args[2]
	}

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, err
	}

	return &globals{
		name: name,
		dir: dir,
		wg: &sync.WaitGroup{},
		t: t,
	}, nil
}

func (g *globals) Sub(name string, args ...any) Globals {
	g.wg.Add(1)

	name = g.name + "." + fmt.Sprintf(name, args...)

	child := &globals {
		name: name,
		dir: g.dir,
		parent: g,
		wg: &sync.WaitGroup{},
		debug: g.debug,
		t: g.t,
	}
	g.children = append(g.children, child)
	return child
}

func caller(skipFrames int) string {
	pc, _, _, ok := runtime.Caller(skipFrames+2)
	details := runtime.FuncForPC(pc)
	if ok && details != nil {
		return details.Name()
	}
	return ""
}

func (g *globals) Init(args ...interface{}) timing.Profile {
	g.ports = args
	for _, port := range g.ports {
		if c, ok := port.(Recordable); ok {
			c.SetGlobals(g)
		} else if reflect.TypeOf(port).Kind() == reflect.Slice || reflect.TypeOf(port).Kind() == reflect.Array {
			items := reflect.ValueOf(port)
			for i := 0; i < items.Len(); i++ {
				item := items.Index(i).Interface()
				if c, ok := item.(Recordable); ok {
					c.SetGlobals(g)
				}
			}
		}
	}

	if g.t != nil {
		p := g.t.Find(g.name)
		if p != nil {
			return p
		}

		p = g.t.Find(caller(0))
		if p != nil {
			return p
		}
	}

	return timing.NewProfile()
}

func (g *globals) Done() {
	if len(g.children) > 0 {
		g.wg.Wait()
	}

	if r := recover(); r != nil && r != Deadlock {
		panic(r)
	}
	for _, port := range g.ports {
		if c, ok := port.(io.Closer); ok {
			err := c.Close()
			if err != nil {
				fmt.Println(err)
			}
		} else if reflect.TypeOf(port).Kind() == reflect.Slice || reflect.TypeOf(port).Kind() == reflect.Array {
			items := reflect.ValueOf(port)
			for i := 0; i < items.Len(); i++ {
				item := items.Index(i).Interface()
				if c, ok := item.(io.Closer); ok {
					err := c.Close()
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}
	}

	if g.log != nil {	
		g.log.Close()
	}

	if g.debug {
		fmt.Printf("deadlock %s\n", g.name)
	}

	if g.parent != nil {
		g.parent.wg.Done()
	}
}

func (g *globals) Cycle(fJ, start, end float64) {
	if g.log == nil {
		var err error
		g.log, err = os.Create(filepath.Join(g.dir, g.name))
		if err != nil {
			fmt.Println(err)
		}
	}

	if g.log != nil {
		fmt.Fprintf(g.log, "%f\t%f\t%f\n", g.curr+start, g.curr+end, fJ)
	}
	g.curr += end
}

func (g *globals) Name() string {
	return g.name
}

func (g *globals) Dir() string {
	return g.dir
}

func (g *globals) Curr() float64 {
	return g.curr
}

func (g *globals) SetDebug(debug bool) {
	g.debug = debug
	for _, child := range g.children {
		child.SetDebug(debug)
	}
}

func (g *globals) Debug() bool {
	return g.debug
}
