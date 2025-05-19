package qemu

import (
    "fmt"

    "github.com/godbus/dbus/v5"
    "github.com/godbus/dbus/v5/prop"
)

type DisplayListener interface {
    Scanout(width, height, stride, format uint32, data []byte) *dbus.Error
    Update(x, y, width, height int32, stride, format uint32, data []byte) *dbus.Error
    ScanoutDMABUF(fd dbus.UnixFD, width, height, stride, fourcc uint32, modifier uint64, y0_top bool) *dbus.Error
    UpdateDMABUF(x, y, width, height int32) *dbus.Error
    Disable() *dbus.Error
    MouseSet(x, y, on int) *dbus.Error
    CursorDefine(width, height, hot_x, hot_y int, data []byte) *dbus.Error
}

type DisplayListenerUnixMap interface {
    ScanoutMap(fd dbus.UnixFD, offset, width, height, stride, format uint32) *dbus.Error
    UpdateMap(x, y, width, height int32) *dbus.Error
}

type DisplayListenerUnixScanoutDMABUF2 interface {
    ScanoutDMABUF2(fd []dbus.UnixFD, x, y, width, height uint32, offset, stride []uint32, num_planes, fourcc, backing_width, backing_height uint32, modifier uint64, y0_top bool) *dbus.Error
}

type Console struct {
    conn        *dbus.Conn
    console     dbus.BusObject
    label       string
    consoleType string
    width       uint32
    height      uint32
    interfaces  []string
    listeners   []listenerConn
}

type listenerConn struct {
    conn *dbus.Conn
    impl DisplayListener
    prop *prop.Properties
}

func newConsole(conn *dbus.Conn, n uint32) (*Console, error) {
    consolePath := dbus.ObjectPath(fmt.Sprintf(consolePath, n))
    console := conn.Object(qemuIntf, consolePath)

    label, err := getProp(console, consoleLabel)
    if err != nil {
        return nil, err
    }

    ctype, err := getProp(console, consoleType)
    if err != nil {
        return nil, err
    }

    width, err := getProp(console, consoleWidth)
    if err != nil {
        return nil, err
    }

    height, err := getProp(console, consoleHeight)
    if err != nil {
        return nil, err
    }

    intf, err := getProp(console, consoleInterfaces)
    if err != nil {
        return nil, err
    }

    return &Console{conn, console, label.(string), ctype.(string), width.(uint32), height.(uint32), intf.([]string), nil}, nil
}

func (c *Console) Label() string {
    return c.label
}

func (c *Console) Type() string {
    return c.consoleType
}

func (c *Console) Width() uint32 {
    return c.width
}

func (c *Console) Height() uint32 {
    return c.height
}

func (c *Console) GetMouse() (*Mouse, error) {
    return newMouse(c.conn, c.console)
}

func (c *Console) GetKeyboard() (*Keyboard, error) {
    return newKeyboard(c.conn, c.console)
}

func (c *Console) RegisterListener(listener DisplayListener) error {
    usFd, themFd, err := socketpair()
    if err != nil {
        return err
    }

    us, err := fdToUnixConn(usFd, "us")
    if err != nil {
        return err
    }

    conn, err := dbus.DialUnix(us)
    if err != nil {
        return err
    }

    err = conn.Export(listener, listenerPath, listenerIntf)
    if err != nil {
        return err
    }

    interfaces := []string{listenerIntf}

    unixListener, ok := listener.(DisplayListenerUnixMap)
    if ok {
        err = conn.Export(unixListener, listenerPath, listenerUnixMapIntf)
        if err != nil {
            return err
        }
        interfaces = append(interfaces, listenerUnixMapIntf)
    }

    unixDmaBuf2Listener, ok := listener.(DisplayListenerUnixScanoutDMABUF2)
    if ok {
        err = conn.Export(unixDmaBuf2Listener, listenerPath, listenerUnixScanoutDMABUF2Intf)
        if err != nil {
            return err
        }
        interfaces = append(interfaces, listenerUnixScanoutDMABUF2Intf)
    }

    fmt.Println("Registering interfaces:", interfaces)

    propsMap := map[string]map[string]*prop.Prop{
        listenerIntf: {
            "Interfaces": {
                Value:    interfaces,
                Writable: false,
                Emit:     prop.EmitConst,
                Callback: nil,
            },
        },
    }

    props, err := prop.Export(conn, listenerPath, propsMap)
    if err != nil {
        return err
    }

    c.listeners = append(c.listeners, listenerConn{conn, listener, props})

    ret := c.console.Call(consoleRegisterListener, 0, themFd)

    if err = conn.Auth(nil); err != nil {
        return err
    }

    // FIXME: hangs here. Is this expected?
    // <- ret.Done

    return ret.Err
}
