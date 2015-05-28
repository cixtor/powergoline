package main

import "fmt"
import "os"
import "strings"

type PowerGoLine struct {
}

func (pogol PowerGoLine) ExitColor(pcolor PowerColor, status string) string {
	var extcolor string

	/**
	 * System Status Codes.
	 *
	 * 0     - Operation success and generic status code.
	 * 1     - Catchall for general errors and failures.
	 * 2     - Misuse of shell builtins, missing command, or permission problem.
	 * 126   - Command invoked cannot execute, permission problem,
	 *         or the command is not an executable binary.
	 * 127   - Command not found, illegal path, or possible typo.
	 * 128   - Invalid argument to exit, only use range 0-255.
	 * 128+n - Fatal error signal where "n" is the PID.
	 * 130   - Script terminated by Control-C.
	 * 255*  - Exit status out of range.
	 */

	if status == "0" {
		extcolor = pcolor.Status.Success
	} else if status == "1" {
		extcolor = pcolor.Status.Failure
	} else if status == "126" {
		extcolor = pcolor.Status.Permission
	} else if status == "127" {
		extcolor = pcolor.Status.NotFound
	} else if status == "128" {
		extcolor = pcolor.Status.InvalidExit
	} else if status == "130" {
		extcolor = pcolor.Status.Terminated
	} else {
		extcolor = pcolor.Status.Misuse
	}

	return extcolor
}

func (pogol PowerGoLine) Username(pcolor PowerColor) {
	var username string = os.Getenv("USERNAME")
	var fg string = pcolor.UsernameFg
	var bg string = pcolor.UsernameBg
	var hbg string = pcolor.HostnameBg

	fmt.Printf("\\[\033[38;5;%s;48;5;%sm\\] %s \\[\033[0m\\]", fg, bg, username)
	fmt.Printf("\\[\033[38;5;%s;48;5;%sm\\]\uE0B0\\[\033[0m\\]", bg, hbg)
}

func (pogol PowerGoLine) Hostname() {
	hostname, err := os.Hostname()

	if err != nil {
		hostname = "unknown"
	}

	fmt.Printf("\\[\033[38;5;255;48;5;012m\\] %s \\[\033[0m\\]", hostname)
	fmt.Printf("\\[\033[38;5;012;48;5;161m\\]\uE0B0\\[\033[0m\\]")
}

func (pogol PowerGoLine) WorkingDirectory(pcolor PowerColor, status string) {
	var homedir string = os.Getenv("HOME")
	var workingdir string = os.Getenv("PWD")
	var shortdir string = strings.Replace(workingdir, homedir, "", -1)
	var cleandir string = strings.Trim(shortdir, "/")
	var extcolor string = pogol.ExitColor(pcolor, status)

	// Print the user home directory path.
	fmt.Printf("\\[\033[38;5;255;48;5;161m\\] ~ \\[\033[0m\\]")

	if cleandir == "" {
		fmt.Printf("\\[\033[38;5;161;48;5;%sm\\]\uE0B0\\[\033[0m\\]", extcolor)
	} else {
		fmt.Printf("\\[\033[38;5;161;48;5;238m\\]\uE0B0\\[\033[0m\\]")
	}

	// Print the sequence of folders of the current path.
	var maxsegms int = 4
	var segments []string = strings.Split(cleandir, "/")
	var ttlsegms int = len(segments)
	var lastsegm int = (ttlsegms - 1)

	if ttlsegms > maxsegms {
		var newsegms []string = make([]string, 0)

		newsegms = append(newsegms, segments[0])
		newsegms = append(newsegms, "\u2026")
		newsegms = append(newsegms, segments[lastsegm-2])
		newsegms = append(newsegms, segments[lastsegm-1])
		newsegms = append(newsegms, segments[lastsegm])

		segments = newsegms
		lastsegm = maxsegms
	}

	for key, folder := range segments {
		if folder != "" {
			fmt.Printf("\\[\033[38;5;251;48;5;238m\\] %s \\[\033[0m\\]", folder)

			if key == lastsegm {
				fmt.Printf("\\[\033[38;5;238;48;5;%sm\\]\uE0B0\\[\033[0m\\]", extcolor)
			} else {
				fmt.Printf("\\[\033[38;5;251;48;5;238m\\]\uE0B1\\[\033[0m\\]")
			}
		}
	}
}

func (pogol PowerGoLine) RootSymbol(pcolor PowerColor, status string) {
	var extcolor string = pogol.ExitColor(pcolor, status)

	fmt.Printf("\\[\033[38;5;255;48;5;%sm\\] $ \\[\033[0m\\]", extcolor)
	fmt.Printf("\\[\033[38;5;%sm\\]\uE0B0\\[\033[0m\\]", extcolor)
	fmt.Printf("\u0020\n")
}
