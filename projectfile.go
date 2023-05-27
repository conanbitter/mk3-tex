package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/shlex"
)

type TextureEntry struct {
	Name            string
	Filename        string
	HasTransparency bool
	TransparentX    int
	TransparentY    int
}

type ProjectFile struct {
	Colors   int
	Offset   int
	Textures []TextureEntry
}

func OpenProject(filename string) ProjectFile {
	var result ProjectFile
	result.Textures = make([]TextureEntry, 0)

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	folder, err := filepath.Abs(filepath.Dir(filename))
	if err != nil {
		log.Fatal(err)
	}

	names := make(map[string]struct{})

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		command := false
		if (len(text)) == 0 {
			continue
		}
		if text[0] == '#' {
			command = true
			text = text[1:]
		}
		fields, err := shlex.Split(text)
		if err != nil {
			log.Fatal(err)
		}
		if len(fields) == 0 {
			continue
		}
		if command {
			switch fields[0] {
			case "colors":
				if len(fields) < 2 {
					log.Fatal("Not enough arguments for command 'colors'")
				}
				colors, err := strconv.Atoi(fields[1])
				if err != nil || colors < 1 || colors > 256 {
					log.Fatal("Wrong argument for command 'colors'")
				}
				result.Colors = colors
			case "offset":
				if len(fields) < 2 {
					log.Fatal("Not enough arguments for command 'offset'")
				}
				offset, err := strconv.Atoi(fields[1])
				if err != nil || offset < 0 || offset > 255 {
					log.Fatal("Wrong argument for command 'offset'")
				}
				result.Offset = offset
			}
		} else {
			if len(fields) != 2 && len(fields) != 4 {
				log.Fatal("Wrong number of argument for texture")
			}

			name := fields[0]
			if len(name) > 16 {
				oldname := name
				name = name[:16]
				fmt.Printf("Name \"%s\" will be cropped to \"%s\"\n", oldname, name)
			}
			if _, ok := names[name]; ok {
				log.Fatalf("Name \"%s\" is not unique", name)
			}
			names[name] = struct{}{}

			path := fields[1]
			if !filepath.IsAbs(path) {
				path = filepath.Join(folder, path)
			}

			transparency := false
			tx := 0
			ty := 0
			if len(fields) == 4 {
				transparency = true
				tx, err = strconv.Atoi(fields[2])
				if err != nil {
					log.Fatal("Wrong argument for X coordinate")
				}
				ty, err = strconv.Atoi(fields[3])
				if err != nil {
					log.Fatal("Wrong argument for Y coordinate")
				}
			}
			result.Textures = append(result.Textures, TextureEntry{
				Name:            name,
				Filename:        path,
				HasTransparency: transparency,
				TransparentX:    tx,
				TransparentY:    ty,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	if result.Colors+result.Offset > 256 {
		log.Fatalf("Wrong number of colors (%d+%d>256)", result.Colors, result.Offset)
	}
	return result
}
