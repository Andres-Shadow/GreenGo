package greengo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	defaultBranch         = "main"
	defaultWorkspace      = "greengo-workspace"
	defaultComposeCommand = "docker"
)

// DeployConfig describes the default Git to Docker Compose deployment flow.
type DeployConfig struct {
	RepoURL        string
	Branch         string
	Workspace      string
	ComposeFiles   []string
	ComposeCommand string
	Build          bool
	Detach         bool
}

// Normalize fills safe defaults for optional deployment settings.
func (c DeployConfig) Normalize() DeployConfig {
	if c.Branch == "" {
		c.Branch = defaultBranch
	}
	if c.Workspace == "" {
		c.Workspace = defaultWorkspace
	}
	if c.ComposeCommand == "" {
		c.ComposeCommand = defaultComposeCommand
	}
	return c
}

// Validate checks the required deployment settings.
func (c DeployConfig) Validate() error {
	if c.RepoURL == "" {
		return errors.New("repository URL is required")
	}
	return nil
}

// NewDockerComposePipeline returns a deployment pipeline that clones the target
// repository and runs Docker Compose from the cloned project.
func NewDockerComposePipeline(config DeployConfig, opts ...Option) (*Pipeline, error) {
	config = config.Normalize()
	if err := config.Validate(); err != nil {
		return nil, err
	}

	pipeline := NewPipeline("docker-compose-deploy", opts...)
	pipeline.AddStage("prepare workspace", func(ctx context.Context, run RunContext) error {
		workspace := run.Workspace
		if workspace == "" {
			workspace = config.Workspace
		}
		return EnsureCleanDir(workspace)
	})

	pipeline.AddStage("clone repository", func(ctx context.Context, run RunContext) error {
		workspace := run.Workspace
		if workspace == "" {
			workspace = config.Workspace
		}
		parent := filepath.Dir(workspace)
		if err := os.MkdirAll(parent, 0o755); err != nil {
			return err
		}
		_, err := run.Runner.Run(ctx, Command{
			Name: "git",
			Args: []string{"clone", "--branch", config.Branch, "--single-branch", config.RepoURL, workspace},
		})
		return err
	})

	pipeline.AddStage("docker compose up", func(ctx context.Context, run RunContext) error {
		workspace := run.Workspace
		if workspace == "" {
			workspace = config.Workspace
		}

		args := composeBaseArgs(config.ComposeCommand)
		for _, file := range config.ComposeFiles {
			args = append(args, "-f", file)
		}
		args = append(args, "up")
		if config.Build {
			args = append(args, "--build")
		}
		if config.Detach {
			args = append(args, "-d")
		}

		_, err := run.Runner.Run(ctx, Command{
			Name: config.ComposeCommand,
			Args: args,
			Dir:  workspace,
		})
		if err != nil && config.ComposeCommand == defaultComposeCommand {
			return fmt.Errorf("%w; install Docker Compose v2 or set ComposeCommand to docker-compose", err)
		}
		return err
	})

	return pipeline, nil
}

func composeBaseArgs(command string) []string {
	if command == "docker-compose" {
		return nil
	}
	return []string{"compose"}
}
