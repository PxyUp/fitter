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

		fmt.Print(colorGreen + r.prompt() + colorReset)

		input, err := r.reader.ReadString('\n')
		if err != nil {
			// io.EOF arrives on Ctrl-D and on a closed pipe.
			fmt.Println("\nGoodbye!")
			return nil
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
		case "new", "reset":
			r.agent.Reset()
			fmt.Printf("%sStarted a new session.%s\n\n", colorYellow, colorReset)
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

// prompt hints whether the next request refines the current config or starts
// something new.
func (r *REPL) prompt() string {
	if r.agent.HasHistory() {
		return "refine> "
	}
	return "> "
}

func (r *REPL) processRequest(ctx context.Context, request string) error {
	if r.agent.HasHistory() {
		fmt.Printf("\n%sRefining config...%s\n", colorGray, colorReset)
	} else {
		fmt.Printf("\n%sGenerating Fitter config...%s\n", colorGray, colorReset)
	}

	result, err := r.agent.GenerateConfig(ctx, request)
	if err != nil {
		return err
	}

	r.printConfig(result)

	if !r.askConfirmation() {
		fmt.Printf("%sSkipped. Describe a change to refine it, or type 'new' to start over.%s\n\n", colorYellow, colorReset)
		return nil
	}

	fmt.Printf("\n%sExecuting...%s\n", colorGray, colorReset)

	output, err := r.agent.Execute(result.Config)
	if err != nil {
		return err
	}

	r.printResult(output)
	return nil
}

func (r *REPL) printWelcome() {
	fmt.Println()
	fmt.Printf("%s╔══════════════════════════════════════════════════════════════╗%s\n", colorCyan, colorReset)
	fmt.Printf("%s║           Fitter Agent - AI-Powered Data Extraction           ║%s\n", colorCyan, colorReset)
	fmt.Printf("%s╚══════════════════════════════════════════════════════════════╝%s\n", colorCyan, colorReset)
	fmt.Println()
	fmt.Println("Describe what you want to extract. Follow-up messages refine the")
	fmt.Println("previous config. Type 'help' for commands.")
	fmt.Println()
}

func (r *REPL) printHelp() {
	fmt.Println()
	fmt.Printf("%sAvailable Commands:%s\n", colorYellow, colorReset)
	fmt.Println("  help, h, ?     - Show this help message")
	fmt.Println("  new, reset     - Forget the current config and start fresh")
	fmt.Println("  clear, cls     - Clear the screen")
	fmt.Println("  exit, quit, q  - Exit the agent")
	fmt.Println()
	fmt.Printf("%sExample Requests:%s\n", colorYellow, colorReset)
	fmt.Println("  - Get the top 10 stories from HackerNews with titles and scores")
	fmt.Println("  - Fetch Bitcoin price from CoinGecko API")
	fmt.Println("  - Scrape headlines from news.ycombinator.com")
	fmt.Println("  - Get weather data for London from wttr.in")
	fmt.Println()
	fmt.Printf("%sRefining:%s\n", colorYellow, colorReset)
	fmt.Println("  After a config is generated, just say what to change:")
	fmt.Println("  - Only return 5 items")
	fmt.Println("  - Also include the article URL")
	fmt.Println()
}

func (r *REPL) printConfig(result *Result) {
	fmt.Println()
	if result.Notes != "" {
		fmt.Printf("%s%s%s\n\n", colorGray, result.Notes, colorReset)
	}
	fmt.Printf("%s┌─ Generated Fitter Config ───────────────────────────────────────%s\n", colorBlue, colorReset)
	fmt.Println(result.JSON)
	fmt.Printf("%s└──────────────────────────────────────────────────────────────────%s\n", colorBlue, colorReset)
	fmt.Println()
}

func (r *REPL) printResult(output string) {
	fmt.Println()
	fmt.Printf("%s┌─ Result ────────────────────────────────────────────────────────%s\n", colorGreen, colorReset)
	fmt.Println(output)
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
