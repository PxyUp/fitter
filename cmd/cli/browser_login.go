package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/oauthflow"
	"github.com/mxschmitt/playwright-go"
)

func runBrowserLogin(args []string) {
	fs := flag.NewFlagSet("browser-login", flag.ExitOnError)
	urlFlag := fs.String("url", "", "Login page url to open")
	stateFile := fs.String("storage-state", "", "Path to save the session storage state, e.g. ~/.fitter/sessions/site.json")
	browserFlag := fs.String("browser", string(config.Chromium), "Chromium|FireFox|WebKit — use the same browser as the scraping config")
	install := fs.Bool("install", false, "Install playwright browsers first")
	indexedDB := fs.Bool("indexeddb", false, "Include IndexedDB in the snapshot (Firebase Auth and similar)")
	_ = fs.Parse(args)

	if *urlFlag == "" {
		fmt.Fprintln(os.Stderr, "--url is required")
		os.Exit(1)
	}
	if *stateFile == "" {
		fmt.Fprintln(os.Stderr, "--storage-state is required")
		os.Exit(1)
	}

	if *install {
		err := playwright.Install(&playwright.RunOptions{Verbose: false})
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to install playwright: %s\n", err.Error())
			os.Exit(1)
		}
	}

	pw, err := playwright.Run(&playwright.RunOptions{Verbose: false})
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not start playwright: %s\n", err.Error())
		os.Exit(1)
	}
	defer func() {
		_ = pw.Stop()
	}()

	var browserType playwright.BrowserType
	switch config.PlaywrightBrowser(*browserFlag) {
	case config.Chromium:
		browserType = pw.Chromium
	case config.FireFox:
		browserType = pw.Firefox
	case config.WebKit:
		browserType = pw.WebKit
	default:
		fmt.Fprintf(os.Stderr, "unknown browser %q, use Chromium|FireFox|WebKit\n", *browserFlag)
		os.Exit(1)
	}

	browser, err := browserType.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not launch browser: %s\n", err.Error())
		os.Exit(1)
	}
	defer func() {
		_ = browser.Close()
	}()

	var contextOpts playwright.BrowserNewContextOptions
	expandedStateFile := oauthflow.ExpandPath(*stateFile)
	if _, errStat := os.Stat(expandedStateFile); errStat == nil {
		contextOpts.StorageStatePath = playwright.String(expandedStateFile)
		fmt.Fprintln(os.Stderr, "Existing storage state loaded, the session will be extended")
	}

	browserCtx, err := browser.NewContext(contextOpts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not create browser context: %s\n", err.Error())
		os.Exit(1)
	}

	page, err := browserCtx.NewPage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not create page: %s\n", err.Error())
		os.Exit(1)
	}

	_, err = page.Goto(*urlFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not open %s: %s\n", *urlFlag, err.Error())
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "Log in in the opened browser window.")
	fmt.Fprint(os.Stderr, "Press Enter here when you are done to save the session... ")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')

	state, err := browserCtx.StorageState(playwright.BrowserContextStorageStateOptions{
		IndexedDB: playwright.Bool(*indexedDB),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to get storage state: %s\n", err.Error())
		os.Exit(1)
	}

	err = oauthflow.SaveJSONFile(expandedStateFile, state)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to save storage state: %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Session saved to %s (%d cookies, %d origins)\n", expandedStateFile, len(state.Cookies), len(state.Origins))
	fmt.Fprintln(os.Stderr, "\nUse it in a browser_config like this:")
	fmt.Fprintf(os.Stdout, `{
  "playwright": {
    "browser": %q,
    "storage_state_file": %q
  }
}
`, *browserFlag, *stateFile)
}
