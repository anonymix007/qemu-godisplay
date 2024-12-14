package main

import (
    "encoding/binary"
    "fmt"
    "os"

    "github.com/go-gst/go-glib/glib"
    "github.com/go-gst/go-gst/gst"
    "github.com/go-gst/go-gst/gst/allocators"
    "github.com/go-gst/go-gst/gst/app"
    "github.com/go-gst/go-gst/gst/video"
    "github.com/godbus/dbus/v5"

    "qemu"
)

var formats = map[int]string{
    0x20028888: "BGRA",  // PIXMAN_a8r8g8b8
    0x20020888: "BGRx",  // PIXMAN_x8r8g8b8
    0x20038888: "RGBA",  // PIXMAN_a8b8g8r8
    0x20030888: "RGBx",  // PIXMAN_x8b8g8r8
    0x20088888: "ARGB",  // PIXMAN_b8g8r8a8
    0x20080888: "xRGB",  // PIXMAN_b8g8r8x8
    0x20098888: "ABGR",  // PIXMAN_r8g8b8a8
    0x20090888: "xBGR",  // PIXMAN_r8g8b8x8
    0x18020888: "BGR",   // PIXMAN_r8g8b8
    0x18030888: "RGB",   // PIXMAN_b8g8r8
    0x10020565: "BGR16", // PIXMAN_r5g6b5
    0x10021555: "BGR15", // PIXMAN_a1r5g5b5
    0x10020555: "BGR15", // PIXMAN_x1r5g5b5
}

type Picture interface {
    CreateCaps() *gst.Caps
    CreateBuffer() *gst.Buffer
    Update(x, y, width, height, stride uint32, data []byte)
}

type RawPicture struct {
    data []byte
    w    uint32
    h    uint32

    stride uint32

    fmt uint32
    bpp uint32
}

func (p *RawPicture) CreateCaps() *gst.Caps {
    return gst.NewCapsFromString(fmt.Sprintf("video/x-raw,format=%s,width=%d,height=%d", formats[int(p.fmt)], p.w, p.h))
}

func (p *RawPicture) CreateBuffer() *gst.Buffer {
    return gst.NewBufferFromBytes(p.data)
}

func (p *RawPicture) Update(x, y, width, height, stride uint32, data []byte) {
    for i := range height {
        copy(p.data[p.stride*(y+i)+x*p.bpp:],
            data[stride*i:stride*(i+1)])
    }
}

type DmaPicture struct {
    fd  int
    w   uint32
    h   uint32

    stride uint32
    size   int64

    fourcc uint32
    mod    uint64

    alloc *allocators.DmaBufAllocator
}

func (p *DmaPicture) CreateCaps() *gst.Caps {
    bytes := make([]byte, 4)
    binary.LittleEndian.PutUint32(bytes, p.fourcc)
    fourccStr := string(bytes)

    // FIXME: gstreamer does not support 0xffffffffffffff
    // This is most likely incorrect solution, but I don't have a better one yet
    if p.mod == 0 || p.mod == 0xffffffffffffff {
        return gst.NewCapsFromString(fmt.Sprintf("video/x-raw(memory:DMABuf),format=DMA_DRM,width=%d,height=%d,drm-format=%s", p.w, p.h, fourccStr))
    } else {
        return gst.NewCapsFromString(fmt.Sprintf("video/x-raw(memory:DMABuf),format=DMA_DRM,width=%d,height=%d,drm-format=%s:0x%x", p.w, p.h, fourccStr, p.mod))
    }
}

func (p *DmaPicture) CreateBuffer() *gst.Buffer {
    buffer := gst.NewEmptyBuffer()
    memory := p.alloc.AllocDmaBufWithFlags(p.fd, p.size, allocators.FdMemoryFlagDontClose)
    buffer.AppendMemory(memory)
    return buffer
}

func (p *DmaPicture) Update(x, y, width, height, stride uint32, data []byte) {
    // do nothing
}

type ShmemPicture struct {
    fd  int
    w   uint32
    h   uint32

    size   int64
    offset uint32
    stride uint32
    fmt    uint32

    alloc *allocators.FdAllocator
}

