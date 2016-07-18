/*
fbgrab - grab PNG from Linux framebuffer
Written in 2016 by <Ahmet Inan> <xdsopl@googlemail.com>
To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights to this software to the public domain worldwide. This software is distributed without any warranty.
You should have received a copy of the CC0 Public Domain Dedication along with this software. If not, see <http://creativecommons.org/publicdomain/zero/1.0/>.
*/

package main

import (
	"os"
	"fmt"
	"image/png"
	"framebuffer"
)

func die(err interface{}) {
	fmt.Println(err)
	os.Exit(1)
}

func main() {
	fb, err := framebuffer.Open("/dev/fb0")
	if err != nil { die(err) }
	name := "fbgrab.png"
	file, err := os.Create(name)
	if err != nil { die(err) }
	if err := png.Encode(file, fb); err != nil { die(err) }
}

