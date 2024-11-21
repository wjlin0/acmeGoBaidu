package baidu

// Type is a string that identifies a particular challenge type and version of ACME challenge.
type Type string

func (t Type) String() string {
	return string(t)
}