func (p *ShmemPicture) CreateCaps() *gst.Caps {
    return gst.NewCapsFromString(fmt.Sprintf("video/x-raw,format=%s,width=%d,height=%d", formats[int(p.fmt)], p.w, p.h))
}

func (p *ShmemPicture) CreateBuffer() *gst.Buffer {
    buffer := gst.NewEmptyBuffer()
    memory := p.alloc.AllocFd(p.fd, p.size, allocators.FdMemoryFlagDontClose)
    buffer.AppendMemory(memory)
    return buffer
}

func (p *ShmemPicture) Update(x, y, width, height, stride uint32, data []byte) {
    // do nothing
}

type DisplayListener struct {
    src *app.Source

    caps *gst.Caps
    img  Picture
}

func (dl *DisplayListener) Scanout(width, height, stride, format uint32, data []byte) *dbus.Error {
    fmt.Printf("Scanout: resolution %dx%d, stride %d, fmt %d, data %d\n", width, height, stride, format, len(data))
    dl.img = &RawPicture{data, width, height, stride, format, stride / width}
    dl.caps = dl.img.CreateCaps()

    sample := gst.NewSample(dl.img.CreateBuffer(), dl.caps)
    dl.src.PushSample(sample)

    return nil
}

func (dl *DisplayListener) Update(x, y, width, height int32, stride, format uint32, data []byte) *dbus.Error {
    fmt.Printf("Update: rect pos (%d,%d) size %dx%d, stride %d, fmt %d, data %d\n", x, y, width, height, stride, format, len(data))
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
    fmt.Printf("ScanoutDMABUF: resolution %dx%d, stride %d, fmt %x:%x, fd %d, normal y: %t\n", width, height, stride, fourcc, modifier, fd, y0_top)
    // TODO: Handle y0_top
    dl.img = &DmaPicture{int(fd), width, height, stride, int64(stride) * int64(height), fourcc, modifier, allocators.NewDmaBufAllocator()}
    dl.caps = dl.img.CreateCaps()

    return nil
}

func (dl *DisplayListener) UpdateDMABUF(x, y, width, height int32) *dbus.Error {
    fmt.Printf("UpdateDMABUF: rect pos (%d,%d) size %dx%d\n", x, y, width, height)
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
    fmt.Printf("ScanoutMap: resolution %dx%d, stride %d, fmt %x, fd %d, offset %d\n", width, height, stride, format, fd, offset)
    dl.img = &ShmemPicture{int(fd), width, height, int64(stride) * int64(height), offset, stride, format, allocators.NewFdAllocator()}
    dl.caps = dl.img.CreateCaps()

    sample := gst.NewSample(dl.img.CreateBuffer(), dl.caps)
    dl.src.PushSample(sample)

    return nil
}

func (dl *DisplayListener) UpdateMap(x, y, width, height int32) *dbus.Error {
    fmt.Printf("UpdateMap: rect pos (%d,%d) size %dx%d\n", x, y, width, height)

    dl.img.Update(uint32(x), uint32(y), uint32(width), uint32(height), 0, nil)
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

    fmt.Printf("  Mouse is absolute: %t\n", mouse.IsAbsolute())

    gst.Init(nil)

    mainLoop := glib.NewMainLoop(glib.MainContextDefault(), false)

    pipeline, err := gst.NewPipelineFromString("appsrc format=time do-timestamp=true stream-type=stream is-live=true name=src ! videoconvert ! glimagesink")
    if err != nil {
        panic(err)
    }

    elem, err := pipeline.GetElementByName("src")
    if err != nil {
        panic(err)
    }

    src := app.SrcFromElement(elem)

    listener := &DisplayListener{src, nil, nil}

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
                            fmt.Println("Key down:", key)
                        } else {
                            panic("wtf")
                        }
                    case video.NavigationEventKeyRelease:
                        key, ok := event.ParseKeyEvent()
                        if ok {
                            fmt.Println("Key up:", key)
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
