package main

import (
	"image/jpeg"
	"os"

	asc "asciiconvert"
)

func main() {
	file, _ := os.Open("anna.jpg")
	defer file.Close()
	img, _ := jpeg.Decode(file)

	asc.AsciiToConsole(img, 15)

}
