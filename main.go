package main

import (
	"bytes"
	"commitgen/internal"
	"flag"
	"log"
	"log/slog"

	"github.com/go-git/go-git"
)

func main() {
	logger := internal.InitLogger()
	defer logger.Sync()

	projectPath := flag.String("path", ".", "Path to project")
	flag.Parse()

	if projectPath == nil {
		log.Fatalf("Project path is empty")
	}

	repo, err := git.PlainOpen(*projectPath)
	if err != nil {
		slog.Error("failed  to open repository", "cause", err)
		return
	}

	head, err := repo.Head()
	if err != nil {
		slog.Error("failed to get HEAD commit", "cause", err)
		return
	}

	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		slog.Error("failed to get commit object", "cause", err)
		return
	}

	if commit.NumParents() == 0 {
		log.Fatalf("Commit has no parents")
	}

	parentCommit, err := commit.Parent(0)
	if err != nil {
		slog.Error("failed to get parent commit", "cause", err)
		return
	}

	patch, err := parentCommit.Patch(commit)
	if err != nil {
		slog.Error("failed to get patch", "cause", err)
		return
	}

	var diffBuffer bytes.Buffer

	_, err = diffBuffer.WriteString(patch.String())
	if err != nil {
		slog.Error("failed to write diff to buffer", "cause", err)
		return
	}

	log.Println("Git diff given to buffer:")
	log.Println(diffBuffer.String())
}
