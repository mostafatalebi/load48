package tests

import (
	"github.com/mostafatalebi/loadtest/pkg/assertions"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAssertionsChainRunner(t *testing.T) {
	ass := assertions.NewAssertionManager(map[string]assertions.Assertion{
		assertions.AssertBodyString : assertions.NewAssertionFromName(assertions.AssertBodyString),
		assertions.AssertStatusIsOk : assertions.NewAssertionFromName(assertions.AssertStatusIsOk),
	})
	var err error
	err = ass.Get(assertions.AssertBodyString).SetInput("i am body from test")
	assert.NoError(t, err)
	err = ass.Get(assertions.AssertBodyString).SetTest("test")
	assert.NoError(t, err)
	err = ass.Get(assertions.AssertStatusIsOk).SetInput([]int{8,16,32,64})
	assert.NoError(t, err)
	err = ass.Get(assertions.AssertStatusIsOk).SetTest(32)
	assert.NoError(t, err)
	err = ass.ChainRunner(assertions.AssertStatusIsOk, assertions.AssertBodyString)
	assert.NoError(t, err)
}

func TestAssertionsChainRunner_musReturnError(t *testing.T) {
	ass := assertions.NewAssertionManager(map[string]assertions.Assertion{
		assertions.AssertBodyString : assertions.NewAssertionFromName(assertions.AssertBodyString),
		assertions.AssertStatusIsOk : assertions.NewAssertionFromName(assertions.AssertStatusIsOk),
	})
	var err error
	err = ass.Get(assertions.AssertBodyString).SetInput("i am body from test")
	assert.NoError(t, err)
	err = ass.Get(assertions.AssertBodyString).SetTest("test")
	assert.NoError(t, err)
	err = ass.Get(assertions.AssertStatusIsOk).SetInput([]int{8,16,32,64})
	assert.NoError(t, err)
	err = ass.Get(assertions.AssertStatusIsOk).SetTest(800)
	assert.NoError(t, err)
	err = ass.ChainRunner(assertions.AssertStatusIsOk, assertions.AssertBodyString)
	assert.Error(t, err)
	assert.Equal(t, "failed to assert that 'test' exists in the 'input'", err.Error())
}

func TestAssertionsChainRunner_musSkipNonExistingAssertion(t *testing.T) {
	ass := assertions.NewAssertionManager(map[string]assertions.Assertion{
		//assertions.AssertBodyString : assertions.NewAssertionFromName(assertions.AssertBodyString),
		assertions.AssertStatusIsOk : assertions.NewAssertionFromName(assertions.AssertStatusIsOk),
	})
	var err error
	err = ass.Get(assertions.AssertStatusIsOk).SetInput([]int{8,16,32,64})
	assert.NoError(t, err)
	err = ass.Get(assertions.AssertStatusIsOk).SetTest(8)
	assert.NoError(t, err)
	err = ass.ChainRunner(assertions.AssertStatusIsOk, assertions.AssertBodyString)
	assert.NoError(t, err)
}
