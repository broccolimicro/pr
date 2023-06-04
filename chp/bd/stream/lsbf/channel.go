package lsbf

import (
	"math/big"

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

func (self *sender) Offer(value int64, args ...float64) timing.Signal {
	var send timing.Signal = make(chan float64, 1)

	go func() {
		defer timing.Catch(chan float64(send))
		send <- self.Send(value, args...)
	}()

	return send
}

func (self *sender) Send(value int64, args ...float64) float64 {
	return self.raw.SendStream(FromInt64(value, self.base), args...)
}

func (self *sender) Watch(args ...float64) timing.Signal {
	return self.raw.Watch(args...)
}

func (self *sender) Ready() bool {
	return self.raw.Ready()
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

func (r *receiver) Expect(args ...float64) timing.Action[int64] {
	var recv timing.Action[int64] = make(chan timing.Value[int64], 1)

	go func() {
		defer timing.Catch(chan timing.Value[int64](recv))
		v, t := r.Recv(args...)
		recv <- timing.Value[int64]{t, v}
	}()

	return recv
}

func (self *receiver) Recv(args ...float64) (int64, float64) {
	v, t := self.raw.RecvStream(args...)
	return ToInt64(v, self.base), t
}

// Not supported by this receiver type, autopanic
func (self *receiver) Read(args ...float64) timing.Action[int64] {
	var result timing.Action[int64] = make(chan timing.Value[int64], 1)
	close(result)
	return result
}

// Not supported by this receiver type, autopanic
func (self *receiver) Valid() bool {
	panic(timing.Deadlock)
	return false
}

// Not supported by this receiver type, autopanic
func (self *receiver) Probe(args ...float64) (int64, float64) {
	panic(timing.Deadlock)
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

func (self *bigsender) Offer(value *big.Int, args ...float64) timing.Signal {
	var send timing.Signal = make(chan float64, 1)

	go func() {
		defer timing.Catch(chan float64(send))
		send <- self.Send(value, args...)
	}()

	return send
}

func (self *bigsender) Send(value *big.Int, args ...float64) float64 {
	return self.raw.SendStream(FromBigInt(value, self.base), args...)
}

func (self *bigsender) Watch(args ...float64) timing.Signal {
	return self.raw.Watch(args...)
}

func (self *bigsender) Ready() bool {
	return self.raw.Ready()
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

func (r *bigreceiver) Expect(args ...float64) timing.Action[*big.Int] {
	var recv timing.Action[*big.Int] = make(chan timing.Value[*big.Int], 1)

	go func() {
		defer timing.Catch(chan timing.Value[*big.Int](recv))
		v, t := r.Recv(args...)
		recv <- timing.Value[*big.Int]{t, v}
	}()

	return recv
}

func (self *bigreceiver) Recv(args ...float64) (*big.Int, float64) {
	v, t := self.raw.RecvStream(args...)
	return ToBigInt(v, self.base), t
}

// Not supported by this bigreceiver type, autopanic
func (self *bigreceiver) Read(args ...float64) timing.Action[*big.Int] {
	var result timing.Action[*big.Int] = make(chan timing.Value[*big.Int], 1)
	close(result)
	return result
}

// Not supported by this bigreceiver type, autopanic
func (self *bigreceiver) Valid() bool {
	panic(timing.Deadlock)
	return false
}

// Not supported by this bigreceiver type, autopanic
func (self *bigreceiver) Probe(args ...float64) (*big.Int, float64) {
	panic(timing.Deadlock)
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

func (self *parallelsender) Offer(value int64, args ...float64) timing.Signal {
	var send timing.Signal = make(chan float64, 1)

	go func() {
		defer timing.Catch(chan float64(send))
		send <- self.Send(value, args...)
	}()

	return send
}

func (self *parallelsender) Send(value int64, args ...float64) float64 {
	digits := FromInt64(value, self.base)

	ts := timing.Max()
	for i, digit := range digits {
		if i < len(self.raw) {
			ts.Add(self.raw[i].OfferToken(i == len(digits)-1, digit, args...))
		}
	}
	return ts.Get()
}

func (self *parallelsender) Watch(args ...float64) timing.Signal {
	var result timing.Signal = make(chan float64, 1)
	close(result)
	return result
}

func (self *parallelsender) Ready() bool {
	panic(timing.Deadlock)
	return false
}

func (self *parallelsender) Wait(args ...float64) float64 {
	panic(timing.Deadlock)
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

func (r *parallelreceiver) Expect(args ...float64) timing.Action[int64] {
	var recv timing.Action[int64] = make(chan timing.Value[int64], 1)

	go func() {
		defer timing.Catch(chan timing.Value[int64](recv))
		v, t := r.Recv(args...)
		recv <- timing.Value[int64]{t, v}
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
func (self *parallelreceiver) Read(args ...float64) timing.Action[int64] {
	var result timing.Action[int64] = make(chan timing.Value[int64], 1)
	close(result)
	return result
}

// Not supported by this parallelreceiver type, autopanic
func (self *parallelreceiver) Valid() bool {
	panic(timing.Deadlock)
	return false
}

// Not supported by this parallelreceiver type, autopanic
func (self *parallelreceiver) Probe(args ...float64) (int64, float64) {
	panic(timing.Deadlock)
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

func (self *bigparallelsender) Offer(value *big.Int, args ...float64) timing.Signal {
	var send timing.Signal = make(chan float64, 1)

	go func() {
		defer timing.Catch(chan float64(send))
		send <- self.Send(value, args...)
	}()

	return send
}

func (self *bigparallelsender) Send(value *big.Int, args ...float64) float64 {
	digits := FromBigInt(value, self.base)

	ts := timing.Max()
	for i, digit := range digits {
		if i < len(self.raw) {
			ts.Add(self.raw[i].OfferToken(i == len(digits)-1, digit, args...))
		}
	}
	return ts.Get()
}

func (self *bigparallelsender) Watch(args ...float64) timing.Signal {
	var result timing.Signal = make(chan float64, 1)
	close(result)
	return result
}

func (self *bigparallelsender) Ready() bool {
	panic(timing.Deadlock)
	return false
}

func (self *bigparallelsender) Wait(args ...float64) float64 {
	panic(timing.Deadlock)
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

func (r *bigparallelreceiver) Expect(args ...float64) timing.Action[*big.Int] {
	var recv timing.Action[*big.Int] = make(chan timing.Value[*big.Int], 1)

	go func() {
		defer timing.Catch(chan timing.Value[*big.Int](recv))
		v, t := r.Recv(args...)
		recv <- timing.Value[*big.Int]{t, v}
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
func (self *bigparallelreceiver) Read(args ...float64) timing.Action[*big.Int] {
	var result timing.Action[*big.Int] = make(chan timing.Value[*big.Int], 1)
	close(result)
	return result
}

// Not supported by this channel type, autopanic
func (self *bigparallelreceiver) Valid() bool {
	panic(timing.Deadlock)
	return false
}

// Not supported by this channel type, autopanic
func (self *bigparallelreceiver) Probe(args ...float64) (*big.Int, float64) {
	panic(timing.Deadlock)
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

