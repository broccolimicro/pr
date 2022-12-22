package timing

import (
	"os"
)

type Profile interface {
	Find(name string) float64
}

type profile struct {
	values map[string]float64
}

func NewProfile() Profile {
	return &profile{
		values: make(map[string]float64),
	}
}

func (p *profile) Find(name string) float64 {
	delay, ok := p.values[name]
	if ok {
		return delay
	}
	return 0.0
}

type ProfileSet interface {
	Find(name string) Profile
}

type profileSet struct {
	profiles map[string]Profile
}

func NewProfileSet() ProfileSet {
	return &profileSet{
		profiles: make(map[string]Profile),
	}
}

func LoadProfileSet(path string) (ProfileSet, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	p, err := ParseReader(path, file)
	if err != nil {
		return nil, err
	}

	return p.(ProfileSet), nil
}

func (s *profileSet) Find(name string) Profile {
	p, ok := s.profiles[name]
	if ok {
		return p
	}
	return nil
}
