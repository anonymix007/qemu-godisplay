package qemu

import (
    "github.com/godbus/dbus/v5"
)

type Keyboard struct {
    conn     *dbus.Conn
    keyboard dbus.BusObject
}

type KeyboardModifier int

const (
    Scroll KeyboardModifier = 1 << 0
    Num    KeyboardModifier = 1 << 1
    Caps   KeyboardModifier = 1 << 2
)

func newKeyboard(conn *dbus.Conn, keyboard dbus.BusObject) (*Keyboard, error) {
    return &Keyboard{conn, keyboard}, nil
}

func (k *Keyboard) GetModifiers() KeyboardModifier {
    mod, err := getProp(k.keyboard, keyboardModifiers)
    if err != nil {
        mod = 0
    }
    return KeyboardModifier(mod.(int))
}

func (k *Keyboard) Press(keycode uint32) {
    k.keyboard.Call(keyboardPress, 0, keycode)
}

func (k *Keyboard) Release(keycode uint32) {
    k.keyboard.Call(keyboardRelease, 0, keycode)
}
