package httpx

import (
	"fmt"
)

func ExampleImageGIF() {
	m := NewImageGIF()

	fmt.Println(m.ContentType())
	// Output:
	// image/gif
}

func ExampleImageSVG() {
	m := NewImageSVG()

	fmt.Println(m.ContentType())
	// Output:
	// image/svg+xml
}

func ExampleAudioWebm() {
	m := NewAudioWebm()

	fmt.Println(m.ContentType())
	// Output:
	// audio/webm
}

func ExampleAudioMp3() {
	m := NewAudioMp3()

	fmt.Println(m.ContentType())
	// Output:
	// audio/mpeg
}

func ExampleCSS() {
	m := NewCSS()

	fmt.Println(m.ContentType())
	// Output:
	// text/css
}

func ExampleHTML() {
	m := NewHTML()

	fmt.Println(m.ContentType())
	// Output:
	// text/html
}

func ExampleImageWebp() {
	m := NewImageWebp()

	fmt.Println(m.ContentType())
	// Output:
	// image/webp
}

func ExampleVideoOgg() {
	m := NewVideoOgg()

	fmt.Println(m.ContentType())
	// Output:
	// video/ogg
}

func ExampleImageJPEG() {
	m := NewImageJPEG()

	fmt.Println(m.ContentType())
	// Output:
	// image/jpeg
}

func ExampleImagePNG() {
	m := NewImagePNG()

	fmt.Println(m.ContentType())
	// Output:
	// image/png
}

func ExampleImageBmp() {
	m := NewImageBmp()

	fmt.Println(m.ContentType())
	// Output:
	// image/bmp
}

func ExampleAudioOgg() {
	m := NewAudioOgg()

	fmt.Println(m.ContentType())
	// Output:
	// audio/ogg
}

func ExampleApplicationOgg() {
	m := NewApplicationOgg()

	fmt.Println(m.ContentType())
	// Output:
	// application/ogg
}

func ExamplePlain() {
	m := NewPlain()

	fmt.Println(m.ContentType())
	// Output:
	// text/plain
}

func ExampleAudioWave() {
	m := NewAudioWave()

	fmt.Println(m.ContentType())
	// Output:
	// audio/wav
}

func ExampleAudioMidi() {
	m := NewAudioMidi()

	fmt.Println(m.ContentType())
	// Output:
	// audio/midi
}

func ExampleVideoWebm() {
	m := NewVideoWebm()

	fmt.Println(m.ContentType())
	// Output:
	// video/webm
}
