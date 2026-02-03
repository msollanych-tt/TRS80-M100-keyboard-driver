package config

import (
	"flag"
	"log/slog"
	"time"
)

// Config holds all configuration for the keyboard driver
type Config struct {
	// Timing configuration
	DebounceDelay  time.Duration // Delay for debouncing key presses
	ScanInterval   time.Duration // Interval between matrix scans
	KeyRepeatDelay time.Duration // Minimum time between repeated key presses
	PostKeyDelay   time.Duration // Delay after emitting a key event

	// GPIO configuration
	GPIOChip string // GPIO chip device (e.g., "gpiochip0" or "gpiochip4")

	// Logging
	Debug bool
}

// Default returns a Config with sensible defaults
func Default() *Config {
	return &Config{
		DebounceDelay:  10 * time.Millisecond,
		ScanInterval:   5 * time.Millisecond,
		KeyRepeatDelay: 100 * time.Millisecond,
		PostKeyDelay:   50 * time.Millisecond,
		GPIOChip:       "gpiochip0", // Most common for older Pi models
		Debug:          false,
	}
}

// ParseFlags parses command-line flags and returns a Config
func ParseFlags() *Config {
	cfg := Default()

	var debounceMs, scanMs, repeatMs, postKeyMs int

	flag.IntVar(&debounceMs, "debounce", 10, "Debounce delay in milliseconds")
	flag.IntVar(&scanMs, "scan-interval", 5, "Matrix scan interval in milliseconds")
	flag.IntVar(&repeatMs, "key-repeat", 100, "Minimum delay between repeated keypresses in milliseconds")
	flag.IntVar(&postKeyMs, "post-key-delay", 50, "Delay after key emission in milliseconds")
	flag.StringVar(&cfg.GPIOChip, "gpio-chip", "gpiochip0", "GPIO chip device (gpiochip0 or gpiochip4)")
	flag.BoolVar(&cfg.Debug, "debug", false, "Enable debug logging")

	flag.Parse()

	cfg.DebounceDelay = time.Duration(debounceMs) * time.Millisecond
	cfg.ScanInterval = time.Duration(scanMs) * time.Millisecond
	cfg.KeyRepeatDelay = time.Duration(repeatMs) * time.Millisecond
	cfg.PostKeyDelay = time.Duration(postKeyMs) * time.Millisecond

	return cfg
}

// SetupLogging configures slog based on debug flag
func (c *Config) SetupLogging() {
	level := slog.LevelInfo
	if c.Debug {
		level = slog.LevelDebug
	}
	slog.SetLogLoggerLevel(level)
}
