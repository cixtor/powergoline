package main

import "bytes"
import "os/exec"

type ExtBinary struct {
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
