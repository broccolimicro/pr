package chp

import (
	"sync"
	"fmt"
	"errors"
	"io"
	"strconv"
	"os"
	"path/filepath"
	"reflect"
)

var Deadlock error = errors.New("Deadlock")
var Conflict error = errors.New("Conflict")

type Void struct {}

var Null = Void{}

type Action[vtype interface{}] struct {
	T float64
	V vtype
}

type Value[T interface{}] chan Action[T]

func (p Value[T]) Recv() (T, float64) {
	a, ok := <-p
	if !ok {
		panic(Deadlock)
	}
	return a.V, a.T
}

type Signal chan float64

func (p Signal) Send() float64 {
	a, ok := <-p
	if !ok {
		panic(Deadlock)
	}
	return a
}

type Recordable interface {
	SetGlobals(g Globals)
}

type Sender[T interface{}] interface {
	io.Closer
	Recordable

	Offer(value T, args ...float64) Signal
	Send(value T, args ...float64) float64
	Ready(args ...float64) <-chan float64
	Wait(args ...float64) float64
}

type Receiver[T interface{}] interface {
	io.Closer
	Recordable

	Expect(args ...float64) Value[T]
	Recv(args ...float64) (T, float64)
	Valid(args ...float64) <-chan Action[T]
	Probe(args ...float64) (T, float64)
}

func Recover[T interface{}](c chan T) {
	r := recover()
	if r == Deadlock {
		close(c)
	} else if r != nil {
		panic(r)
	}
}

type channel[T interface{}] struct {
	name string
	read int
	write int
	buffer []Action[T]
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
		buffer: make([]Action[T], slack+1), 
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
			buffer: make([]Action[T], slack+1), 
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
	c.cond.L.Lock()
	defer c.cond.L.Unlock()

	for c.full() {
		if c.sendDead() {
			c.cond.Signal()
			return false
		}
		c.cond.Wait()
	}

	c.sendMu.Lock()
	return true
}

func (c *channel[T]) EndSend() (float64, bool) {
	c.cond.L.Lock()
	defer c.cond.L.Unlock()
	defer c.sendMu.Unlock()

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
	c.cond.L.Lock()
	defer c.cond.L.Unlock()

	for c.empty() {
		if c.recvDead() {
			c.cond.Signal()
			return false
		}
		c.cond.Wait()
	}

	c.recvMu.Lock()
	return true
}

func (c *channel[T]) EndRecv(t float64) bool {
	c.cond.L.Lock()
	defer c.cond.L.Unlock()
	defer c.recvMu.Unlock()

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
	defer c.cond.L.Unlock()
	defer c.sendMu.Unlock()

	if c.readyTime > t {
		t = c.readyTime
	}
	c.cond.Signal()
	return t, !c.sendDead()
}

func (c *channel[T]) EndProbe() bool {
	c.cond.L.Lock()
	defer c.cond.L.Unlock()
	defer c.recvMu.Unlock()

	c.cond.Signal()
	return !c.recvDead()
}

func (s *sender[T]) SetGlobals(g Globals) {
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

	if s.g.Debug() && s.c.name != "" {
		fmt.Printf("%f ns\t\t%s!%v\t\t%s\n", start, s.c.name, value, s.g.Name())
	}

	if !s.c.BeginSend() {
		panic(Deadlock)
	}

	s.c.buffer[s.c.write] = Action[T]{start, value}
	
	t, ok := s.c.EndSend()
	if !ok {
		panic(Deadlock)
	}

	if s.log != nil {
		s.log.Write(value, t)
	}

	if s.g.Debug() && s.c.name != "" {
		fmt.Printf("%f ns\t\t  %s¡\t\t%s\n", t, s.c.name, s.g.Name())
	}

	return t - s.g.Curr()
}

func (s *sender[T]) Offer(value T, args ...float64) Signal {
	var send Signal = make(chan float64, 1)

	go func() {
		defer Recover(chan float64(send))
		send <- s.Send(value, args...)
	}()

	return send
}

func (s *sender[T]) Ready(args ...float64) <-chan float64 {
	var send chan float64 = make(chan float64, 1)

	go func() {
		defer Recover(send)
		send <- s.Wait(args...)
	}()

	return send
}

func (s *sender[T]) Wait(args ...float64) float64 {
	if s.g == nil {
		panic(fmt.Errorf("you must call g.Init for this sender"))
	}

	var start float64 = s.g.Curr()
	if len(args) > 0 {
		start += args[0]
	}

	if s.g.Debug() && s.c.name != "" {
		fmt.Printf("%f ns\t\t#%s!\t\t%s\n", start, s.c.name, s.g.Name())
	}
	
	if !s.c.BeginSend() {
		panic(Deadlock)
	}
	
	t, ok := s.c.EndWait(start)
	if !ok {
		panic(Deadlock)
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

	if r.g.Debug() && r.c.name != "" {
		fmt.Printf("%f ns\t\t%s?\t\t%s\n", start, r.c.name, r.g.Name())
	}

	if !r.c.BeginRecv() {
		panic(Deadlock)
	}
	
	result := r.c.buffer[r.c.read]
	if start > result.T {
		result.T = start
	}

	if !r.logged && r.log != nil {
		r.log.Write(result.V, result.T)
	}
	r.logged = false

	if !r.c.EndRecv(result.T) {
		panic(Deadlock)
	}

	if r.g.Debug() && r.c.name != "" {
		fmt.Printf("%f ns\t\t  %s¿%v\t\t%s\n", result.T, r.c.name, result.V, r.g.Name())
	}

	return result.V, result.T - r.g.Curr()
}

func (r *receiver[T]) Expect(args ...float64) Value[T] {
	var recv Value[T] = make(chan Action[T], 1)

	go func() {
		defer Recover(chan Action[T](recv))
		v, t := r.Recv(args...)
		recv <- Action[T]{t, v}
	}()

	return recv
}

func (r *receiver[T]) Valid(args ...float64) <-chan Action[T] {
	var recv chan Action[T] = make(chan Action[T], 1)

	go func() {
		defer Recover(recv)
		v, t := r.Probe(args...)
		recv <- Action[T]{t, v}
	}()

	return recv
}

func (r *receiver[T]) Probe(args ...float64) (T, float64) {
	if r.g == nil {
		panic(fmt.Errorf("you must call g.Init for this receiver"))
	}

	var start float64 = r.g.Curr()
	if len(args) > 0 {
		start += args[0]
	}

	if r.g.Debug() && r.c.name != "" {
		fmt.Printf("%f ns\t\t#%s?\t\t%s\n", start, r.c.name, r.g.Name())
	}

	if !r.c.BeginRecv() {
		panic(Deadlock)
	}

	result := r.c.buffer[r.c.read]
	if start > result.T {
		result.T = start
	}

	if !r.logged && r.log != nil {
		r.log.Write(result.V, result.T)
	}
	r.logged = true

	if !r.c.EndProbe() {
		panic(Deadlock)
	}

	if r.g.Debug() && r.c.name != "" {
		fmt.Printf("%f ns\t\t  #%s¿%v\t\t%s\n", result.T, r.c.name, result.V, r.g.Name())
	}

	return result.V, result.T - r.g.Curr()
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

