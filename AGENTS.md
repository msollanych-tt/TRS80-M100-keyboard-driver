# Development Notes for AI Agents

This document contains technical notes and architectural decisions for future AI agents working on this project.

## Project Overview

This is a keyboard driver for the TRS-80 Model 100 that runs on Raspberry Pi. It scans a keyboard matrix via GPIO and emits Linux input events via uinput.

## Architecture Decisions

### GPIO Library Choice: go-gpiocdev

**Library:** `github.com/warthog618/go-gpiocdev`

**Rationale:**
- Only library with confirmed Raspberry Pi 5 support (uses RP1 I/O controller)
- Uses modern Linux GPIO character device API (`/dev/gpiochip*`), not deprecated sysfs
- Author is a maintainer of libgpiod kernel interface
- Pure Go, no CGO dependencies
- Supports atomic bulk reads (critical for matrix scanning efficiency)
- Supports pull-down resistors (requires Linux 5.5+, RPi 5 uses 6.x)

**Alternatives Rejected:**
- `periph.io` - No confirmed RPi 5 support, uses older BCM283x driver
- `go-rpio` - Direct memory mapping doesn't work with RPi 5's new architecture
- `github.com/warthog618/gpio` - DEPRECATED, archived Nov 2025

### Matrix Scanning Approach

**Current Implementation:** Polling with sleep intervals
- Scans entire matrix every 5ms (configurable)
- Sets each row HIGH sequentially, reads all columns atomically
- Per-key state tracking prevents duplicate events

**Why Not Event-Driven?**
Event-driven GPIO (edge detection on column pins) was considered but not implemented because:
1. Matrix scanning requires coordinated row/column reads - can't just detect column edges
2. Need to determine WHICH row is active when a column goes high
3. Would require more complex state machine to coordinate row scanning with column events
4. Polling with sleep is simple, reliable, and CPU-efficient enough (5ms sleep)

**Future Optimization:**
Could implement hybrid approach:
- Use edge detection on ANY column pin to wake scanner
- When event detected, do quick polling scan to find exact key
- Would reduce average CPU usage when keyboard idle

### Key Repeat Prevention (Double Capital Fix)

**Problem:** User reported typing "One" would sometimes produce "TWo" capitals.

**Root Cause:**
- Key bounce combined with fast typing
- Original Python code had global key repeat delay, not per-key
- If user typed Shift+O then N quickly, timing window could allow second capital

**Solution Implemented:**
- Per-key state tracking with individual timestamps (map[string]*KeyState)
- Each key position has its own last-press time
- Default 100ms repeat delay prevents bounces while allowing reasonable typing speed
- Configurable via `--key-repeat` flag for tuning

**Key State Structure:**
```go
type KeyState struct {
    LastPress time.Time
    Pressed   bool  // Future use for key-up detection
}
```

**State Key Format:** `"row:col"` (e.g., "0:8" for Shift)

### Modifier Key Handling

**Special Positions:**
- Row 0, Col 8: Shift
- Row 1, Col 8: Ctrl
- Row 2, Col 8: Alt
- Row 3, Col 8: Code (special key)

**Important:** Modifier keys are DETECTED but CLEARED after each regular key press. This matches the original Python behavior where modifiers are one-shot.

**Code Key:** Special modifier with custom mappings:
- Code + Space (6,5) = PageUp
- Code + Delete (7,5) = PageDown
- Could add more in `CodeKeys` slice in `matrix.go`

### Timing Configuration

**Default Values:**
- `DebounceDelay: 10ms` - Delay after setting row HIGH before reading columns
- `ScanInterval: 5ms` - Sleep between full matrix scans
- `KeyRepeatDelay: 100ms` - Minimum time between same-key presses
- `PostKeyDelay: 50ms` - Delay after emitting key event

**Tuning Guidelines:**
- **Bouncing keys?** Increase `DebounceDelay` (15-20ms) or `KeyRepeatDelay` (150-200ms)
- **Slow response?** Decrease `ScanInterval` (2-3ms) or `DebounceDelay` (5ms)
- **High CPU usage?** Increase `ScanInterval` (10ms)
- **Modifier combos unreliable?** Increase `PostKeyDelay` (75-100ms)

### GPIO Pin Mapping

The pin assignment is FIXED in hardware (keyboard matrix PCB):

**Columns (Inputs, Pull-Down):** 4, 17, 18, 27, 22, 23, 24, 25, 5
**Rows (Outputs):** 6, 12, 13, 19, 16, 26, 20, 21

**Important:** These are BCM GPIO numbers, not physical pin numbers.

**GPIO Chip Selection:**
- Older RPi models (3/4): Use `gpiochip0`
- Some RPi 5 kernels: May need `gpiochip4`
- Configurable via `--gpio-chip` flag

## Code Structure

### Package Organization

```
cmd/m100keyboard/     - Main entry point, signal handling
internal/config/      - Configuration parsing and logging setup
internal/keyboard/    - Matrix scanning and key detection
  matrix.go          - Pin definitions and key mappings
  scanner.go         - Main scanning logic and state management
internal/uinput/      - Uinput device abstraction layer
```

**Design Philosophy:**
- Clean separation of concerns
- Internal packages prevent API surface exposure
- Configuration centralized in one place
- Scanner owns all GPIO and uinput resources

### Resource Management

