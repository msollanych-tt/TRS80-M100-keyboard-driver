# Installation Guide

This guide covers installing and configuring the TRS-80 Model 100 Keyboard Driver on Raspberry Pi.

## Prerequisites

- Raspberry Pi (tested on Pi 5, should work on Pi 3/4)
- Raspberry Pi OS (Raspbian) or compatible Linux distribution
- TRS-80 Model 100 keyboard properly wired to GPIO pins
- Root or sudo access

## Quick Install

### Option 1: Using .deb Package (Recommended for Debian/Raspbian)

```bash
# Download the latest .deb package for ARM64
wget https://github.com/msollanych-tt/TRS80-M100-keyboard-driver/releases/latest/download/m100keyboard_*_linux_arm64.deb

# Install the package
sudo dpkg -i m100keyboard_*_linux_arm64.deb

# The package automatically:
# - Installs the binary to /usr/local/bin/m100keyboard
# - Installs systemd service file
# - Enables uinput kernel module
# - Reloads systemd daemon
```

### Option 2: Using Binary Archive

```bash
# Download the binary archive
wget https://github.com/msollanych-tt/TRS80-M100-keyboard-driver/releases/latest/download/m100keyboard_*_linux_arm64.tar.gz

# Extract the archive
tar xzf m100keyboard_*_linux_arm64.tar.gz

# Install the binary
sudo cp m100keyboard /usr/local/bin/
sudo chmod +x /usr/local/bin/m100keyboard

# Install systemd service file
sudo cp systemd/m100keyboard.service /etc/systemd/system/
```

### Option 3: Build from Source

```bash
# Install Go (if not already installed)
wget https://go.dev/dl/go1.22.0.linux-arm64.tar.gz
sudo tar -C /usr/local -xzf go1.22.0.linux-arm64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Clone the repository
git clone https://github.com/msollanych-tt/TRS80-M100-keyboard-driver.git
cd TRS80-M100-keyboard-driver

# Build the binary
go build -o m100keyboard ./cmd/m100keyboard

# Install
sudo cp m100keyboard /usr/local/bin/
sudo chmod +x /usr/local/bin/m100keyboard

# Install systemd service
sudo cp systemd/m100keyboard.service /etc/systemd/system/
```

## System Configuration

### 1. Enable uinput Kernel Module

The driver requires the `uinput` kernel module to create virtual input devices.

```bash
# Load the module immediately
sudo modprobe uinput

# Verify it loaded
lsmod | grep uinput

# Make it load on boot
echo "uinput" | sudo tee -a /etc/modules

# Verify entry was added
grep uinput /etc/modules
```

### 2. Configure Permissions (Optional)

You can either run the driver as root (using sudo or systemd) or add your user to the `gpio` group.

#### Option A: Run as root (simplest)
The systemd service runs as root by default - no additional configuration needed.

#### Option B: Add user to gpio group
```bash
# Add your user to the gpio group
sudo usermod -a -G gpio $USER

# Log out and back in for the change to take effect
# Or use: newgrp gpio

# Verify group membership
groups | grep gpio
```

**Note:** Even with gpio group membership, you'll still need root access for `/dev/uinput` unless you set up udev rules.

### 3. Configure udev Rules for /dev/uinput (Optional)

To allow non-root users to access `/dev/uinput`:

```bash
# Create udev rule
sudo tee /etc/udev/rules.d/99-uinput.rules > /dev/null <<EOF
KERNEL=="uinput", GROUP="input", MODE="0660"
EOF

# Reload udev rules
sudo udevadm control --reload-rules
sudo udevadm trigger

# Add user to input group
sudo usermod -a -G input $USER

# Log out and back in
```

## Systemd Service Setup

### Enable and Start Service

```bash
# Reload systemd daemon (if you manually installed the service file)
sudo systemctl daemon-reload

# Enable the service to start on boot
sudo systemctl enable m100keyboard

# Start the service now
sudo systemctl start m100keyboard

# Check service status
sudo systemctl status m100keyboard
```

