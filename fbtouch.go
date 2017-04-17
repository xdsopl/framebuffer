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

const EventTypeSyn = 0x00
const EventTypeAbs = 0x03
const EventCodeAbsX = 0x00
const EventCodeAbsY = 0x01

type Position struct {
	X, Y int
}

func square(fb draw.Image, pos Position) {
	for j := 0; j < 10; j++ {
		for i := 0; i < 10; i++ {
			fb.Set(pos.X+i, pos.Y+j, color.White)
		}
	}
}

func painter(fb draw.Image, ev *os.File) {
	pos := Position{-1, -1}
	const ieMax = 64
	const ieSize = int(unsafe.Sizeof(InputEvent{}))
	buf := make([]byte, ieMax * ieSize)
	for {
		n, err := ev.Read(buf)
		if err != nil { die(err) }
		if n == 0 { continue }
		if n % ieSize != 0 { die("partial read") }
		ieNum := n / ieSize
		iev := make([]InputEvent, ieNum)
		err = binary.Read(bytes.NewBuffer(buf[:n]), binary.LittleEndian, &iev)
		if err != nil { die(err) }
		for _, ie := range iev {
			switch ie.Type {
				case EventTypeSyn:
					square(fb, pos)
				case EventTypeAbs:
					switch ie.Code {
						case EventCodeAbsX:
							pos.X = int(ie.Value)
						case EventCodeAbsY:
							pos.Y = int(ie.Value)
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
	fb, err := framebuffer.Open("/dev/fb0")
	if err != nil { die(err) }
	draw.Draw(fb, fb.Bounds(), &image.Uniform{color.Black}, image.ZP, draw.Src)
	painter(fb, ev)
}

