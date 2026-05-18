package greengo

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type fakeRunner struct {
	commands []Command
	results  map[string]Result
	err      error
}

func (r *fakeRunner) Run(ctx context.Context, command Command) (Result, error) {
	r.commands = append(r.commands, command)
	if r.err != nil {
		return Result{}, r.err
	}
	return r.results[command.Name], nil
}

func TestPipelineRunsStagesInOrder(t *testing.T) {
	var order []string

	pipeline := NewPipeline("test")
	pipeline.AddStage("first", func(ctx context.Context, run RunContext) error {
		order = append(order, "first")
		return nil
	})
	pipeline.AddStage("second", func(ctx context.Context, run RunContext) error {
		order = append(order, "second")
		return nil
	})

	err := pipeline.Run(context.Background(), RunContext{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := []string{"first", "second"}
	if !reflect.DeepEqual(order, expected) {
		t.Fatalf("expected %v, got %v", expected, order)
	}
}

func TestPipelineStopsOnFirstError(t *testing.T) {
	expectedErr := errors.New("boom")
	var order []string

	pipeline := NewPipeline("test")
	pipeline.AddStage("first", func(ctx context.Context, run RunContext) error {
		order = append(order, "first")
		return expectedErr
	})
	pipeline.AddStage("second", func(ctx context.Context, run RunContext) error {
		order = append(order, "second")
		return nil
	})

	err := pipeline.Run(context.Background(), RunContext{})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected wrapped error %v, got %v", expectedErr, err)
	}

	expected := []string{"first"}
	if !reflect.DeepEqual(order, expected) {
		t.Fatalf("expected %v, got %v", expected, order)
	}
}

func TestLatestCommitParsesBranchHash(t *testing.T) {
	runner := &fakeRunner{
		results: map[string]Result{
			"git": {Stdout: "abc123\trefs/heads/main\n"},
		},
	}

	commit, err := LatestCommit(context.Background(), runner, "https://example.com/repo.git", "main")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if commit != "abc123" {
		t.Fatalf("expected abc123, got %s", commit)
	}
}

func TestEnsureCleanDirRefusesCurrentDirectory(t *testing.T) {
	err := EnsureCleanDir(".")
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestDockerComposePipelineCommands(t *testing.T) {
	runner := &fakeRunner{results: map[string]Result{"git": {}, "docker": {}}}
	workspace := t.TempDir()
	pipeline, err := NewDockerComposePipeline(DeployConfig{
		RepoURL:        "https://example.com/repo.git",
		Branch:         "main",
		Workspace:      workspace,
		ComposeFiles:   []string{"docker-compose.yml", "docker-compose.prod.yml"},
		ComposeCommand: "docker",
		Build:          true,
		Detach:         true,
	}, WithRunner(runner))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = pipeline.Run(context.Background(), RunContext{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(runner.commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(runner.commands))
	}
	if runner.commands[0].String() != "git clone --branch main --single-branch https://example.com/repo.git "+workspace {
		t.Fatalf("unexpected clone command: %s", runner.commands[0].String())
	}

	expectedCompose := []string{"compose", "-f", "docker-compose.yml", "-f", "docker-compose.prod.yml", "up", "--build", "-d"}
	if !reflect.DeepEqual(runner.commands[1].Args, expectedCompose) {
		t.Fatalf("expected compose args %v, got %v", expectedCompose, runner.commands[1].Args)
	}
}

func TestDockerComposePipelineSupportsLegacyComposeCommand(t *testing.T) {
	runner := &fakeRunner{results: map[string]Result{"git": {}, "docker-compose": {}}}
	pipeline, err := NewDockerComposePipeline(DeployConfig{
		RepoURL:        "https://example.com/repo.git",
		Workspace:      t.TempDir(),
		ComposeCommand: "docker-compose",
		Build:          true,
		Detach:         true,
	}, WithRunner(runner))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = pipeline.Run(context.Background(), RunContext{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedCompose := []string{"up", "--build", "-d"}
	if !reflect.DeepEqual(runner.commands[1].Args, expectedCompose) {
		t.Fatalf("expected compose args %v, got %v", expectedCompose, runner.commands[1].Args)
	}
}
