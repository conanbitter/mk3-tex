package main

import (
	"image"
	"log"
	"os"

	_ "image/jpeg"
	_ "image/png"
)

func LoadImage(filename string) ([]IntColor, error) {
	imgFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer imgFile.Close()
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	result := make([]IntColor, 0, (bounds.Max.X-bounds.Min.X)*(bounds.Max.Y-bounds.Min.Y))
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			result = append(result, IntColor{int(r / 257), int(g / 257), int(b / 257)}.Normalized())
		}
	}
	return result, nil
}

func NormalizeIndexed(image []int) (result []uint8) {
	result = make([]uint8, len(image))
	for i, pixel := range image {
		if pixel < 0 {
			pixel = 0
		}
		if pixel > 255 {
			pixel = 255
		}
		result[i] = uint8(pixel)
	}
	return
}

func ConvertImage(inputImage []IntColor, width int, height int, palette any, indexer ImageIndexer) []uint8 {
	var pal Palette
	switch palt := palette.(type) {
	case Palette:
		pal = palt
	case string:
		pal = PaletteLoad(palt)
	default:
		log.Fatal("Wrong palette type")
	}

	return NormalizeIndexed(indexer(inputImage, pal, width, height))
}
