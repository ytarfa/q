package main

import (
	"fmt"
	"os"
	"strings"
)

const version = "0.1.0"

func main() {
	os.Exit(run())
}

func run() int {
	// Check for init subcommand before flag parsing
	if len(os.Args) > 1 && os.Args[1] == "init" {
		return runInit()
	}

	// Parse flags
	args, limit, flagLimit, err := parseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Handle --version
	if limit == -2 { // sentinel: version requested
		fmt.Println("q " + version)
		return 0
	}

	// Read stdin if piped
	stdinContent, truncated, err := readStdin()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
		return 1
	}
	if truncated {
		fmt.Fprintln(os.Stderr, "Warning: input truncated to 100KB")
	}

	// Resolve question and context
	question, context, err := resolveInput(args, stdinContent)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	// Load config
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}

	// Apply env var overrides
	applyEnvOverrides(cfg)

	// Apply flag overrides
	if flagLimit {
		cfg.Limit = limit
	}

	// Resolve provider defaults
	resolveDefaults(cfg)

	// Validate config
	if err := validateConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}

	// Build prompt
	messages := buildMessages(question, context, cfg.Limit)

	// Call LLM
	response, err := callLLM(cfg, messages)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 3
	}

	// Truncate if needed
	output := truncateResponse(response, cfg.Limit)

	// Print to stdout
	fmt.Println(output)
	return 0
}

func printUsage() {
	usage := `Usage: q [flags] [question...]

Ask an LLM a question from your terminal.

Flags:
  -l, --limit int    Max response length in characters (default 200, 0 = unlimited)
  -v, --version      Print version
  -h, --help         Print help

Subcommands:
  init               Create config file interactively

Examples:
  q what is a goroutine
  q -l 500 explain the CAP theorem
  q -l 0 write a haiku about recursion
  echo "func main() {}" | q
  cat error.log | q what does this mean

Environment:
  Q_API_KEY          API key (overrides config file)
  Q_MODEL            Model name (overrides config file)
  Q_PROVIDER         Provider type (overrides config file)
  Q_LIMIT            Default character limit (overrides config file)

Config: ~/.config/q/config.json
`
	fmt.Print(usage)
}

// parseFlags parses CLI flags and returns remaining args, limit value,
// whether limit was explicitly set, and any error.
// Returns limit=-2 as a sentinel for --version.
func parseFlags(args []string) (remaining []string, limit int, flagLimit bool, err error) {
	limit = -1 // sentinel: not set by flag
	flagLimit = false

	i := 0
	for i < len(args) {
		arg := args[i]

		switch {
		case arg == "-h" || arg == "--help":
			printUsage()
			os.Exit(0)

		case arg == "-v" || arg == "--version":
			return nil, -2, false, nil

		case arg == "-l" || arg == "--limit":
			if i+1 >= len(args) {
				return nil, 0, false, fmt.Errorf("flag %s requires a value", arg)
			}
			i++
			n, parseErr := parseInt(args[i])
			if parseErr != nil {
				return nil, 0, false, fmt.Errorf("invalid limit value: %s", args[i])
			}
			if n < 0 {
				return nil, 0, false, fmt.Errorf("limit must be non-negative")
			}
			limit = n
			flagLimit = true

		case strings.HasPrefix(arg, "-l="):
			val := strings.TrimPrefix(arg, "-l=")
			n, parseErr := parseInt(val)
			if parseErr != nil {
				return nil, 0, false, fmt.Errorf("invalid limit value: %s", val)
			}
			if n < 0 {
				return nil, 0, false, fmt.Errorf("limit must be non-negative")
			}
			limit = n
			flagLimit = true

		case strings.HasPrefix(arg, "--limit="):
			val := strings.TrimPrefix(arg, "--limit=")
			n, parseErr := parseInt(val)
			if parseErr != nil {
				return nil, 0, false, fmt.Errorf("invalid limit value: %s", val)
			}
			if n < 0 {
				return nil, 0, false, fmt.Errorf("limit must be non-negative")
			}
			limit = n
			flagLimit = true

		case strings.HasPrefix(arg, "-") && arg != "-":
			return nil, 0, false, fmt.Errorf("unknown flag: %s", arg)

		default:
			// Everything from here on is the question
			remaining = append(remaining, args[i:]...)
			return remaining, limit, flagLimit, nil
		}

		i++
	}

	return remaining, limit, flagLimit, nil
}

func parseInt(s string) (int, error) {
	n := 0
	neg := false
	if len(s) == 0 {
		return 0, fmt.Errorf("empty string")
	}
	start := 0
	if s[0] == '-' {
		neg = true
		start = 1
	}
	if start >= len(s) {
		return 0, fmt.Errorf("invalid number")
	}
	for i := start; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, fmt.Errorf("invalid number")
		}
		n = n*10 + int(s[i]-'0')
	}
	if neg {
		n = -n
	}
	return n, nil
}

// resolveInput determines the question and optional context from args and stdin.
func resolveInput(args []string, stdinContent string) (question, context string, err error) {
	hasArgs := len(args) > 0
	hasStdin := stdinContent != ""

	switch {
	case hasArgs && hasStdin:
		return strings.Join(args, " "), stdinContent, nil
	case hasArgs && !hasStdin:
		return strings.Join(args, " "), "", nil
	case !hasArgs && hasStdin:
		return "Explain this", stdinContent, nil
	default:
		return "", "", fmt.Errorf("Usage: q [flags] [question...]")
	}
}
