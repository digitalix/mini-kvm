#!/bin/bash

GADGET_NAME="hid_devices"
GADGET_PATH="/sys/kernel/config/usb_gadget/$GADGET_NAME"

if [ ! -d "$GADGET_PATH" ]; then
    echo "Gadget $GADGET_NAME not found"
    exit 1
fi

echo "Disabling USB gadget..."

# Unbind from UDC
echo "" > "$GADGET_PATH/UDC"

# Remove symlinks
rm -f "$GADGET_PATH/configs/c.1/hid.usb0"
rm -f "$GADGET_PATH/configs/c.1/hid.usb1"

# Remove functions
rmdir "$GADGET_PATH/functions/hid.usb0" 2>/dev/null
rmdir "$GADGET_PATH/functions/hid.usb1" 2>/dev/null

# Remove strings
rmdir "$GADGET_PATH/configs/c.1/strings/0x409" 2>/dev/null
rmdir "$GADGET_PATH/configs/c.1" 2>/dev/null
rmdir "$GADGET_PATH/strings/0x409" 2>/dev/null

# Remove gadget directory
rmdir "$GADGET_PATH" 2>/dev/null

echo "USB gadget disabled"