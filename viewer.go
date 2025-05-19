package main

import (
    "fmt"
    "os"

    "github.com/go-gst/go-glib/glib"
    "github.com/go-gst/go-gst/gst"
    "github.com/go-gst/go-gst/gst/app"
    "github.com/go-gst/go-gst/gst/video"
    "github.com/godbus/dbus/v5"

    "qemu"
)

type DisplayListener struct {
    src  *app.Source
    flip *gst.Element

    caps *gst.Caps
    img  Picture
}


func (dl *DisplayListener) Scanout(width, height, stride, format uint32, data []byte) *dbus.Error {
    // fmt.Printf("Scanout: resolution %dx%d, stride %d, fmt %x, data %d\n", width, height, stride, format, len(data))

    dl.flip.SetArg("input-flags-override", "none")
    dl.flip.SetArg("video-direction", "identity")

    dl.img = NewRawPicture(width, height, stride, format, data)
    dl.caps = dl.img.CreateCaps()

    sample := gst.NewSample(dl.img.CreateBuffer(), dl.caps)
    dl.src.PushSample(sample)

    return nil
}

func (dl *DisplayListener) Update(x, y, width, height int32, stride, format uint32, data []byte) *dbus.Error {
    // fmt.Printf("Update: rect pos (%d,%d) size %dx%d, stride %d, fmt %x, data %d\n", x, y, width, height, stride, format, len(data))
    if dl.img == nil {
        fmt.Println("Update before Scanout?")
        return nil
    }

    dl.img.Update(uint32(x), uint32(y), uint32(width), uint32(height), stride, data)
    sample := gst.NewSample(dl.img.CreateBuffer(), dl.caps)
    dl.src.PushSample(sample)

    return nil
}

func (dl *DisplayListener) ScanoutDMABUF(fd dbus.UnixFD, width, height, stride, fourcc uint32, modifier uint64, y0_top bool) *dbus.Error {
    // fmt.Printf("ScanoutDMABUF: resolution %dx%d, stride %d, fmt %x:%x, fd %d, normal y: %t\n", width, height, stride, fourcc, modifier, fd, y0_top)

    if y0_top {
        dl.flip.SetArg("input-flags-override", "left-flipped")
        dl.flip.SetArg("video-direction", "vert")
    } else {
        dl.flip.SetArg("input-flags-override", "none")
        dl.flip.SetArg("video-direction", "identity")
    }

    dl.img = NewDmaPicture(int(fd), width, height, []uint32{0}, []uint32{stride}, fourcc, modifier, y0_top)
    dl.caps = dl.img.CreateCaps()

    sample := gst.NewSample(dl.img.CreateBuffer(), dl.caps)
    dl.src.PushSample(sample)

    return nil
}

func (dl *DisplayListener) UpdateDMABUF(x, y, width, height int32) *dbus.Error {
    // fmt.Printf("UpdateDMABUF: rect pos (%d,%d) size %dx%d\n", x, y, width, height)
    if dl.img == nil {
        fmt.Println("UpdateDMABUF before ScanoutDMABUF?")
        return nil
    }

    dl.img.Update(uint32(x), uint32(y), uint32(width), uint32(height), 0, nil)
    sample := gst.NewSample(dl.img.CreateBuffer(), dl.caps)
    dl.src.PushSample(sample)

    return nil
}

func (dl *DisplayListener) Disable() *dbus.Error {
    fmt.Printf("Disable\n")
    return nil
}

func (dl *DisplayListener) MouseSet(x, y, on int) *dbus.Error {
    fmt.Printf("MouseSet: %d,%d -> %d\n", x, y, on)
    return nil
}

func (dl *DisplayListener) CursorDefine(width, height, hot_x, hot_y int, data []byte) *dbus.Error {
    fmt.Printf("CursorDefine: %dx%d (%d,%d) -> %v\n", width, height, hot_x, hot_y, data)
    return nil
}

func (dl *DisplayListener) ScanoutMap(fd dbus.UnixFD, offset, width, height, stride, format uint32) *dbus.Error {
    // fmt.Printf("ScanoutMap: resolution %dx%d, stride %d, fmt %x, fd %d, offset %d\n", width, height, stride, format, fd, offset)

    dl.flip.SetArg("input-flags-override", "none")
    dl.flip.SetArg("video-direction", "identity")

    dl.img = NewShmemPicture(int(fd), offset, width, height, stride, format)
    dl.caps = dl.img.CreateCaps()

    sample := gst.NewSample(dl.img.CreateBuffer(), dl.caps)
    dl.src.PushSample(sample)

    return nil
}

func (dl *DisplayListener) UpdateMap(x, y, width, height int32) *dbus.Error {
    // fmt.Printf("UpdateMap: rect pos (%d,%d) size %dx%d\n", x, y, width, height)
    if dl.img == nil {
        fmt.Println("UpdateMap before ScanoutMap?")
        return nil
    }

    dl.img.Update(uint32(x), uint32(y), uint32(width), uint32(height), 0, nil)
    sample := gst.NewSample(dl.img.CreateBuffer(), dl.caps)
    dl.src.PushSample(sample)

    return nil
}

