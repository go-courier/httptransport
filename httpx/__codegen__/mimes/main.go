package main

import (
	"sort"

	"github.com/go-courier/codegen"
)

// https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types
var mineTypes = map[string]string{
	"text/plain": "Plain",
	"text/css":   "CSS",
	"text/html":  "HTML",

	"image/gif":     "ImageGIF",
	"image/jpeg":    "ImageJPEG",
	"image/png":     "ImagePNG",
	"image/bmp":     "ImageBmp",
	"image/webp":    "ImageWebp",
	"image/svg+xml": "ImageSVG",

	"audio/wav":  "AudioWave",
	"audio/midi": "AudioMidi",

	"audio/webm": "AudioWebm",
	"video/webm": "VideoWebm",

	"audio/ogg":       "AudioOgg",
	"video/ogg":       "VideoOgg",
	"application/ogg": "ApplicationOgg",

	"audio/mpeg": "AudioMp3",
}

func main() {
	{
		file := codegen.NewFile("httpx", codegen.GeneratedFileSuffix("./mines.go"))

		keys := make([]string, 0)

		for mineType := range mineTypes {
			keys = append(keys, mineType)
		}

		sort.Strings(keys)

		for _, mineType := range keys {
			mineName := mineTypes[mineType]

			typ := codegen.Type(mineName)

			file.WriteBlock(
				file.Expr(`func New?() *? {
	return &?{}
}`, typ, typ, typ),
				file.Expr(`type ? struct {
	`+file.Use("bytes", "Buffer")+`
}`, typ),
				file.Expr(`func (?) ContentType() string {
	return ?
}`, typ, file.Val(mineType)),
			)
		}

		file.WriteFile()
	}

	{
		testFile := codegen.NewFile("httpx", codegen.GeneratedFileSuffix("./mines_test.go"))

		for mineType, mineName := range mineTypes {
			typ := codegen.Type(mineName)

			testFile.WriteBlock(
				testFile.Expr(`func Example?() {
	m := New?()

	`+testFile.Use("fmt", "Println")+`(m.ContentType())
	// Output:
	// `+mineType+`
}`, typ, typ),
			)
		}

		testFile.WriteFile()
	}
}
