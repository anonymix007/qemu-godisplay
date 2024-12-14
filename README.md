# Simple QEMU D-Bus Display Client in Go

## Setup

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
### Manually
- TODO

## Quick Start
```console
go build -o qemu-godisplay viewer/viewer.go
sudo -u libvirt-qemu DBUS_SESSION_BUS_ADDRESS=unix:path=/run/libvirt/qemu/dbus/34-win11-dbus.sock ./qemu-godisplay
```
