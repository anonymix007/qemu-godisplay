package qemu

import (
    "fmt"
    "os"

    "github.com/godbus/dbus/v5"
)

type VM struct {
    conn       *dbus.Conn
    vm         dbus.BusObject
    name       string
    uuid       string
    consoleIDs []uint32
    interfaces []string
}

func NewVM(path ...string) (*VM, error) {
    if len(path) == 1 {
        os.Setenv("DBUS_SESSION_BUS_ADDRESS", path[0])
    } else if len(path) != 0 {
        return nil, fmt.Errorf("wrong number of arguments: %d, expected 0 or 1", len(path))
    }

    conn, err := dbus.SessionBus()

    if err != nil {
        return nil, err
    }

    conn.EnableUnixFDs()
    if !conn.SupportsUnixFDs() {
        return nil, fmt.Errorf("fds are not supported")
    }

    vm := conn.Object(qemuIntf, vmPath)

    name, err := getProp(vm, vmName)
    if err != nil {
        return nil, err
    }

    uuid, err := getProp(vm, vmUUID)
    if err != nil {
        return nil, err
    }

    cons, err := getProp(vm, vmConsoleIDs)
    if err != nil {
        return nil, err
    }

    intf, err := getProp(vm, vmInterfaces)
    if err != nil {
        return nil, err
    }

    return &VM{conn, vm, name.(string), uuid.(string), cons.([]uint32), intf.([]string)}, nil
}

func (vm *VM) Name() string {
    return vm.name
}

func (vm *VM) UUID() string {
    return vm.uuid
}

func (vm *VM) NumConsoles() int {
    return len(vm.consoleIDs)
}

func (vm *VM) GetConsole(n int) (*Console, error) {
    if n >= len(vm.consoleIDs) {
        return nil, fmt.Errorf("console %d does not exist, max is %d", n, len(vm.consoleIDs)-1)
    }

    return newConsole(vm.conn, vm.consoleIDs[n])
}

func (vm *VM) Close() {
    vm.conn.Close()
}
