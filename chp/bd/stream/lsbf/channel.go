package lsbf

import (
	"math/big"
	"sync"

	"git.broccolimicro.io/Broccoli/pr.git/chp"
	"git.broccolimicro.io/Broccoli/pr.git/chp/timing"
	"git.broccolimicro.io/Broccoli/pr.git/chp/bd"
	"git.broccolimicro.io/Broccoli/pr.git/chp/bd/stream"
)

/******************************
*          Channel            *
******************************/

type Sender interface {
	chp.Sender[int64]

	Raw() stream.Sender[int64]
	Base() int64
}

type Receiver interface {
	chp.Receiver[int64]

	Raw() stream.Receiver[int64]
	Base() int64
}

type sender struct {
	raw stream.Sender[int64]
	base int64
}

type receiver struct {
	raw stream.Receiver[int64]
	base int64
}

func Chan(name string, base int64, slack int64) (Sender, Receiver) {
	s, r := stream.Chan[int64](name, slack)
	return &sender{
		raw: s,
		base: base,
	}, &receiver{
		raw: r,
		base: base,
	}
}

func ChanArr(name string, n int, base int64, slack int64) ([]Sender, []Receiver) {
	s, r := stream.ChanArr[int64](name, n, slack)
	S := make([]Sender, n)
	R := make([]Receiver, n)
	for i := 0; i < n; i++ {
		S[i] = &sender{
			raw: s[i],
			base: base,
		}
		R[i] = &receiver{
			raw: r[i],
			base: base,
		}
	}
	return S, R
}

func RawSenders(s []Sender) []stream.Sender[int64] {
	result := make([]stream.Sender[int64], len(s))
	for i := 0; i < len(s); i++ {
		result[i] = s[i].Raw()
	}
	return result
}

func RawReceivers(s []Receiver) []stream.Receiver[int64] {
	result := make([]stream.Receiver[int64], len(s))
	for i := 0; i < len(s); i++ {
		result[i] = s[i].Raw()
	}
	return result
}

func (self *sender) Raw() stream.Sender[int64] {
	return self.raw
}

func (self *sender) Base() int64 {
	return self.base
}

func (self *sender) SetGlobals(g chp.Globals) {
	self.raw.SetGlobals(g)
}

func (self *sender) Offer(value int64, args ...float64) chp.Signal {
	var send chp.Signal = make(chan float64, 1)

	go func() {
		defer chp.Recover(chan float64(send))
		send <- self.Send(value, args...)
	}()

	return send
}

func (self *sender) Send(value int64, args ...float64) float64 {
	return self.raw.SendStream(FromInt64(value, self.base), args...)
}

func (self *sender) Ready(args ...float64) chp.Signal {
	return self.raw.Ready(args...)
}

func (self *sender) Wait(args ...float64) float64 {
	return self.raw.Wait(args...)
}

func (self *sender) Close() error {
	return self.raw.Close()
}

func (self *receiver) Raw() stream.Receiver[int64] {
	return self.raw
}

func (self *receiver) Base() int64 {
	return self.base
}

func (self *receiver) SetGlobals(g chp.Globals) {
	self.raw.SetGlobals(g)
}

func (r *receiver) Expect(args ...float64) chp.Value[int64] {
	var recv chp.Value[int64] = make(chan chp.Action[int64], 1)

	go func() {
		defer chp.Recover(chan chp.Action[int64](recv))
		v, t := r.Recv(args...)
		recv <- chp.Action[int64]{t, v}
	}()

	return recv
}

func (self *receiver) Recv(args ...float64) (int64, float64) {
	v, t := self.raw.RecvStream(args...)
	return ToInt64(v, self.base), t
}

// Not supported by this receiver type, autopanic
func (self *receiver) Valid(args ...float64) chp.Value[int64] {
	var result chp.Value[int64] = make(chan chp.Action[int64], 1)
	close(result)
	return result
}

// Not supported by this receiver type, autopanic
func (self *receiver) Probe(args ...float64) (int64, float64) {
	panic(chp.Deadlock)
	return 0, 0.0
}

func (self *receiver) Close() error {
	return self.raw.Close()
}

/******************************
*          BigChannel         *
******************************/

