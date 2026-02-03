#!/bin/sh
set -e

# Stop and disable the service if it's running
if systemctl is-active --quiet m100keyboard; then
    systemctl stop m100keyboard || true
fi

if systemctl is-enabled --quiet m100keyboard; then
    systemctl disable m100keyboard || true
fi

echo "m100keyboard service stopped and disabled"
