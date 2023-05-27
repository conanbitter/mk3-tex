package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type Texture struct {
	Data   []IntColor
	Width  int
	Height int
	Name   string
}

func main() {
	project := OpenProject("test_assets/project.txt")
	textures := make([]Texture, 0)
	imgdata := make([][]IntColor, 0)
	for _, entry := range project.Textures {
		fmt.Printf("Loading \"%s\" as \"%s\" ...\n", filepath.Base(entry.Filename), entry.Name)
		data, width, height, err := LoadImage(entry.Filename)
		if err != nil {
			log.Fatal(err)
		}
		textures = append(textures, Texture{
			Data:   data,
			Width:  width,
			Height: height,
			Name:   entry.Name,
		})
		imgdata = append(imgdata, data)
	}

	fmt.Println("Calculating palette...")
	palCalc := NewPalCalc(project.Colors, 1000, 10)
	palCalc.Input(imgdata)
	palCalc.Run()
	pal := palCalc.GetPalette()
	pal.Save("palette.json")

	fmt.Println("Saving file...")
	file, err := os.Create("result.txs")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// PALETTE
	binary.Write(file, binary.LittleEndian, uint8(pal.Len()))
	binary.Write(file, binary.LittleEndian, uint8(project.Offset))
	for _, color := range pal {
		binary.Write(file, binary.LittleEndian, uint8(color.R))
		binary.Write(file, binary.LittleEndian, uint8(color.G))
		binary.Write(file, binary.LittleEndian, uint8(color.B))
	}

	// TEXTURES
	binary.Write(file, binary.LittleEndian, uint32(len(textures)))
	for i, tex := range textures {
		fmt.Printf("Adding \"%s\" ...\n", tex.Name)

		var name [16]byte
		copy(name[:], []byte(tex.Name))
		binary.Write(file, binary.LittleEndian, name)

		binary.Write(file, binary.LittleEndian, uint32(tex.Width))
		binary.Write(file, binary.LittleEndian, uint32(tex.Height))

		converted := NormalizeAndOffset(ConvertImage(tex.Data, tex.Width, tex.Height, pal, project.Indexer), project.Offset)
		transparent := -1
		if project.Textures[i].HasTransparency {
			pixel := project.Textures[i].TransparentX + project.Textures[i].TransparentY*tex.Width
			if pixel < len(converted) {
				transparent = int(converted[pixel])
			}
		}
		binary.Write(file, binary.LittleEndian, int16(transparent))
		binary.Write(file, binary.LittleEndian, converted)
	}
}
