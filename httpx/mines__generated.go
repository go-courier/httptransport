package httpx

import (
	"bytes"
)

func NewCSS() *CSS {
	return &CSS{}
}

type CSS struct {
	bytes.Buffer
}

func (CSS) ContentType() string {
	return "text/css"
}

func NewHTML() *HTML {
	return &HTML{}
}

type HTML struct {
	bytes.Buffer
}

func (HTML) ContentType() string {
	return "text/html"
}

func NewImageWebp() *ImageWebp {
	return &ImageWebp{}
}

type ImageWebp struct {
	bytes.Buffer
}

func (ImageWebp) ContentType() string {
	return "image/webp"
}

func NewVideoOgg() *VideoOgg {
	return &VideoOgg{}
}

type VideoOgg struct {
	bytes.Buffer
}

func (VideoOgg) ContentType() string {
	return "video/ogg"
}

func NewImageJPEG() *ImageJPEG {
	return &ImageJPEG{}
}

type ImageJPEG struct {
	bytes.Buffer
}

func (ImageJPEG) ContentType() string {
	return "image/jpeg"
}

func NewImagePNG() *ImagePNG {
	return &ImagePNG{}
}

type ImagePNG struct {
	bytes.Buffer
}

func (ImagePNG) ContentType() string {
	return "image/png"
}

func NewImageBmp() *ImageBmp {
	return &ImageBmp{}
}

type ImageBmp struct {
	bytes.Buffer
}

func (ImageBmp) ContentType() string {
	return "image/bmp"
}

func NewAudioOgg() *AudioOgg {
	return &AudioOgg{}
}

type AudioOgg struct {
	bytes.Buffer
}

func (AudioOgg) ContentType() string {
	return "audio/ogg"
}

func NewApplicationOgg() *ApplicationOgg {
	return &ApplicationOgg{}
}

type ApplicationOgg struct {
	bytes.Buffer
}

func (ApplicationOgg) ContentType() string {
	return "application/ogg"
}

func NewPlain() *Plain {
	return &Plain{}
}

type Plain struct {
	bytes.Buffer
}

func (Plain) ContentType() string {
	return "text/plain"
}

func NewAudioWave() *AudioWave {
	return &AudioWave{}
}

type AudioWave struct {
	bytes.Buffer
}

func (AudioWave) ContentType() string {
	return "audio/wav"
}

func NewAudioMidi() *AudioMidi {
	return &AudioMidi{}
}

type AudioMidi struct {
	bytes.Buffer
}

func (AudioMidi) ContentType() string {
	return "audio/midi"
}

func NewVideoWebm() *VideoWebm {
	return &VideoWebm{}
}

type VideoWebm struct {
	bytes.Buffer
}

func (VideoWebm) ContentType() string {
	return "video/webm"
}

func NewImageGIF() *ImageGIF {
	return &ImageGIF{}
}

type ImageGIF struct {
	bytes.Buffer
}

func (ImageGIF) ContentType() string {
	return "image/gif"
}

func NewImageSVG() *ImageSVG {
	return &ImageSVG{}
}

type ImageSVG struct {
	bytes.Buffer
}

func (ImageSVG) ContentType() string {
	return "image/svg+xml"
}

func NewAudioWebm() *AudioWebm {
	return &AudioWebm{}
}

type AudioWebm struct {
	bytes.Buffer
}

func (AudioWebm) ContentType() string {
	return "audio/webm"
}

func NewAudioMp3() *AudioMp3 {
	return &AudioMp3{}
}

type AudioMp3 struct {
	bytes.Buffer
}

func (AudioMp3) ContentType() string {
	return "audio/mpeg"
}
