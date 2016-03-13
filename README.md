GoSpeccy
========

GoSpeccy is a ZX Spectrum 48k emulator written in Go just for fun.


Build
-----


```
cd <sourcedir>
export GOPATH=$PWD
go get github.com/viert/z80
go get github.com/go-gl/gl/v2.1/gl
go get github.com/go-gl/glfw/v3.1/glfw
go build src/emulator.go
```

Run
-----

Run emulator with ```./emulator -f <romfile>``` command. Yes, you have to get Spectrum 48K ROM file.
