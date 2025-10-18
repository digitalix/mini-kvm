#!/bin/bash

GADGET_DIR="/sys/kernel/config/usb_gadget"
GADGET_NAME="hid_devices"
GADGET_PATH="$GADGET_DIR/$GADGET_NAME"

# Remove existing gadget if it exists
if [ -d "$GADGET_PATH" ]; then
    echo "" > /sys/kernel/config/usb_gadget/$GADGET_NAME/UDC
    rm -rf "$GADGET_PATH"
fi

mkdir -p "$GADGET_PATH"
cd "$GADGET_PATH"

echo 0x1d6b > idVendor
echo 0x0104 > idProduct
echo 0x0100 > bcdDevice
echo 0x0200 > bcdUSB

mkdir -p strings/0x409
echo "fedcba9876543210" > strings/0x409/serialnumber
echo "Radxa" > strings/0x409/manufacturer
echo "Virtual HID" > strings/0x409/product

mkdir -p configs/c.1
mkdir -p configs/c.1/strings/0x409
echo "Config 1" > configs/c.1/strings/0x409/configuration
echo 250 > configs/c.1/MaxPower

# Keyboard function
mkdir -p functions/hid.usb0
echo 1 > functions/hid.usb0/protocol
echo 1 > functions/hid.usb0/subclass
echo 8 > functions/hid.usb0/report_length
echo -ne \\x05\\x01\\x09\\x06\\xa1\\x01\\x05\\x07\\x19\\xe0\\x29\\xe7\\x15\\x00\\x25\\x01\\x75\\x01\\x95\\x08\\x81\\x02\\x95\\x01\\x75\\x08\\x81\\x03\\x95\\x05\\x75\\x01\\x05\\x08\\x19\\x01\\x29\\x05\\x91\\x02\\x95\\x01\\x75\\x03\\x91\\x03\\x95\\x06\\x75\\x08\\x15\\x00\\x25\\x65\\x05\\x07\\x19\\x00\\x29\\x65\\x81\\x00\\xc0 > functions/hid.usb0/report_desc

# Absolute pointer (digitizer/touchscreen) function
mkdir -p functions/hid.usb1
echo 0 > functions/hid.usb1/protocol
echo 0 > functions/hid.usb1/subclass
echo 7 > functions/hid.usb1/report_length

echo -ne \\x05\\x0d\\x09\\x04\\xa1\\x01\\x85\\x01\\x05\\x09\\x19\\x01\\x29\\x03\\x15\\x00\\x25\\x01\\x75\\x01\\x95\\x03\\x81\\x02\\x95\\x05\\x81\\x03\\x05\\x01\\x09\\x30\\x09\\x31\\x16\\x00\\x00\\x26\\xff\\x7f\\x36\\x00\\x00\\x46\\xff\\x7f\\x66\\x00\\x00\\x75\\x10\\x95\\x02\\x81\\x02\\x09\\x38\\x15\\x81\\x25\\x7f\\x75\\x08\\x95\\x01\\x81\\x06\\xc0 > functions/hid.usb1/report_desc
ln -s functions/hid.usb0 configs/c.1/
ln -s functions/hid.usb1 configs/c.1/

UDC=$(ls /sys/class/udc | head -n1)
echo "$UDC" > UDC

chown :vkeyboard /dev/hidg0
chown :vmouse /dev/hidg1

chmod 660 /dev/hidg0
chmod 660 /dev/hidg1

echo "USB HID gadget created successfully"
echo "Keyboard: /dev/hidg0"
echo "Touchscreen: /dev/hidg1"