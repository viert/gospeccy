package main

import (
	"flag"
	"fmt"
	"spectrum"
)

var romFile string

func init() {
	flag.StringVar(&romFile, "f", "", "ZX Spectrum ROM file")
	flag.Parse()
}

func main() {
	if romFile == "" {
		fmt.Println("Usage: emulator -f <ROMfile>")
		return
	}
	err := spectrum.InstallRom(romFile)
	if err != nil {
		panic(err)
	}
	spectrum.Run()
}
