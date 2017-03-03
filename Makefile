SRC_SPECTRUM = src/spectrum/spectrum.go src/spectrum/debugger.go
SRC_PORT = src/port/port.go
SRC_TERMINAL = src/terminal/keyboard.go src/terminal/screen.go
SRC_EMULATOR = $(SRC_SPECTRUM) $(SRC_PORT) $(SRC_TERMINAL)

all: emulator

emulator: $(SRC_EMULATOR)
	GOPATH=$(PWD) go build src/emulator.go

clean:
	rm -f emulator
