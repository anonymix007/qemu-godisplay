# Simple QEMU D-Bus Display Client in Go



## Quick Start

### With `virt-manager`
- Run `virt-manager`
- Open a VM configuration window
- Replace `Display` type to `dbus`
- Turn on the VM and get socket path from XML tab:
```XML
<graphics type="dbus" address="unix:path=/run/libvirt/qemu/dbus/34-win11-dbus.sock">
  <gl enable="yes" rendernode="/dev/dri/by-path/pci-0000:10:00.0-render"/>
</graphics>
```
- Run:
```
go build . && sudo -u libvirt-qemu DBUS_SESSION_BUS_ADDRESS=unix:path=/run/libvirt/qemu/dbus/34-win11-dbus.sock ./qemu-godisplay
```

### Manually
- Launch QEMU with D-Bus display:
```
qemu-system-x86_64 \
    -drive if=pflash,format=raw,unit=0,file=/usr/share/edk2-ovmf/x64/OVMF_CODE.4m.fd,readonly=on \
    -drive if=pflash,format=raw,unit=1,file=./efi_VARS.fd \
    -enable-kvm \
    -cpu host \
    -usb \
    -device virtio-vga,xres=1280,yres=800 \
    -device usb-mouse \
    -device usb-tablet \
    -device usb-kbd \
    -m 16G \
    -smp 8 \
    -display dbus \
    -hda drive.qcow2
```
- Run:
```
go run .
```