Expected output:
```
● m100keyboard.service - TRS-80 Model 100 Keyboard Driver
     Loaded: loaded (/etc/systemd/system/m100keyboard.service; enabled; vendor preset: enabled)
     Active: active (running) since Sun 2024-02-02 10:30:00 UTC; 5s ago
       Docs: https://github.com/msollanych-tt/TRS80-M100-keyboard-driver
   Main PID: 1234 (m100keyboard)
      Tasks: 8 (limit: 4915)
        CPU: 15ms
     CGroup: /system.slice/m100keyboard.service
             └─1234 /usr/local/bin/m100keyboard

Feb 02 10:30:00 raspberrypi systemd[1]: Started TRS-80 Model 100 Keyboard Driver.
Feb 02 10:30:00 raspberrypi m100keyboard[1234]: level=INFO msg="TRS-80 Model 100 Keyboard Driver starting" version=1.0.0
Feb 02 10:30:00 raspberrypi m100keyboard[1234]: level=INFO msg="Created uinput keyboard device"
Feb 02 10:30:00 raspberrypi m100keyboard[1234]: level=INFO msg="Opened GPIO chip" chip=gpiochip0
Feb 02 10:30:00 raspberrypi m100keyboard[1234]: level=INFO msg="Starting keyboard scanner"
```

### Service Management Commands

```bash
# Start the service
sudo systemctl start m100keyboard

# Stop the service
sudo systemctl stop m100keyboard

# Restart the service
sudo systemctl restart m100keyboard

# Check status
sudo systemctl status m100keyboard

# Enable at boot
sudo systemctl enable m100keyboard

# Disable at boot
sudo systemctl disable m100keyboard

# View recent logs
sudo journalctl -u m100keyboard -n 50

# Follow logs in real-time
sudo journalctl -u m100keyboard -f

# View logs since boot
sudo journalctl -u m100keyboard -b
```

## Configuration

### Customizing Service Parameters

To run the driver with custom parameters (e.g., different debounce timing), edit the service file:

```bash
# Edit the service file
sudo systemctl edit --full m100keyboard

# Modify the ExecStart line:
ExecStart=/usr/local/bin/m100keyboard -debounce 15 -key-repeat 150 -debug

# Save and exit

# Reload systemd and restart
sudo systemctl daemon-reload
sudo systemctl restart m100keyboard
```

Available options:
- `-debounce int` - Debounce delay in milliseconds (default 10)
- `-scan-interval int` - Matrix scan interval in milliseconds (default 5)
- `-key-repeat int` - Minimum delay between repeated keypresses in milliseconds (default 100)
- `-post-key-delay int` - Delay after key emission in milliseconds (default 50)
- `-gpio-chip string` - GPIO chip device (default "gpiochip0")
- `-debug` - Enable debug logging

### Example Configurations

#### For keyboards with bouncy keys:
```
ExecStart=/usr/local/bin/m100keyboard -debounce 20 -key-repeat 150
```

#### For faster response time:
```
ExecStart=/usr/local/bin/m100keyboard -scan-interval 2 -debounce 5
```

#### For Raspberry Pi 5 (if gpiochip0 doesn't work):
```
ExecStart=/usr/local/bin/m100keyboard -gpio-chip gpiochip4
```

#### For debugging:
```
ExecStart=/usr/local/bin/m100keyboard -debug
```

## Testing the Installation

### Test 1: Manual Run

Before setting up the service, test the driver manually:

```bash
# Run with debug output
sudo m100keyboard -debug

# Press keys on the TRS-80 keyboard
# You should see debug messages like:
# level=DEBUG msg="Key pressed" row=0 col=0 key=44 shift=false ctrl=false alt=false

# Press Ctrl+C to stop
```

### Test 2: Verify Input Device

```bash
# List input devices
cat /proc/bus/input/devices

# Look for "TRS-80 M100 Keyboard" device
# Should show something like:
# N: Name="TRS-80 M100 Keyboard"
# P: Phys=
# S: Sysfs=/devices/virtual/input/input2
# U: Uniq=
# H: Handlers=sysrq kbd event2
```

