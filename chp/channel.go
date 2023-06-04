package chp

import (
	"sync"
	"fmt"
	"io"
	"strconv"
	"os"
	"path/filepath"
	"reflect"
	
	"git.broccolimicro.io/Broccoli/pr.git/chp/timing"
)

type Void struct {}

var Null = Void{}

type Recordable interface {
	SetGlobals(g Globals)
}

type Sender[T interface{}] interface {
	io.Closer
	Recordable

	// args are optional timing parameters

	// Non-blocking Send
	Offer(value T, args ...float64) timing.Signal
	
	// Blocking Send
	Send(value T, args ...float64) float64
	
	// Non-blocking Probe
	Watch(args ...float64) timing.Signal
	Ready() bool

	// Blocking Probe
	Wait(args ...float64) float64
}

type Receiver[T interface{}] interface {
	io.Closer
	Recordable
	
	// args are optional timing parameters

	// Non-blocking Receive
	Expect(args ...float64) timing.Action[T]

	// Blocking Receive
	Recv(args ...float64) (T, float64)

	// Non-blocking Probe
	Read(args ...float64) timing.Action[T]
	Valid() bool

	// Blocking Probe
	Probe(args ...float64) (T, float64)
	Wait(args ...float64) float64
}

type Waiter interface{
	Wait(args ...float64) float64
}

type Channel[T interface{}] struct {
	S Sender[T]
	R Receiver[T]
}

func (b Channel[T]) SetGlobals(g Globals) {
	b.S.SetGlobals(g)
	b.R.SetGlobals(g)
}

func Senders[T interface{}](B ...interface{}) (s []Sender[T]) {
	for _, b := range B {
		if bi, ok := b.(Channel[T]); ok {
			s = append(s, bi.S)
		} else if bi, ok := b.(Sender[T]); ok {
			s = append(s, bi)
		} else if reflect.TypeOf(b).Kind() == reflect.Slice || reflect.TypeOf(b).Kind() == reflect.Array {
			items := reflect.ValueOf(b)
			for i := 0; i < items.Len(); i++ {
				item := items.Index(i).Interface()
				if bi, ok := item.(Channel[T]); ok {
					s = append(s, bi.S)
				} else if bi, ok := item.(Sender[T]); ok {
					s = append(s, bi)
				} else {
					panic(Misconfigured)
				}
			}
		} else {
			panic(Misconfigured)
		}
	}
	return
}

func Receivers[T interface{}](B ...interface{}) (r []Receiver[T]) {
	for _, b := range B {
		if bi, ok := b.(Channel[T]); ok {
			r = append(r, bi.R)
		} else if bi, ok := b.(Receiver[T]); ok {
			r = append(r, bi)
		} else if reflect.TypeOf(b).Kind() == reflect.Slice || reflect.TypeOf(b).Kind() == reflect.Array {
			items := reflect.ValueOf(b)
			for i := 0; i < items.Len(); i++ {
				item := items.Index(i).Interface()
				if bi, ok := item.(Channel[T]); ok {
					r = append(r, bi.R)
				} else if bi, ok := item.(Receiver[T]); ok {
					r = append(r, bi)
				} else {
					panic(Misconfigured)
				}
			}
		} else {
			panic(Misconfigured)
		}
	}
	return
}

func Recover[T interface{}](c chan T) {
	r := recover()
	if r == timing.Deadlock {
		close(c)
	} else if r != nil {
		panic(r)
	}
}

func OnAction[T interface{}](op func(args ...float64) (T, float64), args ...float64) timing.Action[T] {
	var send timing.Action[T] = make(chan timing.Value[T], 1)

	go func() {
		defer Recover(send)
		v, t := op(args...)
		send <- timing.Value[T]{t, v}
	}()

	return send
}

func OnSignal(op func(args ...float64) float64, args ...float64) timing.Signal {
	var send timing.Signal = make(chan float64, 1)

	go func() {
		defer Recover(send)
		send <- op(args...)
	}()

	return send
}

