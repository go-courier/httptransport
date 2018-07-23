package httpx

import (
	"bytes"

	"github.com/go-courier/courier"
)

func NewAttachment(filename string, contentType string) *Attachment {
	return &Attachment{
		filename:    filename,
		contentType: contentType,
	}
}

type Attachment struct {
	filename    string
	contentType string
	bytes.Buffer
}

func (a *Attachment) ContentType() string {
	if a.contentType == "" {
		return MIME_OCTET_STREAM
	}
	return a.contentType
}

func (a *Attachment) Meta() courier.Metadata {
	metadata := courier.Metadata{}
	metadata.Add(HeaderContentDisposition, "attachment; filename="+a.filename)
	return metadata
}