### Test 3: Test Typing

```bash
# Open a text editor
nano test.txt

# Type using the TRS-80 keyboard
# Test:
# - All letters (a-z)
# - Shift+letters (A-Z)
# - Numbers (0-9)
# - Special characters
# - Ctrl combinations
# - Function keys
```

## Troubleshooting

### Service Won't Start

```bash
# Check for detailed error messages
sudo journalctl -u m100keyboard -n 50 --no-pager

# Common issues:
# - "failed to open GPIO chip": Wrong chip name, try -gpio-chip gpiochip4
# - "failed to create uinput device": Module not loaded, run: sudo modprobe uinput
# - "permission denied": Need sudo or proper group membership
```

### Keys Not Working

```bash
# Run in debug mode to see key detection
sudo m100keyboard -debug

# Check GPIO connections
# - Verify column pins are inputs with pull-down
# - Verify row pins are outputs
# - Check for loose connections
```

### High CPU Usage

```bash
# Check CPU usage
top -p $(pidof m100keyboard)

# If >20% when idle, increase scan interval:
sudo systemctl edit --full m100keyboard
# Change: ExecStart=/usr/local/bin/m100keyboard -scan-interval 10
sudo systemctl daemon-reload
sudo systemctl restart m100keyboard
```

### Permission Errors

```bash
# Error: "permission denied /dev/gpiochip0"
sudo usermod -a -G gpio $USER
# Log out and back in

# Error: "permission denied /dev/uinput"
# Either run as root (systemd does this by default)
# Or set up udev rules (see section above)
```

### Multiple Key Presses

```bash
# If single keypress produces multiple characters:
# Increase debounce and key-repeat delays:
sudo systemctl edit --full m100keyboard
# Change: ExecStart=/usr/local/bin/m100keyboard -debounce 20 -key-repeat 150
sudo systemctl daemon-reload
sudo systemctl restart m100keyboard
```

## Uninstallation

### If installed via .deb package:
```bash
sudo dpkg -r m100keyboard
```

### If installed manually:
```bash
# Stop and disable service
sudo systemctl stop m100keyboard
sudo systemctl disable m100keyboard

# Remove files
sudo rm /usr/local/bin/m100keyboard
sudo rm /etc/systemd/system/m100keyboard.service

# Reload systemd
sudo systemctl daemon-reload

# Optionally remove uinput from modules
sudo sed -i '/^uinput$/d' /etc/modules
```

## Advanced Configuration

### Running Multiple Instances

To run multiple keyboard drivers (e.g., different GPIO pins):

```bash
# Copy and rename service file
sudo cp /etc/systemd/system/m100keyboard.service /etc/systemd/system/m100keyboard2.service

# Edit the new service file
sudo systemctl edit --full m100keyboard2
# Modify Description and ExecStart with different GPIO chip or pins

# Enable and start
sudo systemctl enable m100keyboard2
sudo systemctl start m100keyboard2
```

### Custom GPIO Pin Mapping

The current implementation has fixed pin mappings. To change pins, you need to modify the source code in `internal/keyboard/matrix.go` and rebuild.

### Logging Configuration

To change log verbosity without debug mode:

```bash
# Reduce logging (only errors)
sudo systemctl edit --full m100keyboard
# Add: StandardOutput=null
# Keep: StandardError=journal

# Increase log retention
sudo journalctl --vacuum-time=30d
```

## Getting Help

If you encounter issues:

1. Check the logs: `sudo journalctl -u m100keyboard -f`
2. Run with debug mode: `sudo m100keyboard -debug`
3. Verify hardware connections
4. Check [Troubleshooting section in README](README.md#troubleshooting)
5. Open an issue on GitHub with:
   - Raspberry Pi model and OS version
   - Full error messages from logs
   - Output of `sudo m100keyboard -debug`
   - GPIO chip device (`ls /dev/gpiochip*`)