func On(ports ...interface{}) timing.Signal {
	var send timing.Signal = make(chan float64, 1)

	go func() {
		defer Recover(send)
		t_o := timing.Max()
		for _, b := range ports {
			if bi, ok := b.(Waiter); ok {
				t_o.Add(bi.Wait())
			} else if reflect.TypeOf(b).Kind() == reflect.Slice || reflect.TypeOf(b).Kind() == reflect.Array {
				items := reflect.ValueOf(b)
				for i := 0; i < items.Len(); i++ {
					item := items.Index(i).Interface()
					if bi, ok := item.(Waiter); ok {
						t_o.Add(bi.Wait())
					} else {
						panic(Misconfigured)
					}
				}
			} else {
				panic(Misconfigured)
			}
		}
		send <- t_o.Get()
	}()

	return send
}

type channel[T interface{}] struct {
	name string
	read int
	write int
	buffer []timing.Value[T]
	readyTime float64
	ready bool
	recvBlocked bool
	sendBlocked bool
	sendMu *sync.Mutex
	recvMu *sync.Mutex

	cond *sync.Cond
}

type Logger[T interface{}] interface {
	Write(value T, t float64)
	Close() error	
}

type logger[T interface{}] struct {
	filename string
	log *os.File
}

func Log[T interface{}](filename string) Logger[T] {
	return &logger[T]{
		filename: filename,
		log: nil,
	}
}

func (l *logger[T]) WriteType(t reflect.Type) {
	if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		fmt.Fprintf(l.log, "[")
		l.WriteType(t.Elem())
		fmt.Fprintf(l.log, "]")
	} else if t.Kind() == reflect.Struct {
		fmt.Fprintf(l.log, "{")
		for i := 0; i < t.NumField(); i++ {
			if i != 0 {
				fmt.Fprintf(l.log, " ")
			}
			f := t.Field(i)
			fmt.Fprintf(l.log, "%s:", f.Name)
			l.WriteType(f.Type)
		}
		fmt.Fprintf(l.log, "}")
	} else {
		fmt.Fprintf(l.log, "%s", t.Name())
	}
}

func (l *logger[T]) Write(value T, t float64) {
	if l.log == nil {
		var err error
		l.log, err = os.Create(l.filename)
		if err != nil {
			fmt.Println(err)
			return
		}
		var temp T
		fmt.Fprintf(l.log, "time (ns)\t")
		l.WriteType(reflect.TypeOf(temp))
		fmt.Fprintf(l.log, "\n")
	}

	if l.log != nil {
		fmt.Fprintf(l.log, "%f\t%v\n", t, value)
	}
}

func (l *logger[T]) Close() error {
	if l.log != nil {
		return l.log.Close()
	}
	return nil
}

type sender[T interface{}] struct {
	c *channel[T]
	g Globals

	log Logger[T]
}

type receiver[T interface{}] struct {
	c *channel[T]
	g Globals

	log Logger[T]
	logged bool
}

func Chan[T interface{}](name string, slack int64) (Sender[T], Receiver[T]) {
	c := &channel[T] {
		name: name,
		buffer: make([]timing.Value[T], slack+1), 
		sendMu: &sync.Mutex{},
		recvMu: &sync.Mutex{},
		cond: sync.NewCond(&sync.Mutex{}),
	}
	
	s := &sender[T]{
		c: c,
		g: nil,
		log: nil,
	}
	r := &receiver[T]{
		c: c,
		g: nil,
		log: nil,
		logged: false,
	}
	return s, r
}

func ChanArr[T interface{}](name string, n int, slack int64) ([]Sender[T], []Receiver[T]) {
	s := make([]Sender[T], n)
	r := make([]Receiver[T], n)
	
	for i := 0; i < n; i++ {
		c := &channel[T] {
			name: name+"."+strconv.Itoa(i),
			buffer: make([]timing.Value[T], slack+1), 
			sendMu: &sync.Mutex{},
			recvMu: &sync.Mutex{},
			cond: sync.NewCond(&sync.Mutex{}),
		}

		s[i] = &sender[T]{
			c: c,
			g: nil,
			log: nil,
		}
		r[i] = &receiver[T]{
			c: c,
			g: nil,
			log: nil,
			logged: false,
		}
	}

	return s, r
}

