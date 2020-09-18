package assertions

import (
	"errors"
	"strings"
)

type AssertionBodyString struct {
	input string
	test string

}

func (a *AssertionBodyString) SetInput(input interface{}) error {
	v, ok := input.(string)
	if ok {
		a.input = v
		return nil
	} else if vv, ok := input.([]byte); ok {
		a.input = string(vv)
		return nil
	} else {
		return errors.New("input must be string for body-string assertion")
	}
}
func (a *AssertionBodyString) SetTest(test interface{}) error {
	v, ok := test.(string)
	if !ok {
		return errors.New("test must be string for body-string assertion")
	}
	a.test = v
	return nil
}

func (a *AssertionBodyString) Assert() error {
	if strings.Contains(a.input, a.test) {
		return nil
	}
	return errors.New("failed to assert that 'test' exists in the 'input'")
}
