package greengo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// LatestCommit returns the commit hash for branch in a remote Git repository.
func LatestCommit(ctx context.Context, runner Runner, repoURL, branch string) (string, error) {
	if runner == nil {
		runner = ExecRunner{}
	}
	if repoURL == "" {
		return "", errors.New("repository URL is required")
	}
	if branch == "" {
		branch = defaultBranch
	}

	result, err := runner.Run(ctx, Command{
		Name: "git",
		Args: []string{"ls-remote", "--heads", repoURL, branch},
	})
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(result.Stdout, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == "refs/heads/"+branch {
			return fields[0], nil
		}
	}
	return "", fmt.Errorf("branch %q was not found in %s", branch, repoURL)
}

// WatchConfig configures commit polling for a repository branch.
type WatchConfig struct {
	RepoURL    string
	Branch     string
	Interval   time.Duration
	Workspace  string
	Pipeline   *Pipeline
	Runner     Runner
	Logger     Logger
	InitialRun bool
}

// Watch polls a remote branch and runs the pipeline whenever a new commit is found.
func Watch(ctx context.Context, config WatchConfig) error {
	if config.Pipeline == nil {
		return errors.New("pipeline is required")
	}
	if config.Runner == nil {
		config.Runner = ExecRunner{}
	}
	if config.Interval <= 0 {
		config.Interval = 30 * time.Second
	}
	if config.Branch == "" {
		config.Branch = defaultBranch
	}

	var lastCommit string
	for {
		commit, err := LatestCommit(ctx, config.Runner, config.RepoURL, config.Branch)
		if err != nil {
			if config.Logger != nil {
				config.Logger.Printf("greengo: commit check failed: %v", err)
			}
		} else if lastCommit == "" {
			lastCommit = commit
			if config.InitialRun {
				if err := runWatchedPipeline(ctx, config, commit); err != nil {
					return err
				}
			}
		} else if commit != lastCommit {
			lastCommit = commit
			if err := runWatchedPipeline(ctx, config, commit); err != nil {
				return err
			}
		} else if config.Logger != nil {
			config.Logger.Printf("greengo: no new commit on %s (%s)", config.Branch, commit)
		}

		timer := time.NewTimer(config.Interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func runWatchedPipeline(ctx context.Context, config WatchConfig, commit string) error {
	if config.Logger != nil {
		config.Logger.Printf("greengo: running pipeline for commit %s", commit)
	}
	return config.Pipeline.Run(ctx, RunContext{
		Workspace: config.Workspace,
		Commit:    commit,
		Runner:    config.Runner,
		Logger:    config.Logger,
	})
}