func Bus[T interface{}](name string, slack int64) Channel[T] {
	c := &channel[T] {
		name: name,
		buffer: make([]timing.Value[T], slack+1), 
		sendMu: &sync.Mutex{},
		recvMu: &sync.Mutex{},
		cond: sync.NewCond(&sync.Mutex{}),
	}
	
	s := &sender[T]{
		c: c,
		g: nil,
		log: nil,
	}
	r := &receiver[T]{
		c: c,
		g: nil,
		log: nil,
		logged: false,
	}
	return Channel[T]{s, r}
}

func BusArr[T interface{}](name string, n int, slack int64) []Channel[T] {
	t := make([]Channel[T], n)
	
	for i := 0; i < n; i++ {
		c := &channel[T] {
			name: name+"."+strconv.Itoa(i),
			buffer: make([]timing.Value[T], slack+1), 
			sendMu: &sync.Mutex{},
			recvMu: &sync.Mutex{},
			cond: sync.NewCond(&sync.Mutex{}),
		}

		t[i].S = &sender[T]{
			c: c,
			g: nil,
			log: nil,
		}
		t[i].R = &receiver[T]{
			c: c,
			g: nil,
			log: nil,
			logged: false,
		}
	}

	return t
}


func (c *channel[T]) full() bool {
	return c.write == c.read && c.ready
}

func (c *channel[T]) empty() bool {
	return c.read == c.write && !c.ready
}

func (c *channel[T]) sendDead() bool {
	return c.full() && c.recvBlocked || c.sendBlocked
}

func (c *channel[T]) recvDead() bool {
	return c.empty() && c.sendBlocked || c.recvBlocked
}

func (c *channel[T]) incRead(t float64) {
	if c.full() {
		c.readyTime = t
	}

	c.read = (c.read+1)%len(c.buffer)
	c.ready = false

	for i := c.read; i != c.write; i = (i+1)%len(c.buffer) {
		if t > c.buffer[i].T {
			c.buffer[i].T = t
		}
	}
}

func (c *channel[T]) incWrite() int {
	i := c.write
	c.write = (c.write+1)%len(c.buffer)
	c.ready = true
	return i
}

func (c *channel[T]) BeginSend() bool {
	c.sendMu.Lock()
	c.cond.L.Lock()
	defer c.cond.L.Unlock()

	for c.full() {
		if c.sendDead() {
			c.cond.Signal()
			return false
		}
		c.cond.Wait()
	}

	return true
}

func (c *channel[T]) EndSend() (float64, bool) {
	c.cond.L.Lock()
	defer c.sendMu.Unlock()
	defer c.cond.L.Unlock()

	i := c.incWrite()
	c.cond.Signal()
	for c.full() {
		if c.sendDead() {
			return c.buffer[i].T, false
		}
		c.cond.Wait()
	}
	if c.readyTime > c.buffer[i].T {
			c.buffer[i].T = c.readyTime
	}

	return c.buffer[i].T, true
}

func (c *channel[T]) BeginRecv() bool {
	c.recvMu.Lock()
	c.cond.L.Lock()
	defer c.cond.L.Unlock()

	for c.empty() {
		if c.recvDead() {
			c.cond.Signal()
			return false
		}
		c.cond.Wait()
	}

	return true
}

func (c *channel[T]) EndRecv(t float64) bool {
	c.cond.L.Lock()
	defer c.recvMu.Unlock()
	defer c.cond.L.Unlock()

	if c.recvDead() {
		c.cond.Signal()
		return false
	}

	c.incRead(t)

	c.cond.Signal()
	return true
}

func (c *channel[T]) EndWait(t float64) (float64, bool) {
	c.cond.L.Lock()
	defer c.sendMu.Unlock()
	defer c.cond.L.Unlock()

	if c.readyTime > t {
		t = c.readyTime
	}

	if c.sendDead() {
		c.cond.Signal()
		return t, false
	}

	c.cond.Signal()
	return t, true
}

func (c *channel[T]) EndProbe() bool {
	c.cond.L.Lock()
	defer c.recvMu.Unlock()
	defer c.cond.L.Unlock()

	if c.recvDead() {
		c.cond.Signal()
		return false
	}

	c.cond.Signal()
	return true
}

