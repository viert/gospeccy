package spectrum

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/viert/z80"
	"io"
	"net/http"
	"strconv"
)

const (
	DEFAULT_DISASM_LINES = 50
)

type Server struct {
	cpu *z80.Context
}

type Registers struct {
	PC                 string `json:"PC"`
	SP                 string `json:"SP"`
	AF                 string `json:"AF"`
	BC                 string `json:"BC"`
	DE                 string `json:"DE"`
	HL                 string `json:"HL"`
	IX                 string `json:"IX"`
	IY                 string `json:"IY"`
	AFx                string `json:"AF+"`
	BCx                string `json:"BC+"`
	DEx                string `json:"DE+"`
	HLx                string `json:"HL+"`
	R                  string `json:"R"`
	I                  string `json:"I"`
	IFF1               bool   `json:"IFF1"`
	IFF2               bool   `json:"IFF2"`
	IM                 byte   `json:"IM"`
	BreakpointsEnabled bool   `json:"breakpoints_enabled"`
	State              string `json:"cpu_state"`
}

type DisasmEntry struct {
	Addr string `json:"addr"`
	Code string `json:"code"`
}

func NewServer(cpu *z80.Context) *Server {
	server := new(Server)
	server.cpu = cpu
	router := mux.NewRouter()
	router.HandleFunc("/dump/registers", server.registersHandler).Methods("GET")
	router.HandleFunc("/dump/memory", server.memoryHandler).Methods("GET")
	router.HandleFunc("/breakpoints", server.listBreakpointsHandler).Methods("GET")
	router.HandleFunc("/breakpoints/{addr}", server.addBreakpointHandler).Methods("POST")
	router.HandleFunc("/breakpoints/{addr}", server.removeBreakpointHandler).Methods("DELETE")
	router.HandleFunc("/control/{command}", server.controlHandler).Methods("POST")

	http.Handle("/", router)
	return server
}

func (s *Server) registersHandler(w http.ResponseWriter, r *http.Request) {
	regs := new(Registers)
	dump := s.cpu.LatestDump
	regs.AF = fmt.Sprintf("%04X", dump.AF)
	regs.BC = fmt.Sprintf("%04X", dump.BC)
	regs.DE = fmt.Sprintf("%04X", dump.DE)
	regs.HL = fmt.Sprintf("%04X", dump.HL)
	regs.IX = fmt.Sprintf("%04X", dump.IX)
	regs.IY = fmt.Sprintf("%04X", dump.IY)
	regs.AFx = fmt.Sprintf("%04X", dump.AF_)
	regs.BCx = fmt.Sprintf("%04X", dump.BC_)
	regs.DEx = fmt.Sprintf("%04X", dump.DE_)
	regs.HLx = fmt.Sprintf("%04X", dump.HL_)
	regs.SP = fmt.Sprintf("%04X", dump.SP)
	regs.PC = fmt.Sprintf("%04X", dump.PC)
	regs.R = fmt.Sprintf("%02X", dump.R)
	regs.I = fmt.Sprintf("%02X", dump.I)
	regs.IFF1 = dump.IFF1
	regs.IFF2 = dump.IFF2
	regs.State = s.cpu.State()
	regs.IM = s.cpu.IM
	regs.BreakpointsEnabled = s.cpu.GetBPMode()
	data, err := json.Marshal(regs)
	if err != nil {
		http.Error(w, "Error marshalling data", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, string(data))
}

func (s *Server) memoryHandler(w http.ResponseWriter, r *http.Request) {
	var startAddr int64 = 0
	var addr uint16 = 0
	var nextAddr uint16 = 0
	var linesCount int64 = DEFAULT_DISASM_LINES
	var dLine string
	var err error

	r.ParseForm()

	start := r.Form.Get("start")
	if start != "" {
		startAddr, err = strconv.ParseInt(start, 16, 16)
	}

	lines := r.Form.Get("lines")
	if lines != "" {
		linesCount, err = strconv.ParseInt(lines, 10, 16)
		if linesCount < 1 {
			linesCount = DEFAULT_DISASM_LINES
		}
	}

	result := make([]DisasmEntry, linesCount)
	addr = uint16(startAddr)
	for i := 0; i < len(result); i++ {
		dLine, nextAddr = s.cpu.Disassemble(uint16(addr))
		result[i] = DisasmEntry{fmt.Sprintf("%04X", addr), dLine}
		addr = nextAddr
	}

	data, err := json.Marshal(result)
	if err != nil {
		http.Error(w, "Error marshalling data", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, string(data))
}

func (s *Server) listBreakpointsHandler(w http.ResponseWriter, r *http.Request) {
	breakpoints := s.cpu.GetBreakpoints()
	data, err := json.Marshal(breakpoints)
	if err != nil {
		http.Error(w, "Error marshalling data", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, string(data))
}

func (s *Server) addBreakpointHandler(w http.ResponseWriter, r *http.Request) {
	addr := mux.Vars(r)["addr"]
	iAddr, err := strconv.ParseInt(addr, 16, 16)
	if err != nil {
		http.Error(w, "Error in breakpoint address", 400)
	}
	s.cpu.AddBreakpoint(uint16(iAddr))
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, fmt.Sprintf("{\"breakpoint\":\"%s\"}", addr))
}

func (s *Server) removeBreakpointHandler(w http.ResponseWriter, r *http.Request) {
	addr := mux.Vars(r)["addr"]
	iAddr, err := strconv.ParseInt(addr, 16, 16)
	if err != nil {
		http.Error(w, "Error in breakpoint address", 400)
	}
	s.cpu.RemoveBreakpoint(uint16(iAddr))
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, fmt.Sprintf("{\"breakpoint\":\"%s\"}", addr))
}

func (s *Server) controlHandler(w http.ResponseWriter, r *http.Request) {
	command := mux.Vars(r)["command"]
	switch command {
	case "stop":
		s.cpu.Stop()
	case "step":
		if emulatorMode != EM_STEP {
			fmt.Println("Setting emulator to STEP MODE")
			emulatorMode = EM_STEP
		}
		resume <- true
	case "resume":
		s.cpu.Resume()
		if emulatorMode != EM_RUN {
			emulatorMode = EM_RUN
			fmt.Println("Setting emulator to RUN MODE")
			resume <- true
		}
	case "enable_bp":
		s.cpu.SetBPMode(true)
	case "disable_bp":
		s.cpu.SetBPMode(false)
	}
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, fmt.Sprintf("{\"status\": \"ok\",\"command\":\"%s\"}", command))
}

func (s *Server) startWeb(host string, port int) {
	listenAddr := fmt.Sprintf("%s:%d", host, port)
	err := http.ListenAndServe(listenAddr, nil)
	if err != nil {
		panic(err)
	}
}
