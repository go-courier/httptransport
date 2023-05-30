package client_degradation_demo

import (
	git_querycap_com_cloudchain_common_def_miscs "git.querycap.com/cloudchain/common-def/miscs"
)

type DemoApiResp struct {
	Info GitQuerycapComCloudchainCommonDefMiscsTL0 `json:"info"`
}

type GitQuerycapComCloudchainCommonDefMiscsTL0 = git_querycap_com_cloudchain_common_def_miscs.TL0
