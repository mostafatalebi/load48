package assertions

import (
	"errors"
)

type AssertionStatusIsOk struct {
	input []int
	test int

}

func (a *AssertionStatusIsOk) SetInput(input interface{}) error {
	v, ok := input.([]int)
	if ok {
		a.input = v
		return nil
	} else {
		return errors.New("input must be string for body-string assertion")
	}
}
func (a *AssertionStatusIsOk) SetTest(test interface{}) error {
	v, ok := test.(int)
	if !ok {
		return errors.New("test must be string for body-string assertion")
	}
	a.test = v
	return nil
}

func (a *AssertionStatusIsOk) Assert() error {
	if a.input != nil && len(a.input) > 0 {
		for _, v := range a.input {
			if v == a.test {
				return nil
			}
		}
	}
	return errors.New("failed to assert that 'test' exists in the 'input'")
}
