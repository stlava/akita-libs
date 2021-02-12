package akiuri

import (
	"fmt"
	"strings"
)

const (
	Scheme = "akita://"
)

type ObjectType int

const (
	UNKNOWN_TYPE ObjectType = iota
	SPEC
	TRACE // aka learn session
)

func stringToObjectType(s string) ObjectType {
	switch s {
	case "spec":
		return SPEC
	case "trace":
		return TRACE
	}
	return UNKNOWN_TYPE
}

func (o ObjectType) String() string {
	switch o {
	case SPEC:
		return "spec"
	case TRACE:
		return "trace"
	default:
		return "unknown"
	}
}

type URI struct {
	ServiceName string
	ObjectName  string
	ObjectType  ObjectType
}

func (u URI) String() string {
	return fmt.Sprintf(Scheme+"%s:%s:%s", u.ServiceName, u.ObjectType, u.ObjectName)
}

func (u URI) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}

func (u *URI) UnmarshalText(data []byte) error {
	text := string(data)
	if !strings.HasPrefix(text, Scheme) {
		return fmt.Errorf("%q does not start with %q", text, Scheme)
	}

	parts := strings.Split(text[len(Scheme):], ":")
	if len(parts) != 3 {
		return fmt.Errorf("%q does not have 3 parts", text)
	}

	objT := stringToObjectType(parts[1])
	if objT == UNKNOWN_TYPE {
		return fmt.Errorf("%q is an unknown object type", parts[1])
	}

	u.ServiceName = parts[0]
	u.ObjectType = objT
	u.ObjectName = parts[2]

	return nil
}

func Parse(s string) (URI, error) {
	var u URI
	err := u.UnmarshalText([]byte(s))
	return u, err
}
