package keyboard

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/bendahl/uinput"
	"github.com/msollanych-tt/TRS80-M100-keyboard-driver/internal/config"
	uinputdev "github.com/msollanych-tt/TRS80-M100-keyboard-driver/internal/uinput"
	"github.com/warthog618/go-gpiocdev"
)

// KeyState tracks the state of a single key
type KeyState struct {
	LastPress time.Time
	Pressed   bool
}

// Scanner handles keyboard matrix scanning and input
type Scanner struct {
	config    *config.Config
	device    *uinputdev.Device
	columns   *gpiocdev.Lines
	rows      *gpiocdev.Lines
	chip      *gpiocdev.Chip
	keyStates map[string]*KeyState
	mu        sync.Mutex

	// Modifier state
	shift bool
	ctrl  bool
	alt   bool
	code  bool
}

// New creates a new keyboard scanner
func New(cfg *config.Config) (*Scanner, error) {
	// Create uinput device
	dev, err := uinputdev.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create uinput device: %w", err)
	}

	// Open GPIO chip
	chip, err := gpiocdev.NewChip(cfg.GPIOChip)
	if err != nil {
		dev.Close()
		return nil, fmt.Errorf("failed to open GPIO chip %s: %w", cfg.GPIOChip, err)
	}

	slog.Info("Opened GPIO chip", "chip", cfg.GPIOChip)

	pinCfg := DefaultPinConfig()

	// Request column pins as inputs with pull-down
	columns, err := chip.RequestLines(
		pinCfg.ColumnPins,
		gpiocdev.AsInput,
		gpiocdev.WithPullDown,
	)
	if err != nil {
		chip.Close()
		dev.Close()
		return nil, fmt.Errorf("failed to request column pins: %w", err)
	}

	slog.Info("Configured column pins as inputs with pull-down", "pins", pinCfg.ColumnPins)

	// Request row pins as outputs (initially low)
	rowValues := make([]int, len(pinCfg.RowPins))
	rows, err := chip.RequestLines(
		pinCfg.RowPins,
		gpiocdev.AsOutput(rowValues...),
	)
	if err != nil {
		columns.Close()
		chip.Close()
		dev.Close()
		return nil, fmt.Errorf("failed to request row pins: %w", err)
	}

	slog.Info("Configured row pins as outputs", "pins", pinCfg.RowPins)

	return &Scanner{
		config:    cfg,
		device:    dev,
		columns:   columns,
		rows:      rows,
		chip:      chip,
		keyStates: make(map[string]*KeyState),
	}, nil
}

// Run starts the keyboard scanner
func (s *Scanner) Run(ctx context.Context) error {
	slog.Info("Starting keyboard scanner",
		"scan_interval", s.config.ScanInterval,
		"debounce", s.config.DebounceDelay,
		"key_repeat", s.config.KeyRepeatDelay,
	)

	ticker := time.NewTicker(s.config.ScanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Stopping keyboard scanner")
			return ctx.Err()
		case <-ticker.C:
			s.scanMatrix()
		}
	}
}

// scanMatrix scans the entire keyboard matrix
func (s *Scanner) scanMatrix() {
	pinCfg := DefaultPinConfig()

	for rowIdx := range pinCfg.RowPins {
		// Set this row high
		s.setRow(rowIdx, 1)

		// Small delay for signal to stabilize
		time.Sleep(s.config.DebounceDelay)

		// Read all columns
		colValues := make([]int, len(pinCfg.ColumnPins))
		if err := s.columns.Values(colValues); err != nil {
			slog.Error("Failed to read column values", "error", err)
			s.setRow(rowIdx, 0)
			continue
		}

		// IMPORTANT: Pre-scan for modifiers in this row BEFORE processing keys
		// This ensures Shift+A works (A is col 1, Shift is col 8)
		rowHasShift := colValues[ShiftCol] == 1 && (rowIdx == 0 || rowIdx == 4 || rowIdx == 5 || rowIdx == 6)
		rowHasCtrl := rowIdx == CtrlRow && colValues[CtrlCol] == 1
		rowHasAlt := rowIdx == AltRow && colValues[AltCol] == 1

		// Check each column for state changes
		for colIdx, value := range colValues {
			keyID := fmt.Sprintf("%d:%d", rowIdx, colIdx)
			s.mu.Lock()
			state, exists := s.keyStates[keyID]
			if !exists {
				state = &KeyState{Pressed: false}
				s.keyStates[keyID] = state
			}

			// Detect key state transitions
			if value == 1 && !state.Pressed {
				// Set modifiers from current row scan before emitting regular keys
				if rowHasShift {
					s.shift = true
				}
				if rowHasCtrl {
					s.ctrl = true
				}
				if rowHasAlt {
					s.alt = true
				}

				// LOW->HIGH: Key just pressed
				state.Pressed = true
				s.mu.Unlock()
				s.handleKeyPress(rowIdx, colIdx)
			} else if value == 0 && state.Pressed {
				// HIGH->LOW: Key just released
				state.Pressed = false
				s.mu.Unlock()
				s.handleKeyRelease(rowIdx, colIdx)
			} else {
				// No state change
				s.mu.Unlock()
			}
		}

		// Set row back to low
		s.setRow(rowIdx, 0)
	}
}

// setRow sets a specific row pin to the given value
func (s *Scanner) setRow(rowIdx, value int) {
	pinCfg := DefaultPinConfig()
	values := make([]int, len(pinCfg.RowPins))
	values[rowIdx] = value

	if err := s.rows.SetValues(values); err != nil {
		slog.Error("Failed to set row value", "row", rowIdx, "value", value, "error", err)
	}
}

