package client

import (
	github_com_go_courier_statuserror "github.com/go-courier/statuserror"
)

var _ interface {
	github_com_go_courier_statuserror.StatusError
} = (*StatusError)(nil)

func (v StatusError) StatusErr() *github_com_go_courier_statuserror.StatusErr {
	return &github_com_go_courier_statuserror.StatusErr{
		Key:            v.Key(),
		Code:           v.Code(),
		Msg:            v.Msg(),
		CanBeTalkError: v.CanBeTalkError(),
	}
}

func (v StatusError) Error() string {
	return v.StatusErr().Error()
}

func (v StatusError) StatusCode() int {
	return github_com_go_courier_statuserror.StatusCodeFromCode(int(v))
}

func (v StatusError) Code() int {
	if withServiceCode, ok := (interface{})(v).(github_com_go_courier_statuserror.StatusErrorWithServiceCode); ok {
		return withServiceCode.ServiceCode() + int(v)
	}
	return int(v)

}

func (v StatusError) Key() string {
	switch v {
	case RequestTransformFailed:
		return "RequestTransformFailed"
	case RequestFailed:
		return "RequestFailed"
	case ReadFailed:
		return "ReadFailed"
	}
	return "UNKNOWN"
}

func (v StatusError) Msg() string {
	switch v {
	case RequestTransformFailed:
		return "transform request failed"
	case RequestFailed:
		return "request failed"
	case ReadFailed:
		return "read failed"
	}
	return "-"
}

func (v StatusError) CanBeTalkError() bool {
	switch v {
	case RequestTransformFailed:
		return false
	case RequestFailed:
		return false
	case ReadFailed:
		return false
	}
	return false
}
