#!/bin/sh
set -e

# Enable uinput module
if ! lsmod | grep -q uinput; then
    modprobe uinput || true
fi

# Ensure uinput loads on boot
if ! grep -q "^uinput" /etc/modules 2>/dev/null; then
    echo "uinput" >> /etc/modules || true
fi

# Reload systemd daemon
systemctl daemon-reload || true

echo "m100keyboard installed successfully"
echo "To enable at boot: sudo systemctl enable m100keyboard"
echo "To start now: sudo systemctl start m100keyboard"