// handleKeyPress handles a detected key press (only called on key down transition)
func (s *Scanner) handleKeyPress(row, col int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Handle modifier keys
	if s.handleModifier(row, col) {
		return
	}

	// Handle CODE key combinations
	if s.code && s.handleCodeKey(row, col) {
		s.code = false
		return
	}

	// Get the key code
	if row >= len(KeyMatrix) || col >= len(KeyMatrix[row]) {
		slog.Warn("Invalid key position", "row", row, "col", col)
		return
	}

	key := KeyMatrix[row][col]

	slog.Debug("Key pressed", "row", row, "col", col, "key", key,
		"shift", s.shift, "ctrl", s.ctrl, "alt", s.alt, "code", s.code)

	// Emit the key with modifiers
	if err := s.emitKey(key); err != nil {
		slog.Error("Failed to emit key", "error", err)
	}

	// Only clear CODE key (it's one-shot)
	// Shift, Ctrl, Alt stay active until released
	s.code = false

	// Post-key delay
	time.Sleep(s.config.PostKeyDelay)
}

// handleKeyRelease handles a detected key release
func (s *Scanner) handleKeyRelease(row, col int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if this is a modifier key being released
	switch {
	case col == ShiftCol && (row == 0 || row == 4 || row == 5 || row == 6):
		// Rows 0, 4, 5, 6 all have Shift in column 8
		s.shift = false
		slog.Debug("Shift key released", "row", row)
		// Emit the actual shift key up event
		if err := s.device.KeyUp(KeyMatrix[row][col]); err != nil {
			slog.Error("Failed to emit shift key up", "error", err)
		}
	case row == CtrlRow && col == CtrlCol:
		s.ctrl = false
		slog.Debug("Ctrl key released")
		// Emit the actual ctrl key up event
		if err := s.device.KeyUp(KeyMatrix[row][col]); err != nil {
			slog.Error("Failed to emit ctrl key up", "error", err)
		}
	case row == AltRow && col == AltCol:
		s.alt = false
		slog.Debug("Alt key released")
		// Emit the actual alt key up event
		if err := s.device.KeyUp(KeyMatrix[row][col]); err != nil {
			slog.Error("Failed to emit alt key up", "error", err)
		}
	case row == CodeRow && col == CodeCol:
		s.code = false
		slog.Debug("Code key released")
		// CODE key is special - don't emit it
	default:
		// Regular key released - just log it
		slog.Debug("Key released", "row", row, "col", col)
	}
}

// handleModifier checks if this is a modifier key and updates state
func (s *Scanner) handleModifier(row, col int) bool {
	switch {
	case col == ShiftCol && (row == 0 || row == 4 || row == 5 || row == 6):
		// Rows 0, 4, 5, 6 all have Shift in column 8
		s.shift = true
		slog.Debug("Shift key detected", "row", row)
		// Emit the actual shift key down event
		if err := s.device.KeyDown(KeyMatrix[row][col]); err != nil {
			slog.Error("Failed to emit shift key down", "error", err)
		}
		return true
	case row == CtrlRow && col == CtrlCol:
		s.ctrl = true
		slog.Debug("Ctrl key detected")
		// Emit the actual ctrl key down event
		if err := s.device.KeyDown(KeyMatrix[row][col]); err != nil {
			slog.Error("Failed to emit ctrl key down", "error", err)
		}
		return true
	case row == AltRow && col == AltCol:
		s.alt = true
		slog.Debug("Alt key detected")
		// Emit the actual alt key down event
		if err := s.device.KeyDown(KeyMatrix[row][col]); err != nil {
			slog.Error("Failed to emit alt key down", "error", err)
		}
		return true
	case row == CodeRow && col == CodeCol:
		s.code = true
		slog.Debug("Code key detected")
		// CODE key is special - don't emit it
		return true
	}
	return false
}

// handleCodeKey checks if this is a special CODE key combination
func (s *Scanner) handleCodeKey(row, col int) bool {
	for _, ck := range CodeKeys {
		if ck.Row == row && ck.Col == col {
			slog.Debug("Code combination detected", "row", row, "col", col, "key", ck.Key)
			if err := s.device.EmitKey(ck.Key); err != nil {
				slog.Error("Failed to emit code key", "error", err)
			}
			return true
		}
	}
	return false
}

// emitKey emits a key press with the current modifier state
func (s *Scanner) emitKey(key int) error {
	if s.ctrl {
		return s.device.EmitCombo([]int{uinput.KeyLeftctrl}, key)
	}
	if s.shift {
		return s.device.EmitCombo([]int{uinput.KeyLeftshift}, key)
	}
	if s.alt {
		return s.device.EmitCombo([]int{uinput.KeyLeftalt}, key)
	}
	return s.device.EmitKey(key)
}

// Close closes the scanner and releases resources
func (s *Scanner) Close() error {
	slog.Info("Closing keyboard scanner")

	var errs []error

	if s.rows != nil {
		if err := s.rows.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close row pins: %w", err))
		}
	}

	if s.columns != nil {
		if err := s.columns.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close column pins: %w", err))
		}
	}

	if s.chip != nil {
		if err := s.chip.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close GPIO chip: %w", err))
		}
	}

	if s.device != nil {
		if err := s.device.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close uinput device: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing scanner: %v", errs)
	}

	return nil
}
