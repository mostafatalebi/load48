package assertions

import (
	"errors"
	"fmt"
)

type AssertionManager struct {
	assertions map[string]Assertion
}

func NewAssertionManagerWithDefaults(assertionsMap map[string]Assertion) *AssertionManager {
	if assertionsMap == nil {
		assertionsMap = make(map[string]Assertion, 0)
	}
	for _, v := range DefaultAssertions {
		assertionsMap[v] = ListOfAssertions[v]
	}


	return &AssertionManager{
		assertions: assertionsMap,
	}
}

func (a *AssertionManager) Exists(name string) bool {
	if a.assertions == nil || len(a.assertions) == 0 {
		return false
	}
	if _, ok := a.assertions[name]; ok {
		return true
	}
	return false
}

func (a *AssertionManager) Run(name string) error {
	if a.Exists(name) {
		return a.assertions[name].Assert()
	}
	return errors.New(fmt.Sprintf("assertion %v found", name))
}

func (a *AssertionManager) Get(name string) Assertion {
	if a.Exists(name) {
		return a.assertions[name]
	}
	return nil
}

// Runs a list of assertions, does not error if one or more
// of them does not exists. If none of the given assertions
// exists, it returns error.
// it also returns error on the first instance of an assertion's
// failure
func (a *AssertionManager) ChainRunner(names ...string) error {
	if names == nil || len(names) == 0 {
		return errors.New("no assertion specified")
	}
	anyExists := false
	for _, v := range names {
		if a.Exists(v) {
			anyExists = true
			if err := a.Run(v); err != nil {
				return err
			}
		}
	}

	if anyExists {
		return nil
	}
	return errors.New("none of the given assertions have been registered")
}
