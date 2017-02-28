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
	inDebug     bool
)

const (
	tactLength     = 285.7142857142857 * 2
	tStatesToBreak = 35000
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
	inDebug = false
	cpuContext = z80.NewContext(true)
	cpuContext.MemoryRead = memoryRead
	cpuContext.MemoryWrite = memoryWrite
	cpuContext.IoRead = ioRead
	cpuContext.IoWrite = ioWrite
	terminal.InitScreenWindow(memory)
}

func startSpectrum() {
	srv := NewServer(cpuContext)
	go srv.startWeb("localhost", 5335)
	lastStop := time.Now()
	var tStates uint64 = 0
	var accTstates uint64 = 0
	for {
		tStates = cpuContext.ExecuteTStates(tStatesToBreak)
		accTstates += tStates
		timedelta = time.Now().Sub(lastStop)
		targetsleep = time.Duration(tactLength * float64(tStates))
		realsleep = time.Duration(targetsleep) - timedelta
		if realsleep > 0 {
			time.Sleep(realsleep)
		}
		lastStop = time.Now()
		if accTstates >= tStatesToBreak {
			accTstates -= tStatesToBreak
			cpuContext.Int(255)
		}
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
