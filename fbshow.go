/*
fbshow - show image on Linux framebuffer
Written in 2016 by <Ahmet Inan> <xdsopl@googlemail.com>
To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights to this software to the public domain worldwide. This software is distributed without any warranty.
You should have received a copy of the CC0 Public Domain Dedication along with this software. If not, see <http://creativecommons.org/publicdomain/zero/1.0/>.
*/

package main

import (
	"os"
	"fmt"
	"flag"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"framebuffer"
)

func die(err interface{}) {
	fmt.Println(err)
	os.Exit(1)
}

func main() {
	flag.Parse()
	if len(flag.Args()) != 1 { die("usage: fbshow file") }
	name := flag.Args()[0]
	file, err := os.Open(name)
	if err != nil { die(err) }
	img, _, err := image.Decode(file)
	if err != nil { die(err) }
	fb, err := framebuffer.Open("/dev/fb0")
	if err != nil { die(err) }
	draw.Draw(fb, img.Bounds().Sub(img.Bounds().Min).Add(fb.Bounds().Min).Add(fb.Bounds().Size().Sub(img.Bounds().Size()).Div(2)), img, img.Bounds().Min, draw.Src)
}

