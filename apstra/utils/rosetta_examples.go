package utils

import (
	"fmt"
	"github.com/orsinium-labs/enum"
)

// ExampleEnumOne is the sort of thing which ordinarily
// would be found in the apstra-go-sdk package
type ExampleEnumOne enum.Member[string]

var (
	ExampleEnumOneFoo  = ExampleEnumOne{Value: "foo"} // friendly value: FOO
	ExampleEnumOneBar  = ExampleEnumOne{Value: "bar"} // friendly value: BAR
	ExampleEnumOneBaz  = ExampleEnumOne{Value: "baz"} // friendly value: baz
	ExampleEnumOneVals = enum.New(ExampleEnumOneFoo, ExampleEnumOneBar, ExampleEnumOneBaz)
)

func (o ExampleEnumOne) String() string {
	return o.Value
}

func (o *ExampleEnumOne) FromString(s string) error {
	t := ExampleEnumOneVals.Parse(s)
	if t == nil {
		return fmt.Errorf("failed to parse ExampleEnumOne %q", s)
	}
	o.Value = t.Value
	return nil
}

// ExampleEnumTwo is the sort of thing which ordinarily
// would be found in the apstra-go-sdk package
type ExampleEnumTwo enum.Member[string]

var (
	ExampleEnumTwoFoo  = ExampleEnumTwo{Value: "foo"} // friendly value: Foo or _foo_
	ExampleEnumTwoBar  = ExampleEnumTwo{Value: "bar"} // friendly value: Bar or _bar_
	ExampleEnumTwoBaz  = ExampleEnumTwo{Value: "baz"} // friendly value: baz
	ExampleEnumTwoVals = enum.New(ExampleEnumTwoFoo, ExampleEnumTwoBar, ExampleEnumTwoBaz)
)

func (o ExampleEnumTwo) String() string {
	return o.Value
}

func (o *ExampleEnumTwo) FromString(s string) error {
	t := ExampleEnumTwoVals.Parse(s)
	if t == nil {
		return fmt.Errorf("failed to parse ExampleEnumTwo %q", s)
	}
	o.Value = t.Value
	return nil
}
