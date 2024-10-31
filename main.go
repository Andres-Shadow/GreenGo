package main

import (
	"fmt"
	"time"
)

const (
	RepoLink       = "" // Repository Link
	BranchName     = "" // Branch Name
	CheckInterval  = 5 * time.Second
)

func main() {
	lastCommit := ""

	for {
		branchCommit, err := fetchLatestCommit()
		if err != nil {
			fmt.Println("Error fetching remote commits:", err)
			time.Sleep(CheckInterval)
			continue
		}

		if lastCommit != "" && lastCommit != branchCommit {
			fmt.Printf("Processing new commit: %s\n", branchCommit)
			processPipeline()
		} else {
			fmt.Printf("No new commit found: %s\n", lastCommit)
		}

		lastCommit = branchCommit
		time.Sleep(CheckInterval)
	}
}
