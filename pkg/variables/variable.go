package variable

import (
	"encoding/json"
	"errors"
	"github.com/mostafatalebi/loadtest/pkg/logger"
	"github.com/tidwall/gjson"
)

const (
	CtJson = "json"

	VarString = "string"
	VarNumber = "number"
	VarArr    = "array"
	VarObj    = "object"
)

type VariableMap map[string]*VariableEntry

type VariableEntry struct {
	Type  string `yaml:"type"`
	Path  string `yaml:"path"`
	Value string `yaml:"-"`
}

type VariablesExtracted map[string]interface{}

type VariableParser interface {
	ParseString(content, path string) (string, error)
	ParseNumber(content, path string) (string, error)
	ParseArray(content, path string) ([]interface{}, error)
	ParseObject(content, path string) (map[string]interface{}, error)
}

type VariableAnalysis struct {
	contentType string
	content     string
	parser      VariableParser
	baseMap     VariableMap
}

func NewVariableAnalysis(varMap VariableMap, content, contentType string) (*VariableAnalysis, error) {
	var parser VariableParser
	if contentType == CtJson {
		parser = NewJsonVariableParser()
		if !gjson.Valid(content) {
			return nil, errors.New("wrong json format")
		}
	}
	return &VariableAnalysis{
		content:     content,
		parser:      parser,
		contentType: CtJson,
		baseMap:     varMap,
	}, nil
}

func (v *VariableAnalysis) Extract() VariableMap {
	if v.baseMap != nil {
		ve := VariableMap{}
		for k, vv := range v.baseMap {
			switch vv.Type {
			case VarString:
				vs, err := v.parser.ParseString(v.content, vv.Path)
				if err != nil {
					logger.Error("variable extraction failed", err.Error())
					continue
				}
				vv.Value = vs
				ve[k] = vv
			case VarNumber:
				vs, err := v.parser.ParseNumber(v.content, vv.Path)
				if err != nil {
					logger.Error("variable extraction failed", err.Error())
					continue
				}
				vv.Value = vs
				ve[k] = vv
			case VarArr:
				vs, err := v.parser.ParseArray(v.content, vv.Path)
				if err != nil {
					logger.Error("variable extraction failed", err.Error())
					continue
				}
				sv, err := json.Marshal(vs)
				if err != nil {
					sv = nil
				}
				vv.Value = string(sv)
				ve[k] = vv
			case VarObj:
				vs, err := v.parser.ParseArray(v.content, vv.Path)
				if err != nil {
					logger.Error("variable extraction failed", err.Error())
					continue
				}
				sv, err := json.Marshal(vs)
				if err != nil {
					sv = nil
				}
				vv.Value = string(sv)
				ve[k] = vv
			}
		}
		return ve
	}
	return nil
}

func Merge(vars VariableMap, otherVars VariableMap) VariableMap {
	var newVars = make(VariableMap, 0)
	if vars != nil {
		for k, v := range vars {
			newVars[k] = v
		}
	}
	if otherVars != nil {
		for k, v := range otherVars {
			newVars[k] = v
		}
	}
	return newVars
}