type BigSender interface {
	chp.Sender[*big.Int]

	Raw() stream.Sender[int64]
	Base() int64
}

type BigReceiver interface {
	chp.Receiver[*big.Int]

	Raw() stream.Receiver[int64]
	Base() int64
}

type bigsender struct {
	raw stream.Sender[int64]
	base int64
}

type bigreceiver struct {
	raw stream.Receiver[int64]
	base int64
}

func BigChan(name string, base int64, slack int64) (BigSender, BigReceiver) {
	s, r := stream.Chan[int64](name, slack)
	return &bigsender{
		raw: s,
		base: base,
	}, &bigreceiver{
		raw: r,
		base: base,
	}
}

func BigChanArr(name string, n int, base int64, slack int64) ([]BigSender, []BigReceiver) {
	s, r := stream.ChanArr[int64](name, n, slack)
	S := make([]BigSender, n)
	R := make([]BigReceiver, n)
	for i := 0; i < n; i++ {
		S[i] = &bigsender{
			raw: s[i],
			base: base,
		}
		R[i] = &bigreceiver{
			raw: r[i],
			base: base,
		}
	}
	return S, R
}

func RawBigSenders(s []BigSender) []stream.Sender[int64] {
	result := make([]stream.Sender[int64], len(s))
	for i := 0; i < len(s); i++ {
		result[i] = s[i].Raw()
	}
	return result
}

func RawBigReceivers(s []BigReceiver) []stream.Receiver[int64] {
	result := make([]stream.Receiver[int64], len(s))
	for i := 0; i < len(s); i++ {
		result[i] = s[i].Raw()
	}
	return result
}

func (self *bigsender) Raw() stream.Sender[int64] {
	return self.raw
}

func (self *bigsender) Base() int64 {
	return self.base
}

func (self *bigsender) SetGlobals(g chp.Globals) {
	self.raw.SetGlobals(g)
}

func (self *bigsender) Offer(value *big.Int, args ...float64) chp.Signal {
	var send chp.Signal = make(chan float64, 1)

	go func() {
		defer chp.Recover(chan float64(send))
		send <- self.Send(value, args...)
	}()

	return send
}

func (self *bigsender) Send(value *big.Int, args ...float64) float64 {
	return self.raw.SendStream(FromBigInt(value, self.base), args...)
}

func (self *bigsender) Ready(args ...float64) chp.Signal {
	return self.raw.Ready(args...)
}

func (self *bigsender) Wait(args ...float64) float64 {
	return self.raw.Wait(args...)
}

func (self *bigsender) Close() error {
	return self.raw.Close()
}

func (self *bigreceiver) Raw() stream.Receiver[int64] {
	return self.raw
}

func (self *bigreceiver) Base() int64 {
	return self.base
}

func (self *bigreceiver) SetGlobals(g chp.Globals) {
	self.raw.SetGlobals(g)
}

func (r *bigreceiver) Expect(args ...float64) chp.Value[*big.Int] {
	var recv chp.Value[*big.Int] = make(chan chp.Action[*big.Int], 1)

	go func() {
		defer chp.Recover(chan chp.Action[*big.Int](recv))
		v, t := r.Recv(args...)
		recv <- chp.Action[*big.Int]{t, v}
	}()

	return recv
}

func (self *bigreceiver) Recv(args ...float64) (*big.Int, float64) {
	v, t := self.raw.RecvStream(args...)
	return ToBigInt(v, self.base), t
}

// Not supported by this bigreceiver type, autopanic
func (self *bigreceiver) Valid(args ...float64) chp.Value[*big.Int] {
	var result chp.Value[*big.Int] = make(chan chp.Action[*big.Int], 1)
	close(result)
	return result
}

// Not supported by this bigreceiver type, autopanic
func (self *bigreceiver) Probe(args ...float64) (*big.Int, float64) {
	panic(chp.Deadlock)
	return nil, 0.0
}

func (self *bigreceiver) Close() error {
	return self.raw.Close()
}

/******************************
*      ParallelChannel        *
******************************/

type ParallelSender interface {
	chp.Sender[int64]

	Raw() []stream.Sender[int64]
	Base() int64
}

type ParallelReceiver interface {
	chp.Receiver[int64]

	Raw() []stream.Receiver[int64]
	Base() int64
}

