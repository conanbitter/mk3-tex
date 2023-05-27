package main

import (
	"encoding/json"
	"math"
	"os"
	"sort"
)

type FloatColor struct {
	R, G, B float64
}

type IntColor struct {
	R, G, B int
}

type Palette []IntColor

//===== FLOAT COLOR =======

func clipFloat(val float64) float64 {
	if val > 1.0 {
		return 1.0
	}
	if val < 0.0 {
		return 0.0
	}
	return val
}

func (color FloatColor) Normalized() FloatColor {
	return FloatColor{
		clipFloat(color.R),
		clipFloat(color.G),
		clipFloat(color.B)}
}

func (color FloatColor) Distance(other FloatColor) float64 {
	return (color.R-other.R)*(color.R-other.R) +
		(color.G-other.G)*(color.G-other.G) +
		(color.B-other.B)*(color.B-other.B)
}

func (color FloatColor) ToIntColor() IntColor {
	norm := color.Normalized()
	return IntColor{int(norm.R * 255), int(norm.G * 255), int(norm.B * 255)}
}

//===== INT COLOR =======

func clipInt(val int) int {
	if val > 255 {
		return 255
	}
	if val < 0 {
		return 0
	}
	return val
}

func (color IntColor) Normalized() IntColor {
	return IntColor{
		clipInt(color.R),
		clipInt(color.G),
		clipInt(color.B)}
}

func (color IntColor) Distance(other IntColor) uint64 {
	return (uint64(color.R)-uint64(other.R))*(uint64(color.R)-uint64(other.R)) +
		(uint64(color.G)-uint64(other.G))*(uint64(color.G)-uint64(other.G)) +
		(uint64(color.B)-uint64(other.B))*(uint64(color.B)-uint64(other.B))
}

func (color IntColor) Luma() float32 {
	return 0.2126*float32(color.R) + 0.7152*float32(color.G) + 0.0722*float32(color.B)
}

func (color IntColor) ToFloatColor() FloatColor {
	return FloatColor{float64(color.R) / 255.0, float64(color.G) / 255.0, float64(color.B) / 255.0}
}

//===== PALETTE =======

func NewPalette(colors int) Palette {
	return make([]IntColor, colors)
}

func (pal Palette) Len() int {
	return len(pal)
}

func (pal Palette) Less(i, j int) bool {
	return pal[i].Luma() < pal[j].Luma()
}

func (pal Palette) Swap(i, j int) {
	pal[i], pal[j] = pal[j], pal[i]
}

func (pal Palette) Sort() {
	sort.Sort(pal)
}

func (pal *Palette) Save(filename string) {
	pal.Sort()
	fo, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer fo.Close()
	data, err := json.MarshalIndent(pal, "", "    ")
	if err != nil {
		panic(err)
	}
	_, err = fo.Write(data)
	if err != nil {
		panic(err)
	}
}

func (pal Palette) GetIntColorIndex(color IntColor) (index int) {
	var mindist uint64 = math.MaxUint64
	index = 0
	for i, c := range pal {
		dist := color.Distance(c)
		if dist < mindist {
			mindist = dist
			index = i
		}
	}
	return
}

func (pal Palette) GetFloatColorIndex(color FloatColor) (index int) {
	return pal.GetIntColorIndex(color.ToIntColor())
}

func PaletteLoad(filename string) Palette {
	fi, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	result := &Palette{}
	err = json.Unmarshal(fi, result)
	if err != nil {
		panic(err)
	}
	return *result
}
