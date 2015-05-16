package main

import "fmt"
import "os"
import "strings"

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
	var homedir string = os.Getenv("HOME")
	var workingdir string = os.Getenv("PWD")
	var shortdir string = strings.Replace(workingdir, homedir, "", -1)
	var cleandir string = strings.Trim(shortdir, "/")

	// Print the user home directory path.
	fmt.Printf("\033[38;5;255;48;5;161m ~ \033[0m")

	if cleandir == "" {
		fmt.Printf("\033[38;5;161;48;5;070m\ue0b0\033[0m")
	} else {
		fmt.Printf("\033[38;5;161;48;5;238m\ue0b0\033[0m")
	}

	// Print the sequence of folders of the current path.
	var segments []string = strings.Split(cleandir, "/")
	var ttlsegms int = len(segments)
	var last_segm int = (ttlsegms - 1)

	for key, folder := range segments {
		if folder != "" {
			fmt.Printf("\033[38;5;251;48;5;238m %s \033[0m", folder)

			if key == last_segm {
				fmt.Printf("\033[38;5;238;48;5;070m\ue0b0\033[0m")
			} else {
				fmt.Printf("\033[38;5;251;48;5;238m\ue0b1\033[0m")
			}
		}
	}
}

func (pogol PowerGoLine) RootSymbol() {
	fmt.Printf("\033[38;5;255;48;5;070m $ \033[0m")
	fmt.Printf("\033[38;5;000;38;5;070m\ue0b0\033[0m")
	fmt.Printf("\n")
}
