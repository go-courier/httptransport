package httpx

import (
	fmt "fmt"
)

func ExampleCSS() {
	m := NewCSS()

	fmt.Println(m.ContentType())
	// Output:
	// text/css
}

func ExampleImageWebp() {
	m := NewImageWebp()

	fmt.Println(m.ContentType())
	// Output:
	// image/webp
}

func ExampleVideoWebm() {
	m := NewVideoWebm()

	fmt.Println(m.ContentType())
	// Output:
	// video/webm
}

func ExampleApplicationOgg() {
	m := NewApplicationOgg()

	fmt.Println(m.ContentType())
	// Output:
	// application/ogg
}

func ExampleAudioMp3() {
	m := NewAudioMp3()

	fmt.Println(m.ContentType())
	// Output:
	// audio/mpeg
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

func ExampleAudioMidi() {
	m := NewAudioMidi()

	fmt.Println(m.ContentType())
	// Output:
	// audio/midi
}

func ExampleVideoOgg() {
	m := NewVideoOgg()

	fmt.Println(m.ContentType())
	// Output:
	// video/ogg
}

func ExamplePlain() {
	m := NewPlain()

	fmt.Println(m.ContentType())
	// Output:
	// text/plain
}

func ExampleHTML() {
	m := NewHTML()

	fmt.Println(m.ContentType())
	// Output:
	// text/html
}

func ExampleImageSVG() {
	m := NewImageSVG()

	fmt.Println(m.ContentType())
	// Output:
	// image/svg+xml
}

func ExampleImageGIF() {
	m := NewImageGIF()

	fmt.Println(m.ContentType())
	// Output:
	// image/gif
}

func ExampleAudioWave() {
	m := NewAudioWave()

	fmt.Println(m.ContentType())
	// Output:
	// audio/wav
}

func ExampleAudioWebm() {
	m := NewAudioWebm()

	fmt.Println(m.ContentType())
	// Output:
	// audio/webm
}

func ExampleAudioOgg() {
	m := NewAudioOgg()

	fmt.Println(m.ContentType())
	// Output:
	// audio/ogg
}