func (c *channel[T]) Ready() bool {
	c.cond.L.Lock()
	defer c.cond.L.Unlock()

	return (c.write+1)%len(c.buffer) != c.read
}

func (c *channel[T]) Valid() bool {
	c.cond.L.Lock()
	defer c.cond.L.Unlock()

	return !c.empty()
}

func (s *sender[T]) SetGlobals(g Globals) {
	if s.g != nil {
		panic(Misconfigured)
	}
	s.g = g
	if s.c.name != "" {
		s.log = Log[T](filepath.Join(g.Dir(), g.Name()+"."+s.c.name+".s"))
	}
}

func (s *sender[T]) Send(value T, args ...float64) float64 {
	if s.g == nil {
		panic(fmt.Errorf("you must call g.Init for this sender"))
	}

	var start float64 = s.g.Curr()
	if len(args) > 0 {
		start += args[0]
	}

	s.g.Timing()
	if s.g.Debug() && s.c.name != "" {
		fmt.Printf("%f ns\t\t%s!%v\t\t%s\n", start, s.c.name, value, s.g.Name())
	}

	if !s.c.BeginSend() {
		panic(timing.Deadlock)
	}

	s.c.buffer[s.c.write] = timing.Value[T]{start, value}
	
	s.g.Timing()
	t, ok := s.c.EndSend()
	if !ok {
		panic(timing.Deadlock)
	}

	if s.log != nil {
		s.log.Write(value, t)
	}

	if s.g.Debug() && s.c.name != "" {
		fmt.Printf("%f ns\t\t  %s¡\t\t%s\n", t, s.c.name, s.g.Name())
	}

	return t - s.g.Curr()
}

func (s *sender[T]) Offer(value T, args ...float64) timing.Signal {
	var send timing.Signal = make(chan float64, 1)

	go func() {
		defer Recover(chan float64(send))
		send <- s.Send(value, args...)
	}()

	return send
}

func (s *sender[T]) Watch(args ...float64) timing.Signal {
	var send timing.Signal = make(chan float64, 1)

	go func() {
		defer Recover(send)
		send <- s.Wait(args...)
	}()

	return send
}

func (s *sender[T]) Ready() bool {
	if s.g == nil {
		panic(fmt.Errorf("you must call g.Init for this sender"))
	}

	s.g.Timing()
	rdy := s.c.Ready()
	if s.g.Debug() && s.c.name != "" {
		if rdy {
			fmt.Printf("%f ns\t\t#%s!\t\t%s\n", s.g.Curr(), s.c.name, s.g.Name())
		} else {
			fmt.Printf("%f ns\t\t~#%s!\t\t%s\n", s.g.Curr(), s.c.name, s.g.Name())
		}
	}

	return rdy
}

func (s *sender[T]) Wait(args ...float64) float64 {
	if s.g == nil {
		panic(fmt.Errorf("you must call g.Init for this sender"))
	}

	var start float64 = s.g.Curr()
	if len(args) > 0 {
		start += args[0]
	}

	s.g.Timing()
	if s.g.Debug() && s.c.name != "" {
		fmt.Printf("%f ns\t\t#%s!\t\t%s\n", start, s.c.name, s.g.Name())
	}
	
	if !s.c.BeginSend() {
		panic(timing.Deadlock)
	}
	
	s.g.Timing()
	t, ok := s.c.EndWait(start)
	if !ok {
		panic(timing.Deadlock)
	}

	if s.g.Debug() && s.c.name != "" {
		fmt.Printf("%f ns\t\t  #%s¡\t\t%s\n", t, s.c.name, s.g.Name())
	}

	return t - s.g.Curr()
}

func (s *sender[T]) Close() error {
	s.c.cond.L.Lock()
	defer s.c.cond.L.Unlock()
	if s.log != nil {
		err := s.log.Close()
		if err != nil {
			return err
		}
		s.log = nil
	}
	s.c.sendBlocked = true
	s.c.cond.Signal()
	return nil
}

