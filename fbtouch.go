/*
fbtouch - show touchscreen events on Linux framebuffer
Written in 2017 by <Ahmet Inan> <xdsopl@googlemail.com>
To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights to this software to the public domain worldwide. This software is distributed without any warranty.
You should have received a copy of the CC0 Public Domain Dedication along with this software. If not, see <http://creativecommons.org/publicdomain/zero/1.0/>.
*/

package main

import (
	"os"
	"fmt"
	"flag"
	"bytes"
	"errors"
	"unsafe"
	"syscall"
	"image"
	"image/draw"
	"image/color"
	"framebuffer"
	"encoding/binary"
)

func die(err interface{}) {
	fmt.Println(err)
	os.Exit(1)
}

type InputEvent struct {
	Time syscall.Timeval
	Type, Code uint16
	Value int32
}

type InputAbsInfo struct {
	Value, Minimum, Maximum, Fuzz, Flat, Resolution int32
}

func EVIOCGABS(abs uintptr) uintptr {
	return 2149074240 + abs
}

func GetAbsInfo(ev *os.File, abs uintptr) (InputAbsInfo, error) {
	var tmp InputAbsInfo
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, ev.Fd(), EVIOCGABS(abs), uintptr(unsafe.Pointer(&tmp)))
	if errno != 0 {
		return tmp, &os.SyscallError{"SYS_IOCTL", errno}
	}
	return tmp, nil
}

const EventTypeSyn = 0x00
const EventTypeKey = 0x01
const EventTypeAbs = 0x03
const EventCodeAbsX = 0x00
const EventCodeAbsY = 0x01
const EventCodeBtnTouch = 0x14a

type Position struct {
	X, Y int
}

func GetEvents(ev *os.File) ([]InputEvent, error) {
	const ieMax = 64
	const ieSize = int(unsafe.Sizeof(InputEvent{}))
	buf := make([]byte, ieMax * ieSize)
	n, err := ev.Read(buf)
	if err != nil { return nil, err }
	if n == 0 { return nil, nil }
	if n % ieSize != 0 { return nil, errors.New("partial read") }
	ieNum := n / ieSize
	iev := make([]InputEvent, ieNum)
	err = binary.Read(bytes.NewBuffer(buf[:n]), binary.LittleEndian, &iev)
	if err != nil { return nil, err }
	return iev, nil
}

func square(fb draw.Image, pos Position, col color.Color) {
	radius := 35
	for i := -radius; i <= radius; i++ {
		fb.Set(pos.X+i, pos.Y-radius, col)
		fb.Set(pos.X+i, pos.Y+radius, col)
		fb.Set(pos.X-radius, pos.Y+i, col)
		fb.Set(pos.X+radius, pos.Y+i, col)
	}
}

func singleTouch(fb draw.Image, ev *os.File, absX, absY InputAbsInfo) {
	pos := Position{-1, -1}
	touching := false
	old := pos
	for {
		iev, err := GetEvents(ev)
		if err != nil { die(err) }
		for _, ie := range iev {
			switch ie.Type {
				case EventTypeSyn:
					square(fb, old, color.Black)
					old = pos
					if touching {
						square(fb, pos, color.White)
					}
				case EventTypeKey:
					switch ie.Code {
						case EventCodeBtnTouch:
							if ie.Value == 0 {
								touching = false
							} else {
								touching = true
							}
					}
				case EventTypeAbs:
					switch ie.Code {
						case EventCodeAbsX:
							pos.X = fb.Bounds().Min.X + (fb.Bounds().Dx() * int(ie.Value - absX.Minimum)) / int(absX.Maximum - absX.Minimum)
						case EventCodeAbsY:
							pos.Y = fb.Bounds().Min.Y + (fb.Bounds().Dy() * int(ie.Value - absY.Minimum)) / int(absY.Maximum - absY.Minimum)
					}
			}
		}
	}
}

func main() {
	flag.Parse()
	if len(flag.Args()) != 1 { die("usage: fbtouch /dev/input/eventN") }
	name := flag.Args()[0]
	ev, err := os.Open(name)
	if err != nil { die(err) }
	absX, err := GetAbsInfo(ev, EventCodeAbsX)
	if err != nil { die(err) }
	absY, err := GetAbsInfo(ev, EventCodeAbsY)
	if err != nil { die(err) }
	fb, err := framebuffer.Open("/dev/fb0")
	if err != nil { die(err) }
	draw.Draw(fb, fb.Bounds(), &image.Uniform{color.Black}, image.ZP, draw.Src)
	singleTouch(fb, ev, absX, absY)
}

