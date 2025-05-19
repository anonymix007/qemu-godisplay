package main

import (
    "encoding/binary"
    "fmt"
    "sort"

    "github.com/go-gst/go-gst/gst"
    "github.com/go-gst/go-gst/gst/allocators"
    "github.com/go-gst/go-gst/gst/video"
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

func NewRawPicture(width, height, stride, format uint32, data []byte) Picture {
    return &RawPicture{data, width, height, stride, format, stride / width}
}

func (p *RawPicture) CreateCaps() *gst.Caps {
    return gst.NewCapsFromString(fmt.Sprintf("video/x-raw,format=%s,width=%d,height=%d", formats[int(p.fmt)], p.w, p.h))
}

func (p *RawPicture) CreateBuffer() *gst.Buffer {
    return gst.NewBufferFromBytesNoCopy(p.data)
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

    offset []uint64
    stride []int
    size   int64

    fourcc uint32
    mod    uint64

    alloc *allocators.DmaBufAllocator
}

func NewDmaPicture(fd int, width, height uint32, offset, stride []uint32, fourcc uint32, modifier uint64, y0_top bool) Picture {
    // FIXME: gstreamer does not support 0xffffffffffffff
    // https://gitlab.freedesktop.org/mesa/mesa/-/issues/11629
    // https://gitlab.freedesktop.org/gstreamer/gstreamer/-/merge_requests/8213

    offsets := make([]uint64, len(offset))
    for i, o := range offset {
        offsets[i] = uint64(o)
    }

    strides := make([]int, len(stride))
    for i, s := range stride {
        strides[i] = int(s)
    }

    if !sort.SliceIsSorted(offsets, func (i, j int) bool {return offsets[i] < offsets[j]}) {
        panic(fmt.Errorf("cannot handle unsorted offsets: %v", offsets))
    }

    size := int64(offsets[len(offset) - 1]) + int64(strides[len(stride) - 1]) * int64(height)

    return &DmaPicture{fd, width, height, offsets, strides, size, fourcc, modifier, allocators.NewDmaBufAllocator()}
}

func (p *DmaPicture) CreateCaps() *gst.Caps {
    bytes := make([]byte, 4)
    binary.LittleEndian.PutUint32(bytes, p.fourcc)
    fourccStr := string(bytes)

    if p.mod == 0 {
        return gst.NewCapsFromString(fmt.Sprintf("video/x-raw(memory:DMABuf),format=DMA_DRM,width=%d,height=%d,drm-format=%s", p.w, p.h, fourccStr))
    } else {
        return gst.NewCapsFromString(fmt.Sprintf("video/x-raw(memory:DMABuf),format=DMA_DRM,width=%d,height=%d,drm-format=%s:0x%x", p.w, p.h, fourccStr, p.mod))
    }
}

func (p *DmaPicture) CreateBuffer() *gst.Buffer {
    buffer := gst.NewEmptyBuffer()
    memory := p.alloc.AllocDmaBufWithFlags(p.fd, p.size, allocators.FdMemoryFlagDontClose)
    buffer.AppendMemory(memory)

    video.BufferAddVideoMetaFull(buffer, video.FrameFlagNone, video.FormatFromFOURCC(p.fourcc), uint(p.w), uint(p.h), p.offset, p.stride)
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

func NewShmemPicture(fd int, offset, width, height, stride, format uint32) Picture {
    // FIXME: GStreamer can't handle offset
    if offset != 0 {
        panic(fmt.Sprintf("gstreamer cannot handle non-zero offset %d", offset))
    }

    return &ShmemPicture{fd, width, height, int64(stride) * int64(height), offset, stride, format, allocators.NewFdAllocator()}
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