type parallelsender struct {
	raw []stream.Sender[int64]
	base int64
}

type parallelreceiver struct {
	raw []stream.Receiver[int64]
	base int64
}

func ParallelChan(name string, n int, base int64, slack int64) (ParallelSender, ParallelReceiver) {
	s, r := stream.ChanArr[int64](name, n, slack)
	return &parallelsender{
		raw: s,
		base: base,
	}, &parallelreceiver{
		raw: r,
		base: base,
	}
}

func (self *parallelsender) Raw() []stream.Sender[int64] {
	return self.raw
}

func (self *parallelsender) Base() int64 {
	return self.base
}

func (self *parallelsender) SetGlobals(g chp.Globals) {
	for _, s := range self.raw {
		s.SetGlobals(g)
	}
}

func (self *parallelsender) Offer(value int64, args ...float64) chp.Signal {
	var send chp.Signal = make(chan float64, 1)

	go func() {
		defer chp.Recover(chan float64(send))
		send <- self.Send(value, args...)
	}()

	return send
}

func (self *parallelsender) Send(value int64, args ...float64) float64 {
	digits := FromInt64(value, self.base)
	g := &sync.WaitGroup{}

	ts := timing.Set()
	mu := &sync.Mutex{}
	for i, digit := range digits {
		if i < len(self.raw) {
			g.Add(1)
			go func(g *sync.WaitGroup, c stream.Sender[int64], last bool, val int64) {
				defer g.Done()
				t := c.SendToken(last, val, args...)
				mu.Lock()
				defer mu.Unlock()
				ts.Max(t)
			}(g, self.raw[i], i == len(digits)-1, digit)
		}
	}
	g.Wait()
	return ts.Get()
}

func (self *parallelsender) Ready(args ...float64) chp.Signal {
	var result chp.Signal = make(chan float64, 1)
	close(result)
	return result
}

func (self *parallelsender) Wait(args ...float64) float64 {
	panic(chp.Deadlock)
	return 0.0
}

