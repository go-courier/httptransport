package httptransport

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"reflect"
	"regexp"
	"testing"

	"github.com/go-courier/reflectx"
	"github.com/go-courier/statuserror"
	"github.com/stretchr/testify/require"

	"github.com/go-courier/httptransport/transformers"
)

var reContentTypeWithBoundary = regexp.MustCompile(`Content-Type: multipart/form-data; boundary=([A-Za-z0-9]+)`)

func UnifyRequestData(data []byte) []byte {
	data = bytes.Replace(data, []byte("\r\n"), []byte("\n"), -1)

	if reContentTypeWithBoundary.Match(data) {
		matches := reContentTypeWithBoundary.FindAllSubmatch(data, 1)
		data = bytes.Replace(data, matches[0][1], []byte("boundary1"), -1)
	}

	return data
}

func TestRequestTransformer(t *testing.T) {
	mgr := NewRequestTransformerMgr(nil, nil)

	type Headers struct {
		HInt    int    `in:"header"`
		HString string `in:"header"`
		HBool   bool   `in:"header"`
	}

	type Queries struct {
		QInt            int      `name:"int" in:"query"`
		QString         string   `name:"string" in:"query"`
		QSlice          []string `name:"slice" in:"query"`
		QBytes          []byte   `name:"bytes,omitempty" in:"query"`
		QBytesOmitEmpty []byte   `name:"bytesOmit,omitempty" in:"query"`
		QBytesKeepEmpty []byte   `name:"bytesKeep" in:"query"`
	}

	type Cookies struct {
		CString string   `name:"a" in:"cookie"`
		CSlice  []string `name:"slice" in:"cookie"`
	}

	type Data struct {
		A string
		B string
		C string
	}

	type FormDataMultipart struct {
		Bytes []byte `name:"bytes"`
		A     []int  `name:"a"`
		C     uint   `name:"c" `
		Data  Data   `name:"data"`

		File  *multipart.FileHeader   `name:"file"`
		Files []*multipart.FileHeader `name:"files"`
	}

	cases := []struct {
		name   string
		path   string
		expect string
		req    interface{}
	}{
		{
			"full parameters",
			"/:id",
			`GET /-?bytes=bytes&bytesKeep=&int=1&slice=1&slice=2&string=string HTTP/1.1
Content-Type: application/json; charset=utf-8
Cookie: a=xxx; slice=1; slice=2
Hbool: true
Hint: 1
Hstring: string

{"A":"","B":"","C":""}
`,
			&struct {
				Headers
				Queries
				Cookies
				Data `in:"body"`
				ID   string `name:"id" in:"path"`
			}{
				Headers: Headers{
					HInt:    1,
					HString: "string",
					HBool:   true,
				},
				Queries: Queries{
					QInt:            1,
					QString:         "string",
					QSlice:          []string{"1", "2"},
					QBytes:          []byte("bytes"),
					QBytesKeepEmpty: []byte{},
				},
				Cookies: Cookies{
					CString: "xxx",
					CSlice:  []string{"1", "2"},
				},
			},
		},
		{
			"url-encoded",
			"/",
			`GET / HTTP/1.1
Content-Type: application/x-www-form-urlencoded; param=value

bytesKeep=&int=1&slice=1&slice=2&string=string`,
			&struct {
				Queries `in:"body" mime:"urlencoded"`
			}{
				Queries: Queries{
					QInt:            1,
					QString:         "string",
					QSlice:          []string{"1", "2"},
					QBytesKeepEmpty: []byte{},
				},
			},
		},
		{
			"xml",
			"/",
			`GET / HTTP/1.1
Content-Type: application/xml; charset=utf-8

<Data><A>1</A><B></B><C></C></Data>`,
			&struct {
				Data `in:"body" mime:"xml"`
			}{
				Data: Data{
					A: "1",
				},
			},
		},
		{
			"form-data/multipart",
			"/",
			`GET / HTTP/1.1
Content-Type: multipart/form-data; boundary=5eaf397248958ac38281d1c034e1ad0d4a5f7d986d4c53ac32e8399cbcda

--5eaf397248958ac38281d1c034e1ad0d4a5f7d986d4c53ac32e8399cbcda
Content-Disposition: form-data; name="bytes"
Content-Type: text/plain; charset=utf-8

bytes
--5eaf397248958ac38281d1c034e1ad0d4a5f7d986d4c53ac32e8399cbcda
Content-Disposition: form-data; name="a"
Content-Type: text/plain; charset=utf-8

-1
--5eaf397248958ac38281d1c034e1ad0d4a5f7d986d4c53ac32e8399cbcda
Content-Disposition: form-data; name="a"
Content-Type: text/plain; charset=utf-8

1
--5eaf397248958ac38281d1c034e1ad0d4a5f7d986d4c53ac32e8399cbcda
Content-Disposition: form-data; name="c"
Content-Type: text/plain; charset=utf-8

1
--5eaf397248958ac38281d1c034e1ad0d4a5f7d986d4c53ac32e8399cbcda
Content-Disposition: form-data; name="data"
Content-Type: application/json; charset=utf-8

{"A":"1","B":"","C":""}

--5eaf397248958ac38281d1c034e1ad0d4a5f7d986d4c53ac32e8399cbcda
Content-Disposition: form-data; name="file"; filename="file.text"
Content-Type: application/octet-stream

test
--5eaf397248958ac38281d1c034e1ad0d4a5f7d986d4c53ac32e8399cbcda
Content-Disposition: form-data; name="files"; filename="file1.text"
Content-Type: application/octet-stream

test1
--5eaf397248958ac38281d1c034e1ad0d4a5f7d986d4c53ac32e8399cbcda
Content-Disposition: form-data; name="files"; filename="file2.text"
Content-Type: application/octet-stream

test2
--5eaf397248958ac38281d1c034e1ad0d4a5f7d986d4c53ac32e8399cbcda--
`,
			&struct {
				FormDataMultipart `in:"body" mime:"multipart" boundary:"boundary1"`
			}{
				FormDataMultipart: FormDataMultipart{
					A:     []int{-1, 1},
					C:     1,
					Bytes: []byte("bytes"),
					Data: Data{
						A: "1",
					},
					Files: []*multipart.FileHeader{
						transformers.MustNewFileHeader("files", "file1.text", bytes.NewBufferString("test1")),
						transformers.MustNewFileHeader("files", "file2.text", bytes.NewBufferString("test2")),
					},
					File: transformers.MustNewFileHeader("file", "file.text", bytes.NewBufferString("test")),
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for i := 0; i < 5; i++ {
				rtForSomeRequest, err := mgr.NewRequestTransformer(reflect.TypeOf(c.req))
				require.NoError(t, err)

				req, err := rtForSomeRequest.NewRequest(http.MethodGet, c.path, c.req)
				require.NoError(t, err)

				data, _ := httputil.DumpRequest(req, true)
				require.Equal(t, string(UnifyRequestData([]byte(c.expect))), string(UnifyRequestData(data)))

				rv := reflectx.New(reflectx.Deref(reflect.TypeOf(c.req)))
				e := rtForSomeRequest.DecodeFromRequestInfo(NewRequestInfo(req), rv)
				require.NoError(t, e)
				require.Equal(t, reflectx.Indirect(reflect.ValueOf(c.req)).Interface(), reflectx.Indirect(rv).Interface())
			}
		})
	}
}

type ReqWithPostValidate struct {
	StartedAt string `in:"query"`
}

func (ReqWithPostValidate) PostValidate(badRequest *BadRequest) {
	badRequest.AddErr(fmt.Errorf("ops"), "query", "StartedAt")
}

func TestRequestTransformer_DecodeFromRequestInfo_Failed(t *testing.T) {
	type Nested struct {
		A string `name:"a" validate:"@string[1,]"`
		B string `name:"b" default:"1" validate:"@string[1,]"`
		C string `name:"c" validate:"@string[2,]?"`
	}

	type Data struct {
		A      string `validate:"@string[1,]"`
		B      string `default:"1" validate:"@string[1,]"`
		C      string `validate:"@string[2,]?"`
		Nested Nested
	}

	cases := []struct {
		name string
		path string
		req  interface{}
	}{
		{
			"validate failed",
			"/:id",
			struct {
				ID      string   `in:"path" name:"id" validate:"@string[2,]"`
				QString string   `in:"query" name:"string,omitempty" default:"11" validate:"@string[2,]"`
				QSlice  []string `in:"query" name:"slice,omitempty" validate:"@slice<@string[2,]>[2,]"`
				Data    `in:"body"`
			}{
				ID:      "1",
				QString: "!",
				QSlice:  []string{"11", "1"},
				Data: Data{
					C: "1",
				},
			},
		},
		{
			"post validate",
			"/:id",
			ReqWithPostValidate{},
		},
	}

	mgr := NewRequestTransformerMgr(nil, nil)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rtForSomeRequest, err := mgr.NewRequestTransformer(reflect.TypeOf(c.req))
			require.NoError(t, err)
			if err != nil {
				return
			}

			{
				_, err := rtForSomeRequest.NewRequest(http.MethodGet, c.path, struct{}{})
				require.Error(t, err)
			}

			req, err := rtForSomeRequest.NewRequest(http.MethodGet, c.path, c.req)
			require.NoError(t, err)

			{
				err := rtForSomeRequest.DecodeFromRequestInfo(NewRequestInfo(req), struct{}{})
				require.Error(t, err)
			}

			rv := reflectx.New(reflectx.Deref(reflect.TypeOf(c.req)))
			e := rtForSomeRequest.DecodeFromRequestInfo(NewRequestInfo(req), rv)
			require.Error(t, e)

			for _, ef := range e.(*statuserror.StatusErr).ErrorFields {
				fmt.Println(ef)
			}
		})
	}
}
