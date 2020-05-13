package transformers

import (
	"bytes"
	"context"
	"io"
	"mime"
	"mime/multipart"
	"net/textproto"
	"reflect"
	"testing"

	"github.com/go-courier/ptr"
	"github.com/go-courier/reflectx/typesutil"
	"github.com/stretchr/testify/require"
)

func TestMultipartTransformer(t *testing.T) {
	boundary := "482e04792f538c09f4dafe5a8c5d792c214b257e5126c4092bf5232cdcbf"

	parts := `--482e04792f538c09f4dafe5a8c5d792c214b257e5126c4092bf5232cdcbf
Content-Disposition: form-data; name="PtrBool"
Content-Type: text/plain; charset=utf-8

true
--482e04792f538c09f4dafe5a8c5d792c214b257e5126c4092bf5232cdcbf
Content-Disposition: form-data; name="PtrInt"
Content-Type: text/plain; charset=utf-8

1
--482e04792f538c09f4dafe5a8c5d792c214b257e5126c4092bf5232cdcbf
Content-Disposition: form-data; name="Bool"
Content-Type: text/plain; charset=utf-8

true
--482e04792f538c09f4dafe5a8c5d792c214b257e5126c4092bf5232cdcbf
Content-Disposition: form-data; name="bytes"
Content-Type: text/plain; charset=utf-8

bytes
--482e04792f538c09f4dafe5a8c5d792c214b257e5126c4092bf5232cdcbf
Content-Disposition: form-data; name="first_name"
Content-Type: text/plain; charset=utf-8

test
--482e04792f538c09f4dafe5a8c5d792c214b257e5126c4092bf5232cdcbf
Content-Disposition: form-data; name="StructSlice"
Content-Type: application/json; charset=utf-8

{"Name":"name"}

--482e04792f538c09f4dafe5a8c5d792c214b257e5126c4092bf5232cdcbf
Content-Disposition: form-data; name="StringSlice"
Content-Type: text/plain; charset=utf-8

1
--482e04792f538c09f4dafe5a8c5d792c214b257e5126c4092bf5232cdcbf
Content-Disposition: form-data; name="StringSlice"
Content-Type: text/plain; charset=utf-8

2
--482e04792f538c09f4dafe5a8c5d792c214b257e5126c4092bf5232cdcbf
Content-Disposition: form-data; name="StringSlice"
Content-Type: text/plain; charset=utf-8

3
--482e04792f538c09f4dafe5a8c5d792c214b257e5126c4092bf5232cdcbf
Content-Disposition: form-data; name="StringArray"
Content-Type: text/plain; charset=utf-8

1
--482e04792f538c09f4dafe5a8c5d792c214b257e5126c4092bf5232cdcbf
Content-Disposition: form-data; name="StringArray"
Content-Type: text/plain; charset=utf-8


--482e04792f538c09f4dafe5a8c5d792c214b257e5126c4092bf5232cdcbf
Content-Disposition: form-data; name="StringArray"
Content-Type: text/plain; charset=utf-8

3
--482e04792f538c09f4dafe5a8c5d792c214b257e5126c4092bf5232cdcbf
Content-Disposition: form-data; name="Struct"
Content-Type: application/xml; charset=utf-8

<Sub><Name></Name></Sub>
--482e04792f538c09f4dafe5a8c5d792c214b257e5126c4092bf5232cdcbf
Content-Disposition: form-data; name="Files"; filename="file0.txt"
Content-Type: application/octet-stream

text0
--482e04792f538c09f4dafe5a8c5d792c214b257e5126c4092bf5232cdcbf
Content-Disposition: form-data; name="Files"; filename="file1.txt"
Content-Type: application/octet-stream

text1
--482e04792f538c09f4dafe5a8c5d792c214b257e5126c4092bf5232cdcbf
Content-Disposition: form-data; name="File"; filename="file.txt"
Content-Type: application/octet-stream

text
--482e04792f538c09f4dafe5a8c5d792c214b257e5126c4092bf5232cdcbf--`

	fixMultipart := func(generatedBoundary string) string {
		buf := bytes.NewBuffer(nil)
		m := multipart.NewWriter(buf)
		_ = m.SetBoundary(generatedBoundary)
		reader := multipart.NewReader(bytes.NewBufferString(parts), boundary)

		part, err := reader.NextPart()
		for err != io.EOF {
			p, _ := m.CreatePart(part.Header)
			_, _ = io.Copy(p, part)
			part, err = reader.NextPart()
		}

		m.Close()
		return buf.String()
	}

	type Sub struct {
		Name string
	}

	type TestData struct {
		PtrBoolEmpty *bool `name:",omitempty"`
		PtrBool      *bool `name:",omitempty"`
		PtrInt       *int
		Bool         bool
		Bytes        []byte `name:"bytes"`
		FirstName    string `name:"first_name,omitempty"`
		StructSlice  []Sub
		StringSlice  []string
		StringArray  [3]string
		Struct       Sub                     `mime:"xml"`
		Files        []*multipart.FileHeader `name:",omitempty"`
		File         *multipart.FileHeader   `name:",omitempty"`
	}

	data := TestData{}
	data.PtrBool = ptr.Bool(true)
	data.FirstName = "test"
	data.Bool = true
	data.Bytes = []byte("bytes")
	data.PtrInt = ptr.Int(1)
	data.StringSlice = []string{"1", "2", "3"}
	data.StructSlice = []Sub{
		{
			Name: "name",
		},
	}
	data.StringArray = [3]string{"1", "", "3"}

	data.File, _ = NewFileHeader("File", "file.txt", bytes.NewBufferString("text"))

	data.Files = []*multipart.FileHeader{
		MustNewFileHeader("Files", "file0.txt", bytes.NewBufferString("text0")),
		MustNewFileHeader("Files", "file1.txt", bytes.NewBufferString("text1")),
	}

	ct, _ := TransformerMgrDefault.NewTransformer(context.Background(), typesutil.FromRType(reflect.TypeOf(data)), TransformerOption{
		MIME: "multipart",
	})

	{
		b := bytes.NewBuffer(nil)
		contentType, err := ct.EncodeToWriter(b, data)
		require.NoError(t, err)
		_, params, _ := mime.ParseMediaType(contentType)
		require.Equal(t, fixMultipart(params["boundary"]), b.String())
	}

	{
		b := bytes.NewBufferString(parts)
		testData := TestData{}

		err := ct.DecodeFromReader(b, &testData, textproto.MIMEHeader{
			"Content-type": []string{
				mime.FormatMediaType(ct.String(), map[string]string{
					"boundary": boundary,
				}),
			},
		})

		require.NoError(t, err)
		require.Equal(t, data, testData)
	}
}
