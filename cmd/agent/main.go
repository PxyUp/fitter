package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/PxyUp/fitter/pkg/agent"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/plugins/store"
	"github.com/PxyUp/fitter/pkg/utils"
)

func main() {
	apiKey := flag.String("api-key", "", "Anthropic API key (defaults to $ANTHROPIC_API_KEY)")
	model := flag.String("model", agent.DefaultModel, "Claude model to use")
	effort := flag.String("effort", agent.DefaultEffort, "Reasoning effort (low, medium, high, xhigh, max)")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	plugins := flag.String("plugins", "", "Plugins folder path")

	chromiumLimit := flag.Uint("chromium-limit", 0, "Limit concurrent Chromium instances")
	dockerLimit := flag.Uint("docker-limit", 0, "Limit concurrent Docker containers")
	playwrightLimit := flag.Uint("playwright-limit", 0, "Limit concurrent Playwright instances")

	flag.Parse()

	// Prefer the environment so the key never lands in shell history or ps output.
	if *apiKey == "" && os.Getenv("ANTHROPIC_API_KEY") == "" {
		fmt.Fprintln(os.Stderr, "Error: no Anthropic API key found")
		fmt.Fprintln(os.Stderr, "Set ANTHROPIC_API_KEY, or pass --api-key=<your-anthropic-api-key>")
		os.Exit(1)
	}

	var log logger.Logger = logger.Null
	if *verbose {
		log = logger.NewLogger(*logLevel)
		utils.SetLogger(*logLevel)
	}

	if *plugins != "" {
		if err := store.PluginInitialize(*plugins); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading plugins: %s\n", err.Error())
			os.Exit(1)
		}
	}

	var limits *config.Limits
	if *chromiumLimit > 0 || *dockerLimit > 0 || *playwrightLimit > 0 {
		limits = &config.Limits{
			ChromiumInstance:   uint32(*chromiumLimit),
			DockerContainers:   uint32(*dockerLimit),
			PlaywrightInstance: uint32(*playwrightLimit),
		}
	}

	ag, err := agent.NewAgent(agent.AgentConfig{
		APIKey:  *apiKey,
		Model:   *model,
		Effort:  *effort,
		Logger:  log,
		Limits:  limits,
		Verbose: *verbose,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating agent: %s\n", err.Error())
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nInterrupted. Exiting...")
		cancel()
	}()

	repl := agent.NewREPL(ag, *verbose)
	if err := repl.Run(ctx); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
}
