package main

func main() {
	// Oopsy error isn't checked
	f1()
}

func f1() error {
	return nil
}
