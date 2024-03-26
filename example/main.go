package main

import "fmt"

//go:generate file_line -src=./

func main() {
	fmt.Println("use a place holder:", "[file.go:123]")
}
