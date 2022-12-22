package timing

type TimingSet interface {
	Get() float64
	Max(value float64)
	Min(value float64)
	Avg(value float64)
}

type set struct {
	value float64
	count int
}

func Set(args ...float64) TimingSet {
	if len(args) > 0 {
		return &set{
			value: args[0],
		}
	}
	return &set{}
}

func (t *set) Max(value float64) {
	if t.count == 0 || value > t.value {
		t.value = value
	}
	t.count++
}

func (t *set) Min(value float64) {
	if t.count == 0 || value < t.value {
		t.value = value
	}
	t.count++
}

func (t *set) Avg(value float64) {
	t.value = (t.value*float64(t.count) + value)/float64(t.count+1)
	t.count++
}

func (t *set) Get() float64 {
	return t.value
}

func Max(values ...float64) float64 {
	var result float64 = 0.0
	for i, value := range values {
		if i == 0 || value > result {
			result = value
		}
	}
	return result
}

func Min(values ...float64) float64 {
	var result float64 = 0.0
	for i, value := range values {
		if i == 0 || value < result {
			result = value
		}
	}
	return result
}