func (self *parallelsender) Close() error {
	for i := 0; i < len(self.raw); i++ {
		err := self.raw[i].Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *parallelreceiver) Raw() []stream.Receiver[int64] {
	return self.raw
}

func (self *parallelreceiver) Base() int64 {
	return self.base
}

func (self *parallelreceiver) SetGlobals(g chp.Globals) {
	for _, r := range self.raw {
		r.SetGlobals(g)
	}
}

func (r *parallelreceiver) Expect(args ...float64) chp.Value[int64] {
	var recv chp.Value[int64] = make(chan chp.Action[int64], 1)

	go func() {
		defer chp.Recover(chan chp.Action[int64](recv))
		v, t := r.Recv(args...)
		recv <- chp.Action[int64]{t, v}
	}()

	return recv
}

func (self *parallelreceiver) Recv(args ...float64) (int64, float64) {
	var start float64 = 0.0
	if len(args) > 0 {
		start = args[0]
	}

	var step float64 = 0.0
	if len(args) > 1 {
		step = args[1]
	}

	var token bd.Token[bool, int64]
	var digits []int64

	end := start
	for i := 0; i < len(self.raw) && !token.C; i++ {
		token, end = self.raw[i].Recv(start)
		digits = append(digits, token.D)
		start += step
	}

	return ToInt64(digits, self.base), end
}

// Not supported by this parallelreceiver type, autopanic
func (self *parallelreceiver) Valid(args ...float64) chp.Value[int64] {
	var result chp.Value[int64] = make(chan chp.Action[int64], 1)
	close(result)
	return result
}

// Not supported by this parallelreceiver type, autopanic
func (self *parallelreceiver) Probe(args ...float64) (int64, float64) {
	panic(chp.Deadlock)
	return 0, 0.0
}

func (self *parallelreceiver) Close() error {
	for i := 0; i < len(self.raw); i++ {
		err := self.raw[i].Close()
		if err != nil {
			return err
		}
	}
	return nil
}

/******************************
*      BigChannelArr          *
******************************/

type BigParallelSender interface {
	chp.Sender[*big.Int]

	Raw() []stream.Sender[int64]
	Base() int64
}

type BigParallelReceiver interface {
	chp.Receiver[*big.Int]

	Raw() []stream.Receiver[int64]
	Base() int64
}

type bigparallelsender struct {
	raw []stream.Sender[int64]
	base int64
}

type bigparallelreceiver struct {
	raw []stream.Receiver[int64]
	base int64
}

func BigParallelChan(name string, n int, base int64, slack int64) (BigParallelSender, BigParallelReceiver) {
	s, r := stream.ChanArr[int64](name, n, slack)
	return &bigparallelsender{
		raw: s,
		base: base,
	}, &bigparallelreceiver{
		raw: r,
		base: base,
	}
}

func (self *bigparallelsender) Raw() []stream.Sender[int64] {
	return self.raw
}

func (self *bigparallelsender) Base() int64 {
	return self.base
}

func (self *bigparallelsender) SetGlobals(g chp.Globals) {
	for _, s := range self.raw {
		s.SetGlobals(g)
	}
}

func (self *bigparallelsender) Offer(value *big.Int, args ...float64) chp.Signal {
	var send chp.Signal = make(chan float64, 1)

	go func() {
		defer chp.Recover(chan float64(send))
		send <- self.Send(value, args...)
	}()

	return send
}

func (self *bigparallelsender) Send(value *big.Int, args ...float64) float64 {
	digits := FromBigInt(value, self.base)
	g := &sync.WaitGroup{}

	ts := timing.Set()
	mu := &sync.Mutex{}
	for i, digit := range digits {
		if i < len(self.raw) {
			g.Add(1)
			go func(g *sync.WaitGroup, c stream.Sender[int64], last bool, val int64) {
				defer g.Done()
				t := c.SendToken(last, val, args...)
				mu.Lock()
				defer mu.Unlock()
				ts.Max(t)
			}(g, self.raw[i], i == len(digits)-1, digit)
		}
	}
	g.Wait()
	return ts.Get()
}

func (self *bigparallelsender) Ready(args ...float64) chp.Signal {
	var result chp.Signal = make(chan float64, 1)
	close(result)
	return result
}

func (self *bigparallelsender) Wait(args ...float64) float64 {
	panic(chp.Deadlock)
	return 0.0
}

func (self *bigparallelsender) Close() error {
	for i := 0; i < len(self.raw); i++ {
		err := self.raw[i].Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *bigparallelreceiver) Raw() []stream.Receiver[int64] {
	return self.raw
}

func (self *bigparallelreceiver) Base() int64 {
	return self.base
}

func (self *bigparallelreceiver) SetGlobals(g chp.Globals) {
	for _, r := range self.raw {
		r.SetGlobals(g)
	}
}

func (r *bigparallelreceiver) Expect(args ...float64) chp.Value[*big.Int] {
	var recv chp.Value[*big.Int] = make(chan chp.Action[*big.Int], 1)

	go func() {
		defer chp.Recover(chan chp.Action[*big.Int](recv))
		v, t := r.Recv(args...)
		recv <- chp.Action[*big.Int]{t, v}
	}()

	return recv
}

func (self *bigparallelreceiver) Recv(args ...float64) (*big.Int, float64) {
	var start float64 = 0.0
	if len(args) > 0 {
		start = args[0]
	}

	var step float64 = 0.0
	if len(args) > 1 {
		step = args[1]
	}

	var token bd.Token[bool, int64]
	var digits []int64

	end := start
	for i := 0; i < len(self.raw) && !token.C; i++ {
		token, end = self.raw[i].Recv(start)
		digits = append(digits, token.D)
		start += step
	}

	return ToBigInt(digits, self.base), end
}

// Not supported by this channel type, autopanic
func (self *bigparallelreceiver) Valid(args ...float64) chp.Value[*big.Int] {
	var result chp.Value[*big.Int] = make(chan chp.Action[*big.Int], 1)
	close(result)
	return result
}

// Not supported by this channel type, autopanic
func (self *bigparallelreceiver) Probe(args ...float64) (*big.Int, float64) {
	panic(chp.Deadlock)
	return nil, 0.0
}

func (self *bigparallelreceiver) Close() error {
	for i := 0; i < len(self.raw); i++ {
		err := self.raw[i].Close()
		if err != nil {
			return err
		}
	}
	return nil
}

