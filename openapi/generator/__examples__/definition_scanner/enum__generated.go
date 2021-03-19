package definition_scanner

import (
	"bytes"
	database_sql_driver "database/sql/driver"

	"github.com/pkg/errors"

	github_com_go_courier_enumeration "github.com/go-courier/enumeration"
)

var InvalidEnum = errors.New("invalid Enum type")

func ParseEnumFromLabelString(s string) (Enum, error) {
	switch s {
	case "":
		return ENUM_UNKNOWN, nil
	case "two":
		return ENUM__TWO, nil
	case "one":
		return ENUM__ONE, nil
	}
	return ENUM_UNKNOWN, InvalidEnum
}

func (v Enum) String() string {
	switch v {
	case ENUM_UNKNOWN:
		return ""
	case ENUM__TWO:
		return "TWO"
	case ENUM__ONE:
		return "ONE"
	}
	return "UNKNOWN"
}

func ParseEnumFromString(s string) (Enum, error) {
	switch s {
	case "":
		return ENUM_UNKNOWN, nil
	case "TWO":
		return ENUM__TWO, nil
	case "ONE":
		return ENUM__ONE, nil
	}
	return ENUM_UNKNOWN, InvalidEnum
}

func (v Enum) Label() string {
	switch v {
	case ENUM_UNKNOWN:
		return ""
	case ENUM__TWO:
		return "two"
	case ENUM__ONE:
		return "one"
	}
	return "UNKNOWN"
}

func (v Enum) Int() int {
	return int(v)
}

func (Enum) TypeName() string {
	return "Enum"
}

func (Enum) ConstValues() []github_com_go_courier_enumeration.Enum {
	return []github_com_go_courier_enumeration.Enum{ENUM__ONE, ENUM__TWO}
}

func (v Enum) MarshalText() ([]byte, error) {
	str := v.String()
	if str == "UNKNOWN" {
		return nil, InvalidEnum
	}
	return []byte(str), nil
}

func (v *Enum) UnmarshalText(data []byte) (err error) {
	*v, err = ParseEnumFromString(string(bytes.ToUpper(data)))
	return
}

func (v Enum) Value() (database_sql_driver.Value, error) {
	offset := 0
	if o, ok := (interface{})(v).(github_com_go_courier_enumeration.EnumDriverValueOffset); ok {
		offset = o.Offset()
	}
	return int(v) + offset, nil
}

func (v *Enum) Scan(src interface{}) error {
	offset := 0
	if o, ok := (interface{})(v).(github_com_go_courier_enumeration.EnumDriverValueOffset); ok {
		offset = o.Offset()
	}

	i, err := github_com_go_courier_enumeration.ScanEnum(src, offset)
	if err != nil {
		return err
	}
	*v = Enum(i)
	return nil

}
