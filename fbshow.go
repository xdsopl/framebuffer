/*
fbshow - show image on Linux framebuffer
Written in 2016 by <Ahmet Inan> <xdsopl@googlemail.com>
To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights to this software to the public domain worldwide. This software is distributed without any warranty.
You should have received a copy of the CC0 Public Domain Dedication along with this software. If not, see <http://creativecommons.org/publicdomain/zero/1.0/>.
*/

package main

import (
	"fmt"
	"os"
	"flag"
	"unsafe"
	"syscall"
	"image"
	"image/draw"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
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

type NBGR32 struct {
	Pix []uint8
	Stride int
	Rect image.Rectangle
}

func (p *NBGR32) Bounds() image.Rectangle { return p.Rect }
func (p *NBGR32) ColorModel() color.Model { return color.NRGBAModel }
func (p *NBGR32) PixOffset(x, y int) int { return y * p.Stride + x * 4 }

func (p *NBGR32) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(p.Rect)) { return }
	i := p.PixOffset(x, y)
	c1 := color.NRGBAModel.Convert(c).(color.NRGBA)
	p.Pix[i+0] = c1.B
	p.Pix[i+1] = c1.G
	p.Pix[i+2] = c1.R
}

func (p *NBGR32) At(x, y int) color.Color {
	if !(image.Point{x, y}.In(p.Rect)) { return color.NRGBA{} }
	i := p.PixOffset(x, y)
	return color.NRGBA{p.Pix[i+2], p.Pix[i+1], p.Pix[i+0], 255}
}

func die(err interface{}) {
	fmt.Println(err)
	os.Exit(1)
}

func main() {
	flag.Parse()
	if len(flag.Args()) != 1 { die("usage: fbshow file") }
	imgName := flag.Args()[0]
	imgFile, err := os.Open(imgName)
	if err != nil { die(err) }
	img, _, err := image.Decode(imgFile)
	if err != nil { die(err) }

	fbName := "/dev/fb0"
	fbFile, err := os.OpenFile(fbName, os.O_RDWR, os.ModeDevice)
	if err != nil { die(err) }
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
	var varInfo VarScreenInfo
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fbFile.Fd(), FBIOGET_VSCREENINFO, uintptr(unsafe.Pointer(&varInfo))); errno != 0 {
		die(os.SyscallError{"SYS_IOCTL", errno})
	}
	//fmt.Println("Red.Offset =", varInfo.Red.Offset, "Red.Length =", varInfo.Red.Length, "Red.Msb_right =", varInfo.Red.Msb_right)
	//fmt.Println("Green.Offset =", varInfo.Green.Offset, "Green.Length =", varInfo.Green.Length, "Green.Msb_right =", varInfo.Green.Msb_right)
	//fmt.Println("Blue.Offset =", varInfo.Blue.Offset, "Blue.Length =", varInfo.Blue.Length, "Blue.Msb_right =", varInfo.Blue.Msb_right)
	//fmt.Println("Transp.Offset =", varInfo.Transp.Offset, "Transp.Length =", varInfo.Transp.Length, "Transp.Msb_right =", varInfo.Transp.Msb_right)
	if varInfo.Bits_per_pixel != 32 { die("varInfo.Bits_per_pixel != 32") }
	if varInfo.Blue.Length != 8 { die("varInfo.Blue.Length != 8") }
	if varInfo.Blue.Offset != 0 { die("varInfo.Blue.Offset != 0") }
	if varInfo.Green.Length != 8 { die("varInfo.Green.Length != 8") }
	if varInfo.Green.Offset != 8 { die("varInfo.Green.Offset != 8") }
	if varInfo.Red.Length != 8 { die("varInfo.Red.Length != 8") }
	if varInfo.Red.Offset != 16 { die("varInfo.Red.Offset != 16") }
	if varInfo.Transp.Length != 0 { die("varInfo.Transp.Length != 0") }
	if varInfo.Transp.Offset != 0 { die("varInfo.Transp.Offset != 0") }
	fbMmap, err := syscall.Mmap(int(fbFile.Fd()), 0, int(fixInfo.Smem_len), syscall.PROT_READ | syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil { die(err) }
	fbImg := &NBGR32{fbMmap, int(fixInfo.Line_length), image.Rect(0, 0, int(varInfo.Xres), int(varInfo.Yres)).Add(image.Point{int(varInfo.Xoffset), int(varInfo.Yoffset)})}

	draw.Draw(fbImg, img.Bounds().Sub(img.Bounds().Min).Add(fbImg.Bounds().Min), img, img.Bounds().Min, draw.Src)
}

