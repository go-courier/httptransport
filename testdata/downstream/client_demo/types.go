package client_demo

import (
	bytes "bytes"

	github_com_go_courier_httptransport_httpx "github.com/go-courier/httptransport/httpx"
	github_com_go_courier_httptransport_testdata_server_pkg_types "github.com/go-courier/httptransport/testdata/server/pkg/types"
	github_com_go_courier_statuserror "github.com/go-courier/statuserror"
)

type BytesBuffer = bytes.Buffer

type Data struct {
	ID        string                                                        `json:"id"`
	Label     string                                                        `json:"label"`
	Protocol  GithubComGoCourierHttptransportTestdataServerPkgTypesProtocol `json:"protocol,omitempty"`
	PtrString *string                                                       `json:"ptrString,omitempty"`
	SubData   *SubData                                                      `json:"subData,omitempty"`
}

type GithubComGoCourierHttptransportHttpxAttachment = github_com_go_courier_httptransport_httpx.Attachment

type GithubComGoCourierHttptransportHttpxImagePNG = github_com_go_courier_httptransport_httpx.ImagePNG

type GithubComGoCourierHttptransportHttpxResponse = github_com_go_courier_httptransport_httpx.Response

type GithubComGoCourierHttptransportHttpxStatusFound struct {
	GithubComGoCourierHttptransportHttpxResponse
}

type GithubComGoCourierHttptransportTestdataServerPkgTypesProtocol = github_com_go_courier_httptransport_testdata_server_pkg_types.Protocol

type GithubComGoCourierHttptransportTestdataServerPkgTypesPullPolicy = github_com_go_courier_httptransport_testdata_server_pkg_types.PullPolicy

type GithubComGoCourierStatuserrorErrorField = github_com_go_courier_statuserror.ErrorField

type GithubComGoCourierStatuserrorErrorFields = github_com_go_courier_statuserror.ErrorFields

type GithubComGoCourierStatuserrorStatusErr = github_com_go_courier_statuserror.StatusErr

type IpInfo struct {
	Country     string `json:"country" xml:"country"`
	CountryCode string `json:"countryCode" xml:"countryCode"`
}

type SubData struct {
	Name string `json:"name"`
}
