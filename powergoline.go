package main

import "fmt"
import "os"

type PowerGoLine struct {
}

func (pogol PowerGoLine) Username() {
	var username string = os.Getenv("USERNAME")

	fmt.Printf("\033[38;5;255;48;5;006m %s \033[0m", username)
	fmt.Printf("\033[38;5;006;48;5;012m\ue0b0\033[0m")
}

func (pogol PowerGoLine) Hostname() {
	hostname, _ := os.Hostname()

	fmt.Printf("\033[38;5;255;48;5;012m %s \033[0m", hostname)
	fmt.Printf("\033[38;5;012;48;5;161m\ue0b0\033[0m")
}

func (pogol PowerGoLine) WorkingDirectory() {
	var workingdir string = os.Getenv("PWD")

	fmt.Printf("\033[38;5;255;48;5;161m %s \033[0m", workingdir)
	fmt.Printf("\033[38;5;161;48;5;238m\ue0b0\033[0m")
}

func (pogol PowerGoLine) RootSymbol() {
	fmt.Printf("\033[38;5;255;48;5;070m $ \033[0m")
	fmt.Printf("\033[38;5;000;38;5;070m\ue0b0\033[0m")
	fmt.Printf("\n")
}
