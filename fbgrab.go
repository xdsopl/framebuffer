/*
fbgrab - grab PNG from Linux framebuffer
Written in 2016 by <Ahmet Inan> <xdsopl@googlemail.com>
To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights to this software to the public domain worldwide. This software is distributed without any warranty.
You should have received a copy of the CC0 Public Domain Dedication along with this software. If not, see <http://creativecommons.org/publicdomain/zero/1.0/>.
*/

package main

import (
	"fmt"
	"os"
	"unsafe"
	"syscall"
	"image"
	"image/color"
	"image/png"
)

const FBIOGET_VSCREENINFO = 0x4600
const FBIOPUT_VSCREENINFO = 0x4601
const FBIOGET_FSCREENINFO = 0x4602
const FB_TYPE_PACKED_PIXELS = 0
const FB_VISUAL_TRUECOLOR = 2

// dont worry about uintptr .. there is compatibility code in the kernel :(
type FixScreenInfo struct {
	Id [16]byte
	Smem_start uintptr
	Smem_len, Type, Type_aux, Visual uint32
	Xpanstep, Ypanstep, Ywrapstep uint16
	Line_length uint32
	Mmio_start uintptr
	Mmio_len, Accel uint32
	Capabilities uint16
	Reserved [2]uint16
}

type BitField struct {
	Offset, Length, Msb_right uint32
}

type VarScreenInfo struct {
	Xres, Yres,
	Xres_virtual, Yres_virtual,
	Xoffset, Yoffset,
	Bits_per_pixel, Grayscale uint32
	Red, Green, Blue, Transp BitField
	Nonstd, Activate,
	Height, Width,
	Accel_flags, Pixclock,
	Left_margin, Right_margin, Upper_margin, Lower_margin,
	Hsync_len, Vsync_len, Sync,
	Vmode, Rotate, Colorspace uint32
	Reserved [4]uint32
}

func die(err interface{}) {
	fmt.Println(err)
	os.Exit(1)
}

func main() {
	fbName := "/dev/fb0"
	fbFile, err := os.Open(fbName)
	if err != nil {
		die(err)
	}
	var fixInfo FixScreenInfo
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fbFile.Fd(), FBIOGET_FSCREENINFO, uintptr(unsafe.Pointer(&fixInfo))); errno != 0 {
		die(os.SyscallError{"SYS_IOCTL", errno})
	}
	if fixInfo.Type != FB_TYPE_PACKED_PIXELS {
		die("fixInfo.Type != FB_TYPE_PACKED_PIXELS")
	}
	if fixInfo.Visual != FB_VISUAL_TRUECOLOR {
		die("fixInfo.Visual != FB_VISUAL_TRUECOLOR")
	}
	fbSize := int(fixInfo.Smem_len)
	stride := int(fixInfo.Line_length)
	var varInfo VarScreenInfo
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fbFile.Fd(), FBIOGET_VSCREENINFO, uintptr(unsafe.Pointer(&varInfo))); errno != 0 {
		die(os.SyscallError{"SYS_IOCTL", errno})
	}
	width := int(varInfo.Xres)
	height := int(varInfo.Yres)
	xOffset := int(varInfo.Xoffset)
	yOffset := int(varInfo.Yoffset)
	bitsPerPixel := int(varInfo.Bits_per_pixel)
	bytesPerPixel := bitsPerPixel / 8
	rOffset := int(varInfo.Red.Offset) / 8
	gOffset := int(varInfo.Green.Offset) / 8
	bOffset := int(varInfo.Blue.Offset) / 8
	if bitsPerPixel != 32 {
		die("bitsPerPixel != 32")
	}
	fbMmap, err := syscall.Mmap(int(fbFile.Fd()), 0, fbSize, syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		die(err)
	}
	pngName := "fbgrab.png"
	pngFile, err := os.Create(pngName)
	if err != nil {
		die(err)
	}
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for j := 0; j < height; j++ {
		begin := stride * (j + yOffset)
		end := begin + stride
		fbLine := fbMmap[begin:end]
		for i := 0; i < width; i++ {
			begin := bytesPerPixel * (i + xOffset)
			end := begin + bytesPerPixel
			fbPixel := fbLine[begin:end]
			R := fbPixel[rOffset]
			G := fbPixel[gOffset]
			B := fbPixel[bOffset]
			A := byte(255)
			img.SetNRGBA(i, j, color.NRGBA{R, G, B, A})
		}
	}
	if err := png.Encode(pngFile, img); err != nil {
		die(err)
	}
}

