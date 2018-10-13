package main

import (
	"bytes"
	"testing"
)

func compareRepoStatus(t *testing.T, actual RepoStatus, expected RepoStatus) {
	if !bytes.Equal(actual.Branch, expected.Branch) {
		t.Fatalf("unexpected status.Branch `%s` != `%s`", actual.Branch, expected.Branch)
	}

	if actual.Ahead != expected.Ahead {
		t.Fatalf("unexpected status.Ahead %d != %d", actual.Ahead, expected.Ahead)
	}

	if actual.Behind != expected.Behind {
		t.Fatalf("unexpected status.Behind %d != %d", actual.Behind, expected.Behind)
	}

	if actual.Added != expected.Added {
		t.Fatalf("unexpected status.Added %d != %d", actual.Added, expected.Added)
	}

	if actual.Deleted != expected.Deleted {
		t.Fatalf("unexpected status.Deleted %d != %d", actual.Deleted, expected.Deleted)
	}

	if actual.Modified != expected.Modified {
		t.Fatalf("unexpected status.Modified %d != %d", actual.Modified, expected.Modified)
	}
}

func TestStatusGit(t *testing.T) {
	lines := [][]byte{
		[]byte("## master...origin/master"),
		[]byte("D  deleted.txt"),
		[]byte(" D missing.txt"),
		[]byte("M  patches.go"),
		[]byte(" M changes.go"),
		[]byte("A  newfile.sh"),
		[]byte("?? isadded.json"),
	}

	status, err := repoStatusGitParse(lines)

	if err != nil {
		t.Fatalf("repoStatusGitParse %s", err)
	}

	compareRepoStatus(t, status, RepoStatus{
		Branch:   []byte("master"),
		Modified: 2,
		Deleted:  2,
		Added:    2,
	})
}

func TestStatusGitAhead(t *testing.T) {
	lines := [][]byte{
		[]byte("## master...origin/master [ahead 5]"),
		[]byte("D  deleted.txt"),
		[]byte(" D missing.txt"),
		[]byte("M  patches.go"),
		[]byte(" M changes.go"),
		[]byte("A  newfile.sh"),
		[]byte("?? isadded.json"),
	}

	status, err := repoStatusGitParse(lines)

	if err != nil {
		t.Fatalf("repoStatusGitParse %s", err)
	}

	compareRepoStatus(t, status, RepoStatus{
		Branch:   []byte("master"),
		Modified: 2,
		Deleted:  2,
		Added:    2,
		Ahead:    5,
	})
}

func TestStatusGitBehind(t *testing.T) {
	lines := [][]byte{
		[]byte("## master...origin/master [behind 8]"),
		[]byte("D  deleted.txt"),
		[]byte(" D missing.txt"),
		[]byte("M  patches.go"),
		[]byte(" M changes.go"),
		[]byte("A  newfile.sh"),
		[]byte("?? isadded.json"),
	}

	status, err := repoStatusGitParse(lines)

	if err != nil {
		t.Fatalf("repoStatusGitParse %s", err)
	}

	compareRepoStatus(t, status, RepoStatus{
		Branch:   []byte("master"),
		Modified: 2,
		Deleted:  2,
		Added:    2,
		Behind:   8,
	})
}

func TestStatusGitAheadBehind(t *testing.T) {
	lines := [][]byte{
		[]byte("## master...origin/master [ahead 5, behind 8]"),
		[]byte("D  deleted.txt"),
		[]byte(" D missing.txt"),
		[]byte("M  patches.go"),
		[]byte(" M changes.go"),
		[]byte("A  newfile.sh"),
		[]byte("?? isadded.json"),
	}

	status, err := repoStatusGitParse(lines)

	if err != nil {
		t.Fatalf("repoStatusGitParse %s", err)
	}

	compareRepoStatus(t, status, RepoStatus{
		Branch:   []byte("master"),
		Modified: 2,
		Deleted:  2,
		Added:    2,
		Ahead:    5,
		Behind:   8,
	})
}

func TestStatusMercurial(t *testing.T) {
	lines := [][]byte{
		[]byte("R deleted.txt"),
		[]byte("! missing.txt"),
		[]byte("M patches.go"),
		[]byte("M changes.go"),
		[]byte("A newfile.sh"),
		[]byte("? isadded.json"),
	}

	status, err := repoStatusMercurialParse(lines)

	if err != nil {
		t.Fatalf("repoStatusMercurialParse %s", err)
	}

	compareRepoStatus(t, status, RepoStatus{
		Branch:   []byte("default"),
		Modified: 2,
		Deleted:  2,
		Added:    2,
	})
}
