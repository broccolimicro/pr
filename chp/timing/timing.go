package timing

import "errors"

const (
	MIN = 0
	MAX = 1
	AVG = 2
)

var Deadlock error = errors.New("Deadlock")
var Conflict error = errors.New("Conflict")

type Value[vtype interface{}] struct {
	T float64
	V vtype
}

type Action[T interface{}] chan Value[T]

func (p Action[T]) C() chan Value[T] {
	return chan Value[T](p)
}

func (p Action[T]) Recv() (T, float64) {
	a, ok := <-p
	if !ok {
		panic(Deadlock)
	}
	return a.V, a.T
}

type Signal chan float64

func (p Signal) C() chan float64 {
	return chan float64(p)
}

func (p Signal) Send() float64 {
	a, ok := <-p
	if !ok {
		panic(Deadlock)
	}
	return a
}

type TimingSet interface {
	Get() float64
	Add(value interface{})
}

type set struct {
	promises []Signal
	value float64
	count int
	op int
}

func Min(args ...interface{}) TimingSet {
	result := &set{
		op: MIN,
	}
	for _, arg := range args {
		result.Add(arg)
	}

	return result
}

func Max(args ...interface{}) TimingSet {
	result := &set{
		op: MAX,
	}
	for _, arg := range args {
		result.Add(arg)
	}

	return result
}

func Avg(args ...interface{}) TimingSet {
	result := &set{
		op: AVG,
	}
	for _, arg := range args {
		result.Add(arg)
	}

	return result
}

func (t *set) Add(value interface{}) {
	if f, ok := value.(float64); ok {
		if t.count == 0 || (t.op == MAX && f > t.value) || (t.op == MIN && f < t.value) {
			t.value = f
		} else if t.op == AVG {
			t.value += f
		}
		t.count++
	} else if s, ok := value.(Signal); ok {
		t.promises = append(t.promises, s)
	}
}

func (t *set) Get() float64 {
	for _, p := range t.promises {
		t.Add(p.Send())
	}
	t.promises = []Signal{}
	if t.op == AVG {
		return t.value / float64(t.count)
	}
	return t.value
}

