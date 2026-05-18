package greengo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Logger is the minimal logging contract used by GreenGo.
type Logger interface {
	Printf(format string, v ...any)
}

// LoggerFunc adapts a function into a Logger.
type LoggerFunc func(format string, v ...any)

// Printf implements Logger.
func (f LoggerFunc) Printf(format string, v ...any) {
	f(format, v...)
}

// Command describes a process executed by a pipeline stage.
type Command struct {
	Name string
	Args []string
	Dir  string
	Env  []string
}

// String returns a shell-like representation useful for logs.
func (c Command) String() string {
	parts := append([]string{c.Name}, c.Args...)
	return strings.Join(parts, " ")
}

// Result contains the captured output from a command.
type Result struct {
	Stdout string
	Stderr string
}

// Runner executes commands for command-backed stages.
type Runner interface {
	Run(ctx context.Context, command Command) (Result, error)
}

// ExecRunner executes commands with os/exec.
type ExecRunner struct{}

// Run executes command and captures stdout and stderr.
func (ExecRunner) Run(ctx context.Context, command Command) (Result, error) {
	if command.Name == "" {
		return Result{}, errors.New("command name is required")
	}

	cmd := exec.CommandContext(ctx, command.Name, command.Args...)
	cmd.Dir = command.Dir
	if len(command.Env) > 0 {
		cmd.Env = append(os.Environ(), command.Env...)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := Result{Stdout: stdout.String(), Stderr: stderr.String()}
	if err != nil {
		return result, fmt.Errorf("%s failed: %w: %s", command.String(), err, strings.TrimSpace(result.Stderr))
	}
	return result, nil
}

// RunContext carries execution dependencies and metadata through stages.
type RunContext struct {
	Workspace string
	Commit    string
	Runner    Runner
	Logger    Logger
}

// StageFunc is custom logic executed as a stage.
type StageFunc func(ctx context.Context, run RunContext) error

// Stage is one ordered unit of work in a pipeline.
type Stage struct {
	Name    string
	Command Command
	Run     StageFunc
}

// Pipeline is a reusable sequence of stages.
type Pipeline struct {
	name   string
	stages []Stage
	runner Runner
	logger Logger
}

// Option configures a Pipeline.
type Option func(*Pipeline)

// WithRunner sets the command runner used by command stages.
func WithRunner(runner Runner) Option {
	return func(p *Pipeline) {
		if runner != nil {
			p.runner = runner
		}
	}
}

// WithLogger sets the logger used by pipeline runs.
func WithLogger(logger Logger) Option {
	return func(p *Pipeline) {
		if logger != nil {
			p.logger = logger
		}
	}
}

// NewPipeline creates an empty pipeline.
func NewPipeline(name string, opts ...Option) *Pipeline {
	p := &Pipeline{
		name:   name,
		runner: ExecRunner{},
		logger: log.New(io.Discard, "", 0),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Name returns the pipeline name.
func (p *Pipeline) Name() string {
	return p.name
}

// Stages returns a copy of the configured stages.
func (p *Pipeline) Stages() []Stage {
	stages := make([]Stage, len(p.stages))
	copy(stages, p.stages)
	return stages
}

// AddStage appends a custom stage.
func (p *Pipeline) AddStage(name string, run StageFunc) *Pipeline {
	p.stages = append(p.stages, Stage{Name: name, Run: run})
	return p
}

// AddCommand appends a command-backed stage.
func (p *Pipeline) AddCommand(name string, command Command) *Pipeline {
	p.stages = append(p.stages, Stage{Name: name, Command: command})
	return p
}

// Run executes the pipeline stages in order and stops on the first error.
func (p *Pipeline) Run(ctx context.Context, run RunContext) error {
	if p.runner == nil {
		p.runner = ExecRunner{}
	}
	if p.logger == nil {
		p.logger = log.New(io.Discard, "", 0)
	}
	if run.Runner == nil {
		run.Runner = p.runner
	}
	if run.Logger == nil {
		run.Logger = p.logger
	}

	for i, stage := range p.stages {
		stageName := stage.Name
		if stageName == "" {
			stageName = fmt.Sprintf("stage-%d", i+1)
		}

		run.Logger.Printf("greengo: starting %s", stageName)
		if stage.Run != nil {
			if err := stage.Run(ctx, run); err != nil {
				return fmt.Errorf("stage %q failed: %w", stageName, err)
			}
			continue
		}

		if stage.Command.Name == "" {
			return fmt.Errorf("stage %q has no command or custom runner", stageName)
		}
		if _, err := run.Runner.Run(ctx, stage.Command); err != nil {
			return fmt.Errorf("stage %q failed: %w", stageName, err)
		}
	}
	return nil
}

// EnsureCleanDir removes and recreates dir. It refuses risky paths.
func EnsureCleanDir(dir string) error {
	if strings.TrimSpace(dir) == "" {
		return errors.New("directory is required")
	}
	if strings.TrimSpace(dir) == "." {
		return errors.New("refusing to clean current directory")
	}

	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	root := filepath.VolumeName(abs) + string(filepath.Separator)
	if abs == root || abs == cwd {
		return fmt.Errorf("refusing to clean unsafe directory %q", abs)
	}

	if err := os.RemoveAll(abs); err != nil {
		return err
	}
	return os.MkdirAll(abs, 0o755)
}
