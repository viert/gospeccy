package spectrum

import (
	"github.com/viert/z80"
	"os"
	"port"
	"terminal"
	"time"
)

var (
	cpuContext  *z80.Context
	memory      []byte
	timedelta   time.Duration
	targetsleep time.Duration
	realsleep   time.Duration
)

const (
	tactLength = 285.7142857142857 * 3
)

func memoryRead(addr uint16) byte {
	return memory[addr]
}

func memoryWrite(addr uint16, data byte) {
	if addr >= 16384 {
		memory[addr] = data
	}
	if addr < 16384+6912 {
		terminal.SetRedraw()
	}
}

func ioRead(portNum uint16) byte {
	return port.GetIn(portNum)
}

func ioWrite(portNum uint16, data byte) {
	port.SetOut(portNum, data)
}

func init() {
	memory = make([]byte, 65536)
	cpuContext = z80.NewContext(true)
	cpuContext.MemoryRead = memoryRead
	cpuContext.MemoryWrite = memoryWrite
	cpuContext.IoRead = ioRead
	cpuContext.IoWrite = ioWrite
	terminal.InitScreenWindow(memory)
}

func startSpectrum() {
	// create standard maskable interrupt
	go func() {
		c := time.Tick(20 * time.Millisecond)
		for range c {
			cpuContext.Int(255)
		}
	}()

	lastStop := time.Now()

	var tstates uint64

	for {
		tstates = cpuContext.ExecuteTStates(50000)
		timedelta = time.Now().Sub(lastStop)
		targetsleep = time.Duration(tactLength * float64(tstates))
		realsleep = time.Duration(targetsleep) - timedelta
		if realsleep > 0 {
			time.Sleep(realsleep)
		}
		lastStop = time.Now()
	}
}

func InstallRom(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Read(memory)
	if err != nil {
		return err
	}
	return nil
}

func Run() {
	terminal.StartTerminal(func() {
		go startSpectrum()
	})
}
