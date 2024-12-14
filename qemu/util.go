package qemu

import (
    "net"
    "os"
    "syscall"

    "github.com/godbus/dbus/v5"
)

func getProp(obj dbus.BusObject, prop string) (interface{}, error) {
    ret, err := obj.GetProperty(prop)
    if err != nil {
        return nil, err
    }
    return ret.Value(), nil
}

func fdToUnixConn(fd dbus.UnixFD, name string) (*net.UnixConn, error) {
    c, err := net.FileConn(os.NewFile(uintptr(fd), name))
    if err != nil {
        return nil, err
    }
    return c.(*net.UnixConn), nil
}

func socketpair() (dbus.UnixFD, dbus.UnixFD, error) {
    fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM|syscall.SOCK_NONBLOCK|syscall.SOCK_CLOEXEC, 0)
    if err != nil {
        return -1, -1, err
    }

    return dbus.UnixFD(fds[0]), dbus.UnixFD(fds[1]), nil
}
