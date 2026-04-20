/*
 * Project: ${projectName}
 * Template: Console CLI – SunGo Project Manager
 *
 * Features:
 *   - Command-line flags (--name, --verbose, --count)
 *   - Colored terminal output (ANSI)
 *   - Simple structured logger
 *   - Graceful exit with defer
 */

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// ── ANSI color helpers ────────────────────────────────────────────────────────

const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

func green(s string) string  { return colorGreen + s + colorReset }
func yellow(s string) string { return colorYellow + s + colorReset }
func red(s string) string    { return colorRed + s + colorReset }
func cyan(s string) string   { return colorCyan + s + colorReset }
func bold(s string) string   { return colorBold + s + colorReset }

// ── Logger ────────────────────────────────────────────────────────────────────

type Logger struct {
	verbose bool
	log     *log.Logger
}

func NewLogger(verbose bool) *Logger {
	return &Logger{
		verbose: verbose,
		log:     log.New(os.Stdout, "", 0),
	}
}

func (l *Logger) Info(format string, args ...any) {
	l.log.Printf(green("[INFO]")+" "+format, args...)
}

func (l *Logger) Warn(format string, args ...any) {
	l.log.Printf(yellow("[WARN]")+" "+format, args...)
}

func (l *Logger) Error(format string, args ...any) {
	l.log.Printf(red("[ERR] ")+" "+format, args...)
}

func (l *Logger) Debug(format string, args ...any) {
	if !l.verbose {
		return
	}
	l.log.Printf(cyan("[DBG] ")+" "+format, args...)
}

// ── Config ────────────────────────────────────────────────────────────────────

type Config struct {
	Name    string
	Count   int
	Verbose bool
}

func parseFlags() Config {
	name    := flag.String("name",    "World",  "Name to greet")
	count   := flag.Int("count",     1,        "Number of greetings")
	verbose := flag.Bool("verbose",  false,    "Enable verbose output")
	flag.Parse()

	return Config{
		Name:    *name,
		Count:   *count,
		Verbose: *verbose,
	}
}

// ── Main logic ────────────────────────────────────────────────────────────────

func run(cfg Config, logger *Logger) error {
	logger.Debug("Starting %s with config: name=%s count=%d",
		bold("${projectName}"), cfg.Name, cfg.Count)

	start := time.Now()
	defer func() {
		logger.Info("Done in %s", time.Since(start).Round(time.Millisecond))
	}()

	// Banner
	border := strings.Repeat("─", 40)
	fmt.Println(cyan(border))
	fmt.Printf("%s %s\n", bold("Project:"), green("${projectName}"))
	fmt.Println(cyan(border))

	// Work
	for i := range cfg.Count {
		logger.Info("Hello, %s! (iteration %d/%d)", bold(cfg.Name), i+1, cfg.Count)
	}

	// Example warning
	if cfg.Count > 5 {
		logger.Warn("That's a lot of greetings (%d)!", cfg.Count)
	}

	return nil
}

func main() {
	cfg    := parseFlags()
	logger := NewLogger(cfg.Verbose)

	if err := run(cfg, logger); err != nil {
		logger.Error("%v", err)
		os.Exit(1)
	}
}
