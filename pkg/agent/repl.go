package agent

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorRed    = "\033[31m"
	colorGray   = "\033[90m"
)

type REPL struct {
	agent   *Agent
	reader  *bufio.Reader
	verbose bool
}

func NewREPL(agent *Agent, verbose bool) *REPL {
	return &REPL{
		agent:   agent,
		reader:  bufio.NewReader(os.Stdin),
		verbose: verbose,
	}
}

func (r *REPL) Run(ctx context.Context) error {
	r.printWelcome()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nGoodbye!")
			return ctx.Err()
		default:
		}

		fmt.Print(colorGreen + "> " + colorReset)

		input, err := r.reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Println("\nGoodbye!")
				return nil
			}
			return fmt.Errorf("read error: %w", err)
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		switch strings.ToLower(input) {
		case "exit", "quit", "q":
			fmt.Println("Goodbye!")
			return nil
		case "help", "h", "?":
			r.printHelp()
			continue
		case "clear", "cls":
			fmt.Print("\033[H\033[2J")
			r.printWelcome()
			continue
		}

		if err := r.processRequest(ctx, input); err != nil {
			fmt.Printf("%sError: %s%s\n\n", colorRed, err.Error(), colorReset)
		}
	}
}

func (r *REPL) processRequest(ctx context.Context, request string) error {
	fmt.Printf("\n%sGenerating Fitter config...%s\n", colorGray, colorReset)

	cfg, rawJSON, err := r.agent.GenerateConfig(ctx, request)
	if err != nil {
		return err
	}

	r.printConfig(rawJSON)

	if !r.askConfirmation() {
		fmt.Printf("%sSkipped.%s\n\n", colorYellow, colorReset)
		return nil
	}

	fmt.Printf("\n%sExecuting...%s\n", colorGray, colorReset)

	result, err := r.agent.Execute(cfg)
	if err != nil {
		return err
	}

	r.printResult(result)
	return nil
}

func (r *REPL) printWelcome() {
	fmt.Println()
	fmt.Printf("%s╔══════════════════════════════════════════════════════════════╗%s\n", colorCyan, colorReset)
	fmt.Printf("%s║              Fitter Agent - AI-Powered Data Extraction       ║%s\n", colorCyan, colorReset)
	fmt.Printf("%s╚══════════════════════════════════════════════════════════════╝%s\n", colorCyan, colorReset)
	fmt.Println()
	fmt.Println("Enter your request in natural language. Type 'help' for commands.")
	fmt.Println()
}

func (r *REPL) printHelp() {
	fmt.Println()
	fmt.Printf("%sAvailable Commands:%s\n", colorYellow, colorReset)
	fmt.Println("  help, h, ?     - Show this help message")
	fmt.Println("  clear, cls     - Clear the screen")
	fmt.Println("  exit, quit, q  - Exit the agent")
	fmt.Println()
	fmt.Printf("%sExample Requests:%s\n", colorYellow, colorReset)
	fmt.Println("  - Get the top 10 stories from HackerNews with titles and scores")
	fmt.Println("  - Fetch Bitcoin price from CoinGecko API")
	fmt.Println("  - Scrape headlines from news.ycombinator.com")
	fmt.Println("  - Get weather data for London from wttr.in")
	fmt.Println()
}

func (r *REPL) printConfig(rawJSON string) {
	fmt.Println()
	fmt.Printf("%s┌─ Generated Fitter Config ───────────────────────────────────────%s\n", colorBlue, colorReset)
	fmt.Println(rawJSON)
	fmt.Printf("%s└──────────────────────────────────────────────────────────────────%s\n", colorBlue, colorReset)
	fmt.Println()
}

func (r *REPL) printResult(result string) {
	fmt.Println()
	fmt.Printf("%s┌─ Result ────────────────────────────────────────────────────────%s\n", colorGreen, colorReset)
	fmt.Println(result)
	fmt.Printf("%s└──────────────────────────────────────────────────────────────────%s\n", colorGreen, colorReset)
	fmt.Println()
}

func (r *REPL) askConfirmation() bool {
	fmt.Printf("%sExecute this config? [y/n]: %s", colorYellow, colorReset)

	input, err := r.reader.ReadString('\n')
	if err != nil {
		return false
	}

	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}
