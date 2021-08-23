package types

import (
	"bytes"
	database_sql_driver "database/sql/driver"

	"github.com/pkg/errors"

	github_com_go_courier_enumeration "github.com/go-courier/enumeration"
)

var InvalidProtocol = errors.New("invalid Protocol type")

func ParseProtocolFromLabelString(s string) (Protocol, error) {
	switch s {
	case "":
		return PROTOCOL_UNKNOWN, nil
	case "TCP":
		return PROTOCOL__TCP, nil
	case "https":
		return PROTOCOL__HTTPS, nil
	case "http":
		return PROTOCOL__HTTP, nil
	}
	return PROTOCOL_UNKNOWN, InvalidProtocol
}

func (v Protocol) String() string {
	switch v {
	case PROTOCOL_UNKNOWN:
		return ""
	case PROTOCOL__TCP:
		return "TCP"
	case PROTOCOL__HTTPS:
		return "HTTPS"
	case PROTOCOL__HTTP:
		return "HTTP"
	}
	return "UNKNOWN"
}

func ParseProtocolFromString(s string) (Protocol, error) {
	switch s {
	case "":
		return PROTOCOL_UNKNOWN, nil
	case "TCP":
		return PROTOCOL__TCP, nil
	case "HTTPS":
		return PROTOCOL__HTTPS, nil
	case "HTTP":
		return PROTOCOL__HTTP, nil
	}
	return PROTOCOL_UNKNOWN, InvalidProtocol
}

func (v Protocol) Label() string {
	switch v {
	case PROTOCOL_UNKNOWN:
		return ""
	case PROTOCOL__TCP:
		return "TCP"
	case PROTOCOL__HTTPS:
		return "https"
	case PROTOCOL__HTTP:
		return "http"
	}
	return "UNKNOWN"
}

func (v Protocol) Int() int {
	return int(v)
}

func (Protocol) TypeName() string {
	return "Protocol"
}

func (Protocol) ConstValues() []github_com_go_courier_enumeration.Enum {
	return []github_com_go_courier_enumeration.Enum{PROTOCOL__HTTP, PROTOCOL__HTTPS, PROTOCOL__TCP}
}

func (v Protocol) MarshalText() ([]byte, error) {
	str := v.String()
	if str == "UNKNOWN" {
		return nil, InvalidProtocol
	}
	return []byte(str), nil
}

func (v *Protocol) UnmarshalText(data []byte) (err error) {
	*v, err = ParseProtocolFromString(string(bytes.ToUpper(data)))
	return
}

func (v Protocol) Value() (database_sql_driver.Value, error) {
	offset := 0
	if o, ok := (interface{})(v).(github_com_go_courier_enumeration.EnumDriverValueOffset); ok {
		offset = o.Offset()
	}
	return int(v) + offset, nil
}

func (v *Protocol) Scan(src interface{}) error {
	offset := 0
	if o, ok := (interface{})(v).(github_com_go_courier_enumeration.EnumDriverValueOffset); ok {
		offset = o.Offset()
	}

	i, err := github_com_go_courier_enumeration.ScanEnum(src, offset)
	if err != nil {
		return err
	}
	*v = Protocol(i)
	return nil

}
