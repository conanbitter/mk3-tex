package main

import "fmt"

func main() {
	project := OpenProject("test_assets/project.txt")
	fmt.Println(project)
}
