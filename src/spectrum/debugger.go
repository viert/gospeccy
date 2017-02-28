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
	DEFAULT_DISASM_LINES = 25
)

type Server struct {
	cpu *z80.Context
}

type Registers struct {
	PC   string `json:"PC"`
	SP   string `json:"SP"`
	AF   string `json:"AF"`
	BC   string `json:"BC"`
	DE   string `json:"DE"`
	HL   string `json:"HL"`
	IX   string `json:"IX"`
	IY   string `json:"IY"`
	AFx  string `json:"AF+"`
	BCx  string `json:"BC+"`
	DEx  string `json:"DE+"`
	HLx  string `json:"HL+"`
	R    string `json:"R"`
	I    string `json:"I"`
	IFF1 bool   `json:"IFF1"`
	IFF2 bool   `json:"IFF2"`
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

func (s *Server) startWeb(host string, port int) {
	listenAddr := fmt.Sprintf("%s:%d", host, port)
	err := http.ListenAndServe(listenAddr, nil)
	if err != nil {
		panic(err)
	}
}