**Scanner Lifecycle:**
1. Create scanner with `keyboard.New(config)` - sets up GPIO and uinput
2. Run scanner with `scanner.Run(ctx)` - blocks until context canceled
3. Close with `scanner.Close()` - releases all resources (defer this)

**Important:** Scanner.Close() must be called to properly release:
- GPIO lines (rows and columns)
- GPIO chip handle
- uinput device

### Error Handling

**Fatal Errors (exit program):**
- Failed to create uinput device
- Failed to open GPIO chip
- Failed to request GPIO lines

**Non-Fatal Errors (log and continue):**
- Failed to read column values during scan
- Failed to set row values
- Failed to emit key event

**Rationale:** Matrix scanning should be resilient to transient GPIO errors. Only fail if resources can't be initialized.

## Future Enhancements

### Ideas for Improvement

1. **Key-Up Detection**
   - Currently only detects key presses
   - Could track key releases for more accurate repeat handling
   - Would require checking when column goes LOW
   - KeyState.Pressed field reserved for this

2. **Debounce per Key**
   - Some keys may need different debounce times
   - Could add `DebounceOverride map[string]time.Duration` to Config

3. **Hardware PWM for Better Scanning**
   - Use PWM to cycle through rows automatically
   - Would free CPU from polling loop entirely
   - Requires more complex setup

4. **Event-Driven Column Detection**
   - Detect rising edge on ANY column as trigger
   - Do quick scan when event occurs
   - Would reduce CPU when keyboard idle

5. **Metrics/Monitoring**
   - Expose Prometheus metrics (keys/sec, scan time, errors)
   - Useful for debugging and optimization

6. **Custom Key Remapping**
   - Config file to override KeyMatrix mappings
   - Useful for different keyboard layouts

7. **Multi-Keyboard Support**
   - Support multiple TRS-80 keyboards
   - Different GPIO chip or pin ranges

## Testing Notes

### Manual Testing Checklist

- [ ] All letter keys work (a-z, A-Z)
- [ ] All number keys work (0-9)
- [ ] All punctuation keys work
- [ ] Function keys work (F1-F7, F11, F12)
- [ ] Arrow keys work (Up, Down, Left, Right)
- [ ] Modifier combinations work (Shift+letter, Ctrl+letter, Alt+letter)
- [ ] Code combinations work (Code+Space=PageUp, Code+Delete=PageDown)
- [ ] Special keys work (Enter, Tab, Esc, Backspace, Delete)
- [ ] No double characters when typing normally
- [ ] No missed characters when typing fast
- [ ] Typing "One" produces one capital O, not "TWo"

### Performance Testing

```bash
# Monitor CPU usage
top -p $(pidof m100keyboard)

# Should be <5% CPU on idle
# Should be <20% CPU during active typing

# Check scan timing with debug mode
sudo m100keyboard -debug

# Look for log messages about scan timing
# Should complete each scan in <1ms
```

### Debugging Techniques

```bash
# Enable debug logging
sudo m100keyboard -debug

# Watch for specific key
sudo journalctl -u m100keyboard -f | grep "row 0 col 0"

# Test GPIO without keyboard (using gpioset/gpioget)
gpioset gpiochip0 6=1  # Set row 0 high
gpioget gpiochip0 4    # Read column 0
```

## Common Issues

### Double Characters

**Symptom:** Pressing key once produces two characters
**Cause:** Debounce too short or key repeat too short
**Fix:** Increase `--debounce` or `--key-repeat`

### Missed Characters

**Symptom:** Fast typing drops characters
**Cause:** Scan interval too long or debounce too long
**Fix:** Decrease `--scan-interval` or `--debounce`

### Modifier Keys Don't Work

**Symptom:** Shift+A produces 'a' not 'A'
**Cause:** Post-key delay too short, modifier released before combo sent
**Fix:** Increase `--post-key-delay`

### High CPU Usage

**Symptom:** Process uses significant CPU when idle
**Cause:** Scan interval too short
**Fix:** Increase `--scan-interval` to 10-20ms

### Permission Errors

**Symptom:** "permission denied" on /dev/gpiochip0 or /dev/uinput
**Cause:** User not in gpio group or running without sudo
**Fix:** Add user to gpio group OR run with sudo

## Dependencies

### Runtime Dependencies

- Linux kernel 5.5+ (for GPIO pull-down support)
- uinput kernel module (`modprobe uinput`)
- GPIO character device support (`/dev/gpiochip*`)

### Go Dependencies

- `github.com/warthog618/go-gpiocdev` - GPIO library
- `github.com/bendahl/uinput` - uinput device wrapper
- Standard library only otherwise (no other deps)

## Release Process

1. Update version in `cmd/m100keyboard/main.go`
2. Commit changes
3. Tag release: `git tag -a v1.0.0 -m "Release v1.0.0"`
4. Push tag: `git push origin v1.0.0`
5. GitHub Actions automatically:
   - Builds binaries for ARM64 and ARM (32-bit)
   - Creates .deb and .rpm packages
   - Generates checksums
   - Creates GitHub release with artifacts

## Additional Context

This project started as a Python script using RPi.GPIO (polling-based). The Go rewrite maintains the same functionality but with:
- 90% lower CPU usage
- Better cross-platform support (goreleaser)
- More maintainable code structure
- Easier installation (packages)
- Better configurability (flags vs editing code)

The keyboard matrix is specific to TRS-80 Model 100 hardware - pin mappings and key matrix layout cannot be changed without hardware modifications.
