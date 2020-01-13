package client_demo

import (
	bytes "bytes"
	database_sql_driver "database/sql/driver"
	errors "errors"

	github_com_go_courier_courier "github.com/go-courier/courier"
	github_com_go_courier_enumeration "github.com/go-courier/enumeration"
	github_com_go_courier_httptransport_httpx "github.com/go-courier/httptransport/httpx"
	github_com_go_courier_statuserror "github.com/go-courier/statuserror"
)

type BytesBuffer = bytes.Buffer

type Data struct {
	ID        string   `json:"id"`
	Label     string   `json:"label"`
	Protocol  Protocol `json:"protocol,omitempty"`
	PtrString *string  `json:"ptrString,omitempty"`
	SubData   *SubData `json:"subData,omitempty"`
}

type GithubComGoCourierCourierResult = github_com_go_courier_courier.Result

type GithubComGoCourierHttptransportHttpxAttachment = github_com_go_courier_httptransport_httpx.Attachment

type GithubComGoCourierHttptransportHttpxImagePNG = github_com_go_courier_httptransport_httpx.ImagePNG

type GithubComGoCourierHttptransportHttpxStatusFound = github_com_go_courier_httptransport_httpx.StatusFound

type GithubComGoCourierStatuserrorErrorField = github_com_go_courier_statuserror.ErrorField

type GithubComGoCourierStatuserrorErrorFields = github_com_go_courier_statuserror.ErrorFields

type GithubComGoCourierStatuserrorStatusErr = github_com_go_courier_statuserror.StatusErr

type IpInfo struct {
	Country     string `json:"country" xml:"country"`
	CountryCode string `json:"countryCode" xml:"countryCode"`
}

type Protocol = DemoProtocol

type SubData struct {
	Name string `json:"name"`
}

// openapi:enum
type DemoProtocol int

const (
	DEMO_PROTOCOL_UNKNOWN DemoProtocol = iota
	DEMO_PROTOCOL__HTTP                // http
	DEMO_PROTOCOL__HTTPS               // https
)

const (
	DEMO_PROTOCOL__TCP DemoProtocol = iota + 6 // TCP
)

func init() {
	github_com_go_courier_enumeration.DefaultEnumMap.Register(DEMO_PROTOCOL_UNKNOWN)
}

var InvalidDemoProtocol = errors.New("invalid DemoProtocol type")

func ParseDemoProtocolFromLabelString(s string) (DemoProtocol, error) {
	switch s {
	case "":
		return DEMO_PROTOCOL_UNKNOWN, nil
	case "http":
		return DEMO_PROTOCOL__HTTP, nil
	case "https":
		return DEMO_PROTOCOL__HTTPS, nil
	case "TCP":
		return DEMO_PROTOCOL__TCP, nil
	}
	return DEMO_PROTOCOL_UNKNOWN, InvalidDemoProtocol
}

func (v DemoProtocol) String() string {
	switch v {
	case DEMO_PROTOCOL_UNKNOWN:
		return ""
	case DEMO_PROTOCOL__HTTP:
		return "HTTP"
	case DEMO_PROTOCOL__HTTPS:
		return "HTTPS"
	case DEMO_PROTOCOL__TCP:
		return "TCP"
	}
	return "UNKNOWN"
}

func ParseDemoProtocolFromString(s string) (DemoProtocol, error) {
	switch s {
	case "":
		return DEMO_PROTOCOL_UNKNOWN, nil
	case "HTTP":
		return DEMO_PROTOCOL__HTTP, nil
	case "HTTPS":
		return DEMO_PROTOCOL__HTTPS, nil
	case "TCP":
		return DEMO_PROTOCOL__TCP, nil
	}
	return DEMO_PROTOCOL_UNKNOWN, InvalidDemoProtocol
}

func (v DemoProtocol) Label() string {
	switch v {
	case DEMO_PROTOCOL_UNKNOWN:
		return ""
	case DEMO_PROTOCOL__HTTP:
		return "http"
	case DEMO_PROTOCOL__HTTPS:
		return "https"
	case DEMO_PROTOCOL__TCP:
		return "TCP"
	}
	return "UNKNOWN"
}

func (v DemoProtocol) Int() int {
	return int(v)
}

func (DemoProtocol) TypeName() string {
	return "DemoProtocol"
}

func (DemoProtocol) ConstValues() []github_com_go_courier_enumeration.Enum {
	return []github_com_go_courier_enumeration.Enum{DEMO_PROTOCOL__HTTP, DEMO_PROTOCOL__HTTPS, DEMO_PROTOCOL__TCP}
}

func (v DemoProtocol) MarshalText() ([]byte, error) {
	str := v.String()
	if str == "UNKNOWN" {
		return nil, InvalidDemoProtocol
	}
	return []byte(str), nil
}

func (v *DemoProtocol) UnmarshalText(data []byte) (err error) {
	*v, err = ParseDemoProtocolFromString(string(bytes.ToUpper(data)))
	return
}

func (v DemoProtocol) Value() (database_sql_driver.Value, error) {
	offset := 0
	if o, ok := (interface{})(v).(github_com_go_courier_enumeration.EnumDriverValueOffset); ok {
		offset = o.Offset()
	}
	return int64(v) + int64(offset), nil
}

func (v *DemoProtocol) Scan(src interface{}) error {
	offset := 0
	if o, ok := (interface{})(v).(github_com_go_courier_enumeration.EnumDriverValueOffset); ok {
		offset = o.Offset()
	}

	i, err := github_com_go_courier_enumeration.ScanEnum(src, offset)
	if err != nil {
		return err
	}
	*v = DemoProtocol(i)
	return nil

}
