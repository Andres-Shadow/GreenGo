package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	greengo "github.com/Andres-Shadow/GreenGo"
)

func main() {
	repo := flag.String("repo", os.Getenv("GREENGO_REPO"), "Git repository URL to monitor")
	branch := flag.String("branch", getenv("GREENGO_BRANCH", "main"), "Git branch to monitor")
	workspace := flag.String("workspace", getenv("GREENGO_WORKSPACE", "greengo-workspace"), "Directory used for cloned sources")
	interval := flag.Duration("interval", getenvDuration("GREENGO_INTERVAL", 30*time.Second), "Polling interval")
	composeFiles := flag.String("compose-file", os.Getenv("GREENGO_COMPOSE_FILES"), "Comma-separated docker compose files")
	composeCommand := flag.String("compose-command", getenv("GREENGO_COMPOSE_COMMAND", "docker"), "Compose command: docker or docker-compose")
	initialRun := flag.Bool("initial-run", getenvBool("GREENGO_INITIAL_RUN", false), "Run the pipeline once for the current commit")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.LstdFlags)
	pipeline, err := greengo.NewDockerComposePipeline(greengo.DeployConfig{
		RepoURL:        *repo,
		Branch:         *branch,
		Workspace:      *workspace,
		ComposeFiles:   splitCSV(*composeFiles),
		ComposeCommand: *composeCommand,
		Build:          true,
		Detach:         true,
	}, greengo.WithLogger(logger))
	if err != nil {
		logger.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	err = greengo.Watch(ctx, greengo.WatchConfig{
		RepoURL:    *repo,
		Branch:     *branch,
		Interval:   *interval,
		Workspace:  *workspace,
		Pipeline:   pipeline,
		Logger:     logger,
		InitialRun: *initialRun,
	})
	if err != nil && err != context.Canceled {
		logger.Fatal(err)
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getenvDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return duration
}

func getenvBool(key string, fallback bool) bool {
	value := strings.ToLower(os.Getenv(key))
	switch value {
	case "1", "true", "yes", "y":
		return true
	case "0", "false", "no", "n":
		return false
	default:
		return fallback
	}
}

func splitCSV(value string) []string {
	if value == "" {
		return nil
	}

	raw := strings.Split(value, ",")
	items := make([]string, 0, len(raw))
	for _, item := range raw {
		item = strings.TrimSpace(item)
		if item != "" {
			items = append(items, item)
		}
	}
	return items
}
