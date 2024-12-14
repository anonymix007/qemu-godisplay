package qemu

import (
    "github.com/godbus/dbus/v5"
)

type Mouse struct {
    conn  *dbus.Conn
    mouse dbus.BusObject
    isAbs bool
}

func newMouse(conn *dbus.Conn, mouse dbus.BusObject) (*Mouse, error) {
    isAbs, err := getProp(mouse, mouseIsAbs)
    if err != nil {
        return nil, err
    }

    return &Mouse{conn, mouse, isAbs.(bool)}, nil
}

func (m *Mouse) IsAbsolute() bool {
    return m.isAbs
}

func (m *Mouse) SetAbsPosition(x, y uint32) {
    m.mouse.Call(mouseSetAbsPosition, 0, x, y)
}

func (m *Mouse) Press(button uint32) {
    m.mouse.Call(mousePress, 0, button)
}

func (m *Mouse) Release(button uint32) {
    m.mouse.Call(mouseRelease, 0, button)
}
