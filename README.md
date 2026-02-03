# TRS-80 Model 100 Keyboard Driver

A keyboard driver that enables the TRS-80 Model 100 keyboard to work as a USB HID device on Raspberry Pi.

## Overview

This project allows you to use a vintage TRS-80 Model 100 keyboard with modern computers by interfacing it with a Raspberry Pi. The driver reads the keyboard matrix using GPIO pins and translates keypresses into standard Linux input events via uinput.

## Features

- Full keyboard matrix scanning (8x9 matrix)
- Support for modifier keys (Shift, Control, Alt, Code)
- Software debouncing and key repeat handling
- Systemd service for automatic startup
- Debug logging mode for troubleshooting

## Hardware Requirements

- Raspberry Pi (tested on RPi 5)
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

The driver includes several configurable parameters in the Python script:

- `debounce`: Delay between GPIO checks (default: 0.01s)
- `delay`: Delay after key emission (default: 0.05s)
- `keyrepeat`: Minimum time between repeated keypresses (default: 0.1s)

## License

See LICENSE file for details.

## Credits

- Original code: Codeman
- Enhancements: Mike Sollanych
