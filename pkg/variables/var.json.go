package variable

import (
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"regexp"
)

type VariableJson struct {

}

func NewJsonVariableParser() *VariableJson {
	return &VariableJson{}
}

func (v *VariableJson) ParseString(content, path string) (string, error) {
	r := gjson.Get(content, path)
	if r.Exists() {
		if v, ok := r.Value().(string); ok {
			return v, nil
		}
		return "", errors.New(fmt.Sprintf("key %v is not string", path))
	}
	return "", errors.New(fmt.Sprintf("key %v does not exist", path))
}

func (v *VariableJson) ParseNumber(content, path string) (string, error) {
	r := gjson.Get(content, path)
	if r.Exists() {
		reg := regexp.MustCompile(`^[0-9\.]+$`)
		if v, ok := r.Value().(string); ok {
			if reg.MatchString(v) {
				return v, nil
			}
		}
		return "", errors.New(fmt.Sprintf("key %v is not int", path))
	}
	return "", errors.New(fmt.Sprintf("key %v does not exist", path))
}

func (v *VariableJson) ParseArray(content, path string) ([]interface{}, error) {
	r := gjson.Get(content, path)
	if r.Exists() {
		if v, ok := r.Value().([]interface{}); ok {
			return v, nil
		}
		return nil, errors.New(fmt.Sprintf("key %v is not int", path))
	}
	return nil, errors.New(fmt.Sprintf("key %v does not exist", path))
}

func (v *VariableJson) ParseObject(content, path string) (map[string]interface{}, error) {
	r := gjson.Get(content, path)
	if r.Exists() {
		if v, ok := r.Value().(map[string]interface{}); ok {
			return v, nil
		}
		return nil, errors.New(fmt.Sprintf("key %v is not int", path))
	}
	return nil, errors.New(fmt.Sprintf("key %v does not exist", path))
}