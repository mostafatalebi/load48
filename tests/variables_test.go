package tests

import (
	variable "github.com/mostafatalebi/loadtest/pkg/variables"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVariableExtractor_MustAssertTrue(t *testing.T) {
	vmap := variable.VariableMap{}
	vmap["$username"] = &variable.VariableEntry{
		Type: "string",
		Path: "data.username",
	}
	ct := `{ "data" : { "username" : "robert", "password" : "123456"} }`
	v, err := variable.NewVariableAnalysis(vmap, ct, "json")
	assert.NoError(t, err)
	ve := v.Extract()
	assert.NotNil(t, ve)
	assert.Equal(t, 1, len(ve))
	_, ok := ve["$username"]
	assert.True(t, ok)
	val := ve["$username"].Value
	assert.Equal(t, "robert", val)
}

func TestVariableExtractor_MustAssertFalse(t *testing.T) {
	vmap := variable.VariableMap{}
	vmap["$username"] = &variable.VariableEntry{
		Type: "string",
		Path: "data.wrongUsername",
	}
	ct := `{ "data" : { "username" : "robert", "password" : "123456"} }`
	v, err := variable.NewVariableAnalysis(vmap, ct, "json")
	assert.NoError(t, err)
	ve := v.Extract()
	assert.Len(t, ve, 0)
}

func TestVariableExtractorInt_MustAssertFalse(t *testing.T) {
	vmap := variable.VariableMap{}
	vmap["$username"] = &variable.VariableEntry{
		Type: "string",
		Path: "data.username",
	}
	ct := `{ "data" : { "username" : "robert", "password" : "123456"} }`
	v, err := variable.NewVariableAnalysis(vmap, ct, "json")
	assert.NoError(t, err)
	ve := v.Extract()
	assert.NotNil(t, ve)
	assert.Equal(t, 1, len(ve))
	_, ok := ve["$username"]
	assert.True(t, ok)
}

func TestVariableExtractorInt_MustAssertTypeFalse(t *testing.T) {
	vmap := variable.VariableMap{}
	vmap["$username"] = &variable.VariableEntry{
		Type: "string",
		Path: "data.username",
	}
	ct := `{ "data" : { "username" : 123456, "password" : ""} }`
	v, err := variable.NewVariableAnalysis(vmap, ct, "json")
	assert.NoError(t, err)
	ve := v.Extract()
	assert.Len(t, ve, 0)
}

func TestVariableExtractorArr_MustAssertTrue(t *testing.T) {
	vmap := variable.VariableMap{}
	vmap["$names"] = &variable.VariableEntry{
		Type: "array",
		Path: "data.names",
	}
	ct := `{ "data" : { "names" : ["robert", "john"], "password" : ""} }`
	v, err := variable.NewVariableAnalysis(vmap, ct, "json")
	assert.NoError(t, err)
	ve := v.Extract()
	assert.Len(t, ve, 1)
	vv := ve["$names"].Value
	assert.Len(t, vv, 2)
	if vv != "" {
		assert.Equal(t, "robert", vv)
	}
}
