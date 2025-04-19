package main

import "fmt"

func main() {
	f1("f1")
	f2("")
	f3()
}

func f1(s string) {
	if s == "" {
		panic("empty string")
	}
}

func f2(s string) error {
	if s == "" {
		return fmt.Errorf("empty string")
	}
	return nil
}

func f3() {
	fmt.Println("f3")
}
