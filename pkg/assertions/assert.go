package assertions

const (
	AssertStatusIsOk = "status-is-ok"
	AssertBodyString = "body-string"
)

var ListOfAssertions = map[string]Assertion{
	AssertBodyString : &AssertionBodyString{},
	AssertStatusIsOk : &AssertionStatusIsOk{
		input: []int{200,201},
	},
}

var DefaultAssertions = []string{AssertStatusIsOk}

type Assertion interface {
	SetInput(inp interface{}) error
	SetTest(test interface{}) error
	Assert() error
}

func NewAssertionFromName(assertName string) Assertion {
	if v, ok := ListOfAssertions[assertName]; ok {
		return v
	}
	return nil
}