func (r *receiver[T]) SetGlobals(g Globals) {
	if r.g != nil {
		panic(Misconfigured)
	}
	r.g = g
	if r.c.name != "" {
		r.log = Log[T](filepath.Join(g.Dir(), g.Name()+"."+r.c.name+".r"))
	}
}

func (r *receiver[T]) Recv(args ...float64) (T, float64) {
	if r.g == nil {
		panic(fmt.Errorf("you must call g.Init for this receiver"))
	}

	var start float64 = r.g.Curr()
	if len(args) > 0 {
		start += args[0]
	}

	r.g.Timing()
	if r.g.Debug() && r.c.name != "" {
		fmt.Printf("%f ns\t\t%s?\t\t%s\n", start, r.c.name, r.g.Name())
	}

	if !r.c.BeginRecv() {
		panic(timing.Deadlock)
	}
	
	result := r.c.buffer[r.c.read]
	if start > result.T {
		result.T = start
	}

	if !r.logged && r.log != nil {
		r.log.Write(result.V, result.T)
	}
	r.logged = false

	r.g.Timing()
	if !r.c.EndRecv(result.T) {
		panic(timing.Deadlock)
	}

	if r.g.Debug() && r.c.name != "" {
		fmt.Printf("%f ns\t\t  %s¿%v\t\t%s\n", result.T, r.c.name, result.V, r.g.Name())
	}

	return result.V, result.T - r.g.Curr()
}

func (r *receiver[T]) Expect(args ...float64) timing.Action[T] {
	var recv timing.Action[T] = make(chan timing.Value[T], 1)

	go func() {
		defer Recover(chan timing.Value[T](recv))
		v, t := r.Recv(args...)
		recv <- timing.Value[T]{t, v}
	}()

	return recv
}

func (r *receiver[T]) Read(args ...float64) timing.Action[T] {
	var recv timing.Action[T] = make(chan timing.Value[T], 1)

	go func() {
		defer Recover(recv)
		v, t := r.Probe(args...)
		recv <- timing.Value[T]{t, v}
	}()

	return recv
}

func (r *receiver[T]) Valid() bool {
	if r.g == nil {
		panic(fmt.Errorf("you must call g.Init for this receiver"))
	}

	r.g.Timing()
	val := r.c.Valid()
	if r.g.Debug() && r.c.name != "" {
		if val {
			fmt.Printf("%f ns\t\t#%s?\t\t%s\n", r.g.Curr(), r.c.name, r.g.Name())
		} else {
			fmt.Printf("%f ns\t\t~#%s?\t\t%s\n", r.g.Curr(), r.c.name, r.g.Name())
		}
	}

	return val
}

func (r *receiver[T]) Probe(args ...float64) (T, float64) {
	if r.g == nil {
		panic(fmt.Errorf("you must call g.Init for this receiver"))
	}

	var start float64 = r.g.Curr()
	if len(args) > 0 {
		start += args[0]
	}

	r.g.Timing()
	if r.g.Debug() && r.c.name != "" {
		fmt.Printf("%f ns\t\t#%s?\t\t%s\n", start, r.c.name, r.g.Name())
	}

	if !r.c.BeginRecv() {
		panic(timing.Deadlock)
	}

	result := r.c.buffer[r.c.read]
	if start > result.T {
		result.T = start
	}

	if !r.logged && r.log != nil {
		r.log.Write(result.V, result.T)
	}
	r.logged = true

	r.g.Timing()
	if !r.c.EndProbe() {
		panic(timing.Deadlock)
	}

	if r.g.Debug() && r.c.name != "" {
		fmt.Printf("%f ns\t\t  #%s¿%v\t\t%s\n", result.T, r.c.name, result.V, r.g.Name())
	}

	return result.V, result.T - r.g.Curr()
}

func (r *receiver[T]) Wait(args ...float64) float64 {
	_, t := r.Probe(args...)
	return t
}

func (r *receiver[T]) Close() error {
	r.c.cond.L.Lock()
	defer r.c.cond.L.Unlock()
	if r.log != nil {
		err := r.log.Close()
		if err != nil {
			return err
		}
		r.log = nil
	}
	r.c.recvBlocked = true
	r.c.cond.Signal()
	return nil
}

