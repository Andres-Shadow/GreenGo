package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// execCmd runs a shell command in a specified directory and returns its output
func execCmd(cmd []string, workingDir string, returnJSON bool) ([]byte, error) {
	// Create the command
	command := exec.Command(cmd[0], cmd[1:]...)
	command.Dir = workingDir

	// Capture standard output and error
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()
	if err != nil {
		fmt.Printf("\nERROR EXECUTING TASK: %s\nSTDOUT: %s\nSTDERR: %s\n\n", cmd, stdout.String(), stderr.String())
		return nil, err
	}

	if returnJSON {
		return json.Marshal(stdout.String())
	}
	return stdout.Bytes(), nil
}

// pipelineStage executes a specific stage in the pipeline and logs any errors
func pipelineStage(cmd []string, workingDir string) error {
	fmt.Printf("Executing stage: %s\n", strings.Join(cmd, " "))
	_, err := execCmd(cmd, workingDir, false)
	if err != nil {
		fmt.Printf("Error in stage %s: %v\n", cmd[0], err)
	}
	return err
}

// processPipeline performs each stage of the pipeline for deploying the application
func processPipeline() {
	// Step 1: Remove the previous directory if it exists
	if err := pipelineStage([]string{"rm", "-R", "-f", "files"}, "."); err != nil {
		return // Stop if this step fails
	}

	// Step 2: Create a new directory for cloning the repo
	if err := pipelineStage([]string{"mkdir", "files"}, "."); err != nil {
		return // Stop if this step fails
	}

	// Step 3: Clone the repository into the new directory
	if err := pipelineStage([]string{"git", "clone", RepoLink, "files"}, "."); err != nil {
		return // Stop if this step fails
	}

	// Step 4: Launch Docker containers with docker-compose
	if err := pipelineStage([]string{"docker-compose", "up", "--build", "-d"}, "files"); err != nil {
		return // Stop if this step fails
	}

	// Additional steps can be added here using pipelineStage
}

// fetchLatestCommit retrieves the latest commit from the specified branch in the remote repository
func fetchLatestCommit() (string, error) {
	// Get the list of remote commits
	r, err := execCmd([]string{"git", "ls-remote", "--heads", RepoLink}, ".", false)
	if err != nil {
		return "", err
	}

	// Look for the commit in the specified branch
	commits := strings.Split(string(r), "\n")
	for _, line := range commits {
		if strings.HasSuffix(line, "refs/heads/"+BranchName) {
			return strings.Fields(line)[0], nil
		}
	}
	return "", fmt.Errorf("no branch commit found")
}
