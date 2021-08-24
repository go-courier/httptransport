package definition_scanner

//go:generate courier gen enum Enum
type Enum int

const (
	ENUM_UNKNOWN Enum = iota
	ENUM__ONE         // one
	ENUM__TWO         // two
)
