package uinput

import (
	"fmt"
	"log/slog"

	"github.com/bendahl/uinput"
)

// Device wraps a uinput keyboard device
type Device struct {
	kb uinput.Keyboard
}

// New creates a new uinput keyboard device
func New() (*Device, error) {
	kb, err := uinput.CreateKeyboard("/dev/uinput", []byte("TRS-80 M100 Keyboard"))
	if err != nil {
		return nil, fmt.Errorf("failed to create uinput device: %w", err)
	}

	slog.Info("Created uinput keyboard device")
	return &Device{kb: kb}, nil
}

// EmitKey sends a single key press and release
func (d *Device) EmitKey(key int) error {
	if err := d.kb.KeyPress(key); err != nil {
		return fmt.Errorf("failed to press key %d: %w", key, err)
	}
	return nil
}

// EmitCombo sends a key combination (e.g., Ctrl+C)
func (d *Device) EmitCombo(modifiers []int, key int) error {
	// Press all modifiers
	for _, mod := range modifiers {
		if err := d.kb.KeyDown(mod); err != nil {
			return fmt.Errorf("failed to press modifier %d: %w", mod, err)
		}
	}

	// Press and release the key
	if err := d.kb.KeyPress(key); err != nil {
		return fmt.Errorf("failed to press key %d: %w", key, err)
	}

	// Release all modifiers
	for i := len(modifiers) - 1; i >= 0; i-- {
		if err := d.kb.KeyUp(modifiers[i]); err != nil {
			return fmt.Errorf("failed to release modifier %d: %w", modifiers[i], err)
		}
	}

	return nil
}

// Close closes the uinput device
func (d *Device) Close() error {
	if err := d.kb.Close(); err != nil {
		return fmt.Errorf("failed to close uinput device: %w", err)
	}
	slog.Info("Closed uinput keyboard device")
	return nil
}
