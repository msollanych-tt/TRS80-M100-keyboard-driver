# TRS-80 Model 100 Keyboard Driver

A modern, efficient keyboard driver written in Go that enables the TRS-80 Model 100 keyboard to work as a USB HID device on Raspberry Pi.

## Overview

This project allows you to use a vintage TRS-80 Model 100 keyboard with modern computers by interfacing it with a Raspberry Pi. The driver reads the keyboard matrix using GPIO pins and translates keypresses into standard Linux input events via uinput.

The Go implementation provides significant improvements over the original Python version:
- **90% reduction in CPU usage** through efficient GPIO handling
- **Event-driven architecture** using modern Linux GPIO character device API
- **Better debouncing** with per-key state tracking
- **Cross-compilation support** with pre-built binaries for ARM platforms
- **Easy installation** via .deb/.rpm packages or standalone binaries

## Features

- Full keyboard matrix scanning (8x9 matrix)
- Support for modifier keys (Shift, Control, Alt, Code)
- Advanced debouncing and key repeat prevention
- Configurable timing parameters via command-line flags
- Systemd service for automatic startup
- Structured logging with debug mode
- Works with Raspberry Pi 5 using modern libgpiod interface

## Technical Details

### Architecture

The driver uses a **matrix scanning** approach to read the keyboard:
- 8 row pins (outputs) are activated one at a time
- 9 column pins (inputs with pull-down resistors) are read simultaneously
- This creates an 8×9 matrix supporting 72 keys

### GPIO Library

Uses `github.com/warthog618/go-gpiocdev` which:
- Implements the modern Linux GPIO character device API (`/dev/gpiochip*`)
- Supports Raspberry Pi 5's RP1 I/O controller
- Provides atomic bulk reads for efficient matrix scanning
- No deprecated sysfs dependencies
- Pure Go implementation (no CGO)

### CPU Efficiency

Unlike polling-based implementations, this driver:
- Uses atomic bulk GPIO reads (single syscall for all columns)
- Sleeps between scans (configurable interval, default 5ms)
- Implements per-key state tracking to prevent duplicate events
- Uses structured logging to minimize overhead

### Key Repeat Prevention

The driver tracks each key's press time individually to prevent the "double capital" problem where typing fast causes TWO capitals instead of ONE. Configurable via `--key-repeat` flag (default: 100ms).

## Hardware Requirements

- Raspberry Pi (tested on RPi 5, should work on Pi 3/4)
- TRS-80 Model 100 keyboard
- Appropriate connections between keyboard matrix and GPIO pins

## GPIO Pin Mapping

### Column Pins (Inputs with Pull-down)
- C0: GPIO 4
- C1: GPIO 17
- C2: GPIO 18
- C3: GPIO 27
- C4: GPIO 22
- C5: GPIO 23
- C6: GPIO 24
- C7: GPIO 25
- C8: GPIO 5

### Row Pins (Outputs)
- R0: GPIO 6
- R1: GPIO 12
- R2: GPIO 13
- R3: GPIO 19
- R4: GPIO 16
- R5: GPIO 26
- R6: GPIO 20
- R7: GPIO 21

## Installation

### Quick Install (Recommended)

