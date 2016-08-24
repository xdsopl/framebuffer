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
	"time"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"image/gif"
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
	_, str, err := image.DecodeConfig(file)
	if err != nil { die(err) }
	_, err = file.Seek(0, 0)
	if err != nil { die(err) }
	fb, err := framebuffer.Open("/dev/fb0")
	if err != nil { die(err) }
	if str == "gif" {
		all, err := gif.DecodeAll(file)
		if err != nil { die(err) }
		img := all.Image[0]
		dst := img.Bounds().Sub(img.Bounds().Min).Add(fb.Bounds().Min).Add(fb.Bounds().Size().Sub(img.Bounds().Size()).Div(2))
		src := img.Bounds().Min
		for {
			for idx, img := range all.Image {
				draw.Draw(fb, dst, img, src, draw.Src)
				time.Sleep(time.Duration(all.Delay[idx]) * 10 * time.Millisecond)
			}
		}
		return
	}
	img, _, err := image.Decode(file)
	if err != nil { die(err) }
	dst := img.Bounds().Sub(img.Bounds().Min).Add(fb.Bounds().Min).Add(fb.Bounds().Size().Sub(img.Bounds().Size()).Div(2))
	src := img.Bounds().Min
	draw.Draw(fb, dst, img, src, draw.Src)
}

