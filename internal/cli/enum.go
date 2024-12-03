package cli

import (
	"fmt"

	"github.com/spf13/pflag"
)

type Enum struct {
	flag    *string
	_type   string
	options []string
}

var _ pflag.Value = &Enum{}

// NewEnum creates a enumerated flag that updates the provided string variable when set.
// The first argument is a hint for use in help text. The second argument is the default value, or "" if there is no default. The remaining arguments are the valid options (a provided default value will be added to the list of options).
func NewEnum(flag *string, _type string, defaultValue string, options ...string) *Enum {
	if defaultValue != "" {
		options = append([]string{defaultValue}, options...)
		*flag = defaultValue
	}
	return &Enum{flag: flag, _type: _type, options: options}
}

func (e *Enum) String() string {
	return *e.flag
}

func (e *Enum) Set(value string) error {
	for _, option := range e.options {
		if option == value {
			*e.flag = value
			return nil
		}
	}
	return fmt.Errorf("invalid value %s", value)
}

func (e *Enum) Type() string {
	return e._type
}
