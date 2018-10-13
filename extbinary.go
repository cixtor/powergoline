package main

import (
	"bytes"
	"errors"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// ExtBinary encapsulates the methods that spawn the execution of external
// binaries like the control version repository tools among others that
// facilitate the popullation of the command line prompt. Additional methods
// can be attached following the same template as the original interface.
type ExtBinary struct{}

// RepositoryStatus holds the information of the current state of a repository,
// this includes the number of untracked files, number of commits ahead from
// remote, number of commits behind compared to the state of the remote
// repository, and nothing in case the state of the local repository is the
// same as the remote version.
type RepositoryStatus struct {
	Nothing       bool
	Committed     bool
	Untracked     bool
	AheadCommits  int
	BehindCommits int
}

// Run executes an external command and returns the output.
func (extbin ExtBinary) Run(command string) ([]byte, error) {
	response, err := exec.Command(command).CombinedOutput() // #nosec

	if err != nil {
		return nil, err
	}

	return bytes.Trim(response, "\n"), nil
}

// GitBranch returns the name of the current Git branch.
func (extbin ExtBinary) GitBranch() ([]byte, error) {
	response, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").CombinedOutput() // #nosec

	if err != nil {
		return nil, err
	}

	return bytes.Trim(response, "\n"), nil
}

// GitStatus returns information about the current state of a Git repository.
func (extbin ExtBinary) GitStatus() (map[string]int, error) {
	response, err := exec.Command("git", "status", "--porcelain", "--ignore-submodules").CombinedOutput() // #nosec

	if err != nil {
		return nil, err
	}

	if len(response) == 0 {
		return nil, errors.New("output is empty")
	}

	output := string(response)
	stats := make(map[string]int)
	regex := regexp.MustCompile(`^(A\s|\sM|\sD|\?\?) .+`)
	lines := strings.Split(output, "\n")

	var modifiedFiles int
	var deletedFiles int
	var addedFiles int

	for _, line := range lines {
		if line == "" {
			continue
		}

		match := regex.FindStringSubmatch(line)

		if len(match) > 0 {
			switch match[1] {
			case " M":
				modifiedFiles++
			case " D":
				deletedFiles++
			case "A ", "??":
				addedFiles++
			}
		}
	}

	stats["modified"] = modifiedFiles
	stats["deleted"] = deletedFiles
	stats["added"] = addedFiles

	return stats, nil
}

// GitStatusExtra includes additional information to the GitStatus output.
func (extbin ExtBinary) GitStatusExtra() (RepositoryStatus, error) {
	var stats RepositoryStatus

	response, err := exec.Command("git", "status", "--ignore-submodules").CombinedOutput() // #nosec

	if err != nil {
		return stats, err
	}

	if len(response) == 0 {
		return stats, errors.New("output is empty")
	}

	output := string(response)

	stats.Nothing = strings.Contains(output, "nothing to commit")
	stats.Committed = strings.Contains(output, "Changes to be committed:")
	stats.Untracked = strings.Contains(output, "Untracked files:")

	pattern := regexp.MustCompile(`(ahead|behind) of .+ by ([0-9]+) commits`)

	if commits := pattern.FindStringSubmatch(output); commits != nil {
		if number, err := strconv.Atoi(commits[2]); err == nil {
			if commits[1] == "ahead" {
				stats.AheadCommits = number
			} else if commits[1] == "behind" {
				stats.BehindCommits = number
			}
		}
	}

	return stats, nil
}

// MercurialBranch returns the name of the current Mercurial branch.
func (extbin ExtBinary) MercurialBranch() ([]byte, error) {
	response, err := exec.Command("hg", "branch").CombinedOutput() // #nosec

	if err != nil {
		return nil, err
	}

	return bytes.Trim(response, "\n"), nil
}

// MercurialStatus returns information about the current state of a Mercurial repository.
func (extbin ExtBinary) MercurialStatus() (map[string]int, error) {
	response, err := exec.Command("hg", "status").CombinedOutput() // #nosec

	if err != nil {
		return nil, err
	}

	if len(response) == 0 {
		return nil, errors.New("output is empty")
	}

	regex := regexp.MustCompile(`^(A|M|R|\!|\?) .+`)

	stats := make(map[string]int)
	output := string(response)
	lines := strings.Split(output, "\n")

	var modifiedFiles int
	var deletedFiles int
	var addedFiles int

	for _, line := range lines {
		if line == "" {
			continue
		}

		match := regex.FindStringSubmatch(line)

		if len(match) > 0 {
			switch match[1] {
			case "M":
				modifiedFiles++
			case "R", "!":
				deletedFiles++
			case "A", "?":
				addedFiles++
			}
		}
	}

	stats["modified"] = modifiedFiles
	stats["deleted"] = deletedFiles
	stats["added"] = addedFiles

	return stats, nil
}