func (dl *DisplayListener) ScanoutDMABUF2(fd []dbus.UnixFD, x, y, width, height uint32, offset, stride []uint32, num_planes, fourcc, backing_width, backing_height uint32, modifier uint64, y0_top bool) *dbus.Error {
    // fmt.Printf("ScanoutDMABUF2: resolution %dx%d (%dx%d), num_planes %d, offset %v, stride %v, fmt %x:%x, fd %d, normal y: %t\n", width, height, backing_width, backing_height, num_planes, offset, stride, fourcc, modifier, fd, y0_top)

    if y0_top {
        dl.flip.SetArg("input-flags-override", "left-flipped")
        dl.flip.SetArg("video-direction", "vert")
    } else {
        dl.flip.SetArg("input-flags-override", "none")
        dl.flip.SetArg("video-direction", "identity")
    }

    if len(fd) != 1 {
        panic(fmt.Errorf("cannot handle dma buffers with %d fds", len(fd)))
    }

    dl.img = NewDmaPicture(int(fd[0]), width, height, offset, stride, fourcc, modifier, y0_top)
    dl.caps = dl.img.CreateCaps()

    sample := gst.NewSample(dl.img.CreateBuffer(), dl.caps)
    dl.src.PushSample(sample)

    return nil
}

func main() {
    vm, err := qemu.NewVM()
    if err != nil {
        panic(err)
    }
    defer vm.Close()

    fmt.Printf("Session ADDR: %s\n", os.Getenv("DBUS_SESSION_BUS_ADDRESS"))
    fmt.Println("Connected to a VM:")

    fmt.Printf("  Name: %v\n", vm.Name())
    fmt.Printf("  UUID: %v\n", vm.UUID())
    fmt.Printf("  Number of consoles: %d\n", vm.NumConsoles())

    console, err := vm.GetConsole(0)
    if err != nil {
        panic(err)
    }

    fmt.Println("Connected to a console 0:")
    fmt.Printf("  %s display \"%s\": %dx%d\n", console.Type(), console.Label(), console.Width(), console.Height())

    mouse, err := console.GetMouse()
    if err != nil {
        panic(err)
    }

    keyboard, err := console.GetKeyboard()
    if err != nil {
        panic(err)
    }

    fmt.Printf("  Mouse is absolute: %t\n", mouse.IsAbsolute())

    gst.Init(nil)

    mainLoop := glib.NewMainLoop(glib.MainContextDefault(), false)

    pipeline, err := gst.NewPipelineFromString("appsrc format=time do-timestamp=true stream-type=stream is-live=true name=src ! glupload ! glcolorconvert ! glviewconvert input-mode-override=left name=flip ! glimagesink")

    if err != nil {
        panic(err)
    }

    elem, err := pipeline.GetElementByName("src")
    if err != nil {
        panic(err)
    }

    src := app.SrcFromElement(elem)

    flip, err := pipeline.GetElementByName("flip")
    if err != nil {
        panic(err)
    }

    listener := &DisplayListener{src, flip, nil, nil}

    err = console.RegisterListener(listener)
    if err != nil {
        panic(err)
    }

    pipeline.GetPipelineBus().AddWatch(func(msg *gst.Message) bool {
        switch msg.Type() {
        case gst.MessageEOS:
            pipeline.BlockSetState(gst.StateNull)
            mainLoop.Quit()
        case gst.MessageError:
            err := msg.ParseError()
            fmt.Println("ERROR:", err.Error())
            if debug := err.DebugString(); debug != "" {
                fmt.Println("DEBUG:", debug)
            }
            mainLoop.Quit()
        case gst.MessageElement:
            if nav := video.ToNavigationMessage(msg); nav != nil {
                if nav.GetType() == video.NavigationMessageEvent {
                    event := nav.GetEvent()

                    switch event.GetType() {
                    case video.NavigationEventInvalid:
                        fmt.Println("Invalid navigation event:", event)
                    case video.NavigationEventKeyPress:
                        key, ok := event.ParseKeyEvent()
                        if ok {
                            keycode, ok := keymap[key]
                            if ok {
                                keyboard.Press(keycode)
                            } else {
                                fmt.Println("Unknown key down:", key)
                            }
                        } else {
                            panic("wtf")
                        }
                    case video.NavigationEventKeyRelease:
                        key, ok := event.ParseKeyEvent()
                        if ok {
                            keycode, ok := keymap[key]
                            if ok {
                                keyboard.Release(keycode)
                            } else {
                                fmt.Println("Unknown key down:", key)
                            }
                        } else {
                            panic("wtf")
                        }
                    case video.NavigationEventMouseButtonPress:
                        button, x, y, ok := event.ParseMouseButtonEvent()
                        if ok {
                            _ = x
                            _ = y
                            //fmt.Println("Mouse down:", button, int(x), int(y))
                            mouse.Press(uint32(button - 1))
                        } else {
                            panic("wtf")
                        }
                    case video.NavigationEventMouseButtonRelease:
                        button, x, y, ok := event.ParseMouseButtonEvent()

                        if ok {
                            _ = x
                            _ = y
                            //fmt.Println("Mouse up:", button, int(x), int(y))
                            mouse.Release(uint32(button - 1))
                        } else {
                            panic("wtf")
                        }
                    case video.NavigationEventMouseMove:
                        x, y, ok := event.ParseMouseMoveEvent()
                        if ok {
                            //fmt.Println("Mouse move:", int(x), int(y))
                            mouse.SetAbsPosition(uint32(x), uint32(y))
                        } else {
                            panic("wtf")
                        }
                    case video.NavigationEventCommand:
                        cmd, ok := event.ParseCommandEvent()
                        if ok {
                            fmt.Println("Command:", cmd)
                        } else {
                            panic("wtf")
                        }
                    case video.NavigationEventMouseScroll:
                        x, y, dx, dy, ok := event.ParseMouseScrollEvent()
                        if ok {
                            fmt.Println("Mouse scroll:", x, y, dx, dy)
                        } else {
                            panic("wtf")
                        }
                    }
                } else {
                    fmt.Println("Unhandled navigation:", nav.GetType())
                }
            } else {
                // Unknown message
                //fmt.Println("Unknown element message:", msg)
            }
        default:
            //fmt.Println("Unknown message:", msg)
        }
        return true
    })

    pipeline.SetState(gst.StatePlaying)

    mainLoop.Run()
}
