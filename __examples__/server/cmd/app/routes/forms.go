package routes

import (
	"context"
	"mime/multipart"

	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport/__examples__/server/pkg/types"
	"github.com/go-courier/httptransport/httpx"

	"github.com/go-courier/httptransport"
)

var FormsRouter = courier.NewRouter(httptransport.Group("/forms"))

func init() {
	RootRouter.Register(FormsRouter)

	FormsRouter.Register(courier.NewRouter(FormURLEncoded{}))
	FormsRouter.Register(courier.NewRouter(FormMultipartWithFile{}))
	FormsRouter.Register(courier.NewRouter(FormMultipartWithFiles{}))
}

// Form URL Encoded
type FormURLEncoded struct {
	httpx.MethodPost
	FormData struct {
		String string   `name:"string"`
		Slice  []string `name:"slice"`
		Data   Data     `name:"data"`
	} `in:"body" mime:"urlencoded"`
}

func (FormURLEncoded) Path() string {
	return "/urlencoded"
}

func (req FormURLEncoded) Output(ctx context.Context) (resp interface{}, err error) {
	return
}

// Form Multipart
type FormMultipartWithFile struct {
	httpx.MethodPost
	FormData struct {
		Map map[types.Protocol]int `name:"map,omitempty"`
		// @deprecated
		String string                `name:"string,omitempty"`
		Slice  []string              `name:"slice,omitempty"`
		Data   Data                  `name:"data,omitempty"`
		File   *multipart.FileHeader `name:"file"`
	} `in:"body" mime:"multipart"`
}

func (req FormMultipartWithFile) Path() string {
	return "/multipart"
}

func (req FormMultipartWithFile) Output(ctx context.Context) (resp interface{}, err error) {
	return
}

// Form Multipart With Files
type FormMultipartWithFiles struct {
	httpx.MethodPost
	FormData struct {
		Files []*multipart.FileHeader `name:"files"`
	} `in:"body" mime:"multipart"`
}

func (FormMultipartWithFiles) Path() string {
	return "/multipart-with-files"
}

func (FormMultipartWithFiles) Output(ctx context.Context) (resp interface{}, err error) {
	return
}