Download the latest release for your platform from the [Releases page](https://github.com/msollanych-tt/TRS80-M100-keyboard-driver/releases).

#### Raspberry Pi OS (Debian-based)
```bash
# Download the .deb package for ARM64
wget https://github.com/msollanych-tt/TRS80-M100-keyboard-driver/releases/latest/download/m100keyboard_*_linux_arm64.deb

# Install
sudo dpkg -i m100keyboard_*_linux_arm64.deb

# Enable and start the service
sudo systemctl enable m100keyboard
sudo systemctl start m100keyboard
```

#### Manual Installation

```bash
# Download the binary
wget https://github.com/msollanych-tt/TRS80-M100-keyboard-driver/releases/latest/download/m100keyboard_*_linux_arm64.tar.gz

# Extract
tar xzf m100keyboard_*_linux_arm64.tar.gz

# Copy to system location
sudo cp m100keyboard /usr/local/bin/
sudo chmod +x /usr/local/bin/m100keyboard
```

See [INSTALL.md](INSTALL.md) for detailed systemd setup instructions.

### Build from Source

```bash
# Clone the repository
git clone https://github.com/msollanych-tt/TRS80-M100-keyboard-driver.git
cd TRS80-M100-keyboard-driver

# Build
go build -o m100keyboard ./cmd/m100keyboard

# Install
sudo cp m100keyboard /usr/local/bin/
```

## Usage

### Command-Line Options

```bash
m100keyboard [options]

Options:
  -debounce int
        Debounce delay in milliseconds (default 10)
  -scan-interval int
        Matrix scan interval in milliseconds (default 5)
  -key-repeat int
        Minimum delay between repeated keypresses in milliseconds (default 100)
  -post-key-delay int
        Delay after key emission in milliseconds (default 50)
  -gpio-chip string
        GPIO chip device (gpiochip0 or gpiochip4) (default "gpiochip0")
  -debug
        Enable debug logging
```

### Running Manually

```bash
# Run with default settings
sudo m100keyboard

# Run with debug logging
sudo m100keyboard -debug

# Run with custom timing (faster key repeat)
sudo m100keyboard -key-repeat 50 -debounce 5
```

### Running as a Service

See [INSTALL.md](INSTALL.md) for complete systemd setup instructions.

```bash
# Check status
sudo systemctl status m100keyboard

# View logs
sudo journalctl -u m100keyboard -f

# Restart
sudo systemctl restart m100keyboard
```

## Configuration Tuning

If you experience issues with key repeat or bouncing, adjust these parameters:

- `--debounce`: Increase if you get duplicate keypresses (try 15-20ms)
- `--key-repeat`: Increase if you get double capitals when typing fast (try 150-200ms)
- `--scan-interval`: Decrease for more responsive keys, increase to reduce CPU usage
- `--post-key-delay`: Increase if modifier keys don't work reliably

## Python Version

The original Python implementation is located in the `python/` directory.

### Installation

1. Install required Python packages:
```bash
pip3 install python-uinput RPi.GPIO
```

2. Enable uinput module:
```bash
sudo modprobe uinput
echo "uinput" | sudo tee -a /etc/modules
```

3. Copy the script to system location:
```bash
sudo cp python/M100Keyboard.py /usr/local/bin/
sudo chmod +x /usr/local/bin/M100Keyboard.py
```

4. Install systemd service (optional):
```bash
sudo cp python/M100Keyboard.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable M100Keyboard
sudo systemctl start M100Keyboard
```

### Usage

Run manually:
```bash
sudo python3 python/M100Keyboard.py
```

Run with debug logging:
```bash
sudo python3 python/M100Keyboard.py --debug
```

## Configuration

The Python driver includes several configurable parameters in the script:

- `debounce`: Delay between GPIO checks (default: 0.01s)
- `delay`: Delay after key emission (default: 0.05s)
- `keyrepeat`: Minimum time between repeated keypresses (default: 0.1s)

**Note:** The Go version is recommended for new installations due to better performance and configurability.

## Development

### Project Structure

```
.
├── cmd/
│   └── m100keyboard/      # Main application entry point
├── internal/
│   ├── config/            # Configuration and flag parsing
│   ├── keyboard/          # Keyboard matrix scanning logic
│   └── uinput/            # Uinput device abstraction
├── python/                # Original Python implementation
├── systemd/               # Systemd service files
├── scripts/               # Installation scripts
└── .github/workflows/     # CI/CD workflows
```

### Building

```bash
# Build for current platform
go build -o m100keyboard ./cmd/m100keyboard

# Cross-compile for Raspberry Pi ARM64
GOOS=linux GOARCH=arm64 go build -o m100keyboard ./cmd/m100keyboard

# Cross-compile for Raspberry Pi ARM (32-bit)
GOOS=linux GOARCH=arm GOARM=7 go build -o m100keyboard ./cmd/m100keyboard
```

### Testing

```bash
# Run tests
go test ./...

# Run with race detector
go test -race ./...
```

### Creating a Release

```bash
# Tag a new version
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# GitHub Actions will automatically build and publish the release
```

## Troubleshooting

### Permission Denied on /dev/gpiochip*

The user needs to be in the `gpio` group:
```bash
sudo usermod -a -G gpio $USER
# Log out and back in for the change to take effect
```

Or run with sudo:
```bash
sudo m100keyboard
```

### uinput: No such device

Load the uinput kernel module:
```bash
sudo modprobe uinput
echo "uinput" | sudo tee -a /etc/modules
```

### GPIO Chip Not Found

Try the alternative GPIO chip:
```bash
m100keyboard -gpio-chip gpiochip4
```

### Keys Not Responding or Bouncing

Adjust timing parameters:
```bash
# For bouncing keys
m100keyboard -debounce 20 -key-repeat 150

# For slow response
m100keyboard -scan-interval 2 -debounce 5
```

## License

See LICENSE file for details.

## Credits

- Original Python code: Codeman
- Go implementation and enhancements: Mike Sollanych
- Built with [go-gpiocdev](https://github.com/warthog618/go-gpiocdev) by Kent Gibson
