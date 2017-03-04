package spectrum

import (
	"fmt"
	"github.com/viert/z80"
	"os"
	"port"
	"terminal"
	"time"
)

const (
	EM_STEP = iota
	EM_RUN
)

var (
	cpuContext   *z80.Context
	memory       []byte
	timedelta    time.Duration
	targetsleep  time.Duration
	realsleep    time.Duration
	emulatorMode int
	resume       chan bool
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
	resume = make(chan bool)
	cpuContext = z80.NewContext(true)
	cpuContext.MemoryRead = memoryRead
	cpuContext.MemoryWrite = memoryWrite
	cpuContext.IoRead = ioRead
	cpuContext.IoWrite = ioWrite
	cpuContext.SetBPMode(true)
	emulatorMode = EM_RUN
	terminal.InitScreenWindow(memory)
}

func resetSpectrum() {
	saveEMode := emulatorMode
	emulatorMode = EM_STEP
	cpuContext.Stop()
	cpuContext.Reset()
	for i := 16384; i < 65536; i++ {
		memory[i] = 0
	}
	emulatorMode = saveEMode
}

func startSpectrum() {
	srv := NewServer(cpuContext)
	go srv.startWeb("localhost", 5335)
	lastStop := time.Now()
	var tStates uint64 = 0
	var accTstates uint64 = 0
	for {
		switch emulatorMode {
		case EM_RUN:
			tStates = cpuContext.ExecuteTStates(z80.T_STATES_TO_BREAK)
			if tStates == 0 {
				fmt.Println("0 tstates while ExecuteTStates, switching to step mode")
				emulatorMode = EM_STEP
			}
		case EM_STEP:
			<-resume
			tStates = cpuContext.Execute()
		}
		accTstates += tStates
		timedelta = time.Now().Sub(lastStop)
		targetsleep = time.Duration(z80.T_LENGTH * float64(tStates))
		realsleep = time.Duration(targetsleep) - timedelta
		if realsleep > 0 {
			time.Sleep(realsleep)
		}
		lastStop = time.Now()
		if accTstates >= z80.T_STATES_TO_BREAK {
			accTstates -= z80.T_STATES_TO_BREAK
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
