package main

import "fmt"
import "os"

type PowerGoLine struct {
}

func (pogol PowerGoLine) Username() {
	var username string = os.Getenv("USERNAME")

	fmt.Printf("%s", username)
}

func (pogol PowerGoLine) Hostname() {
	hostname, _ := os.Hostname()

	fmt.Printf("@%s", hostname)
}

func (pogol PowerGoLine) WorkingDirectory() {
	var workingdir string = os.Getenv("PWD")

	fmt.Printf(" > %s", workingdir)
}

func (pogol PowerGoLine) RootSymbol() {
	fmt.Printf(" $ \n")
}
