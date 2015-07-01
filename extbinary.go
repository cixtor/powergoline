package main

import (
	"bytes"
	"errors"
	"os/exec"
	"regexp"
	"strings"
)

type ExtBinary struct {
}

func (extbin ExtBinary) GitBranch() ([]byte, error) {
	kommand := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	response, err := kommand.CombinedOutput()

	if err != nil {
		return nil, err
	}

	response = bytes.Trim(response, "\n")
	return response, nil
}

func (extbin ExtBinary) GitStatus() (map[string]int, error) {
	kommand := exec.Command("git", "status", "--porcelain", "--ignore-submodules")
	response, err := kommand.CombinedOutput()

	if err != nil {
		return nil, err
	}

	if len(response) == 0 {
		return nil, errors.New("output is empty")
	}

	regex := regexp.MustCompile(`^(A\s|\sM|\sD|\?\?) .+`)
	var output string = string(response)
	var lines []string = strings.Split(output, "\n")
	var modified_files int = 0
	var deleted_files int = 0
	var added_files int = 0
	var stats = make(map[string]int)

	for _, line := range lines {
		if line != "" {
			var match []string = regex.FindStringSubmatch(line)

			if len(match) > 0 {
				switch match[1] {
				case " M":
					modified_files += 1
				case " D":
					deleted_files += 1
				case "A ", "??":
					added_files += 1
				}
			}
		}
	}

	stats["modified"] = modified_files
	stats["deleted"] = deleted_files
	stats["added"] = added_files

	return stats, nil
}

func (extbin ExtBinary) GitStatusExtra() (map[string]bool, error) {
	kommand := exec.Command("git", "status", "--ignore-submodules")
	response, err := kommand.CombinedOutput()

	if err != nil {
		return nil, err
	}

	if len(response) == 0 {
		return nil, errors.New("output is empty")
	}

	var output string = string(response)
	var nothing bool = strings.Contains(output, "nothing to commit")
	var committed bool = strings.Contains(output, "Changes to be committed:")
	var untracked bool = strings.Contains(output, "Untracked files:")

	var stats = map[string]bool{
		"nothing":   nothing,
		"committed": committed,
		"untracked": untracked,
	}

	return stats, nil
}

func (extbin ExtBinary) MercurialBranch() ([]byte, error) {
	kommand := exec.Command("hg", "branch")
	response, err := kommand.CombinedOutput()

	if err != nil {
		return nil, err
	}

	response = bytes.Trim(response, "\n")
	return response, nil
}

func (extbin ExtBinary) MercurialStatus() (map[string]int, error) {
	kommand := exec.Command("hg", "status")
	response, err := kommand.CombinedOutput()

	if err != nil {
		return nil, err
	}

	if len(response) == 0 {
		return nil, errors.New("output is empty")
	}

	regex := regexp.MustCompile(`^(A|M|R|\!|\?) .+`)
	var output string = string(response)
	var lines []string = strings.Split(output, "\n")
	var modified_files int = 0
	var deleted_files int = 0
	var added_files int = 0
	var stats = make(map[string]int)

	for _, line := range lines {
		if line != "" {
			var match []string = regex.FindStringSubmatch(line)

			if len(match) > 0 {
				switch match[1] {
				case "M":
					modified_files += 1
				case "R", "!":
					deleted_files += 1
				case "A", "?":
					added_files += 1
				}
			}
		}
	}

	stats["modified"] = modified_files
	stats["deleted"] = deleted_files
	stats["added"] = added_files

	return stats, nil
}
