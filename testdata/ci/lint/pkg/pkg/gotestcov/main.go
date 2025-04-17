package main

import "fmt"

func main() {
	f1("f1")
	f2()
	f3()
}

func f1(s string) {
	if s == "" {
		panic("empty string")
	}
}

func f2() {
	fmt.Println("f2")
}

func f3() {
	fmt.Println("f3")
}
