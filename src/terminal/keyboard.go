package terminal

import (
	"github.com/go-gl/glfw/v3.1/glfw"
	"port"
)

type keyDesc struct {
	port uint16
	mask byte
}

var (
	keyMap map[glfw.Key][]keyDesc = map[glfw.Key][]keyDesc{
		glfw.KeyQ:         {keyDesc{0xFBFE, 0376}},
		glfw.KeyW:         {keyDesc{0xFBFE, 0375}},
		glfw.KeyE:         {keyDesc{0xFBFE, 0373}},
		glfw.KeyR:         {keyDesc{0xFBFE, 0367}},
		glfw.KeyT:         {keyDesc{0xFBFE, 0357}},
		glfw.KeyA:         {keyDesc{0xFDFE, 0376}},
		glfw.KeyS:         {keyDesc{0xFDFE, 0375}},
		glfw.KeyD:         {keyDesc{0xFDFE, 0373}},
		glfw.KeyF:         {keyDesc{0xFDFE, 0367}},
		glfw.KeyG:         {keyDesc{0xFDFE, 0357}},
		glfw.KeyP:         {keyDesc{0xDFFE, 0376}},
		glfw.KeyO:         {keyDesc{0xDFFE, 0375}},
		glfw.KeyI:         {keyDesc{0xDFFE, 0373}},
		glfw.KeyU:         {keyDesc{0xDFFE, 0367}},
		glfw.KeyY:         {keyDesc{0xDFFE, 0357}},
		glfw.KeyLeftShift: {keyDesc{0xFEFE, 0376}},
		glfw.KeyZ:         {keyDesc{0xFEFE, 0375}},
		glfw.KeyX:         {keyDesc{0xFEFE, 0373}},
		glfw.KeyC:         {keyDesc{0xFEFE, 0367}},
		glfw.KeyV:         {keyDesc{0xFEFE, 0357}},
		glfw.KeyEnter:     {keyDesc{0xBFFE, 0376}},
		glfw.KeyL:         {keyDesc{0xBFFE, 0375}},
		glfw.KeyK:         {keyDesc{0xBFFE, 0373}},
		glfw.KeyJ:         {keyDesc{0xBFFE, 0367}},
		glfw.KeyH:         {keyDesc{0xBFFE, 0357}},
		glfw.KeySpace:     {keyDesc{0x7FFE, 0376}},
		glfw.KeyLeftAlt:   {keyDesc{0x7FFE, 0375}},
		glfw.KeyM:         {keyDesc{0x7FFE, 0373}},
		glfw.KeyN:         {keyDesc{0x7FFE, 0367}},
		glfw.KeyB:         {keyDesc{0x7FFE, 0357}},
		glfw.Key1:         {keyDesc{0xF7FE, 0376}},
		glfw.Key2:         {keyDesc{0xF7FE, 0375}},
		glfw.Key3:         {keyDesc{0xF7FE, 0373}},
		glfw.Key4:         {keyDesc{0xF7FE, 0367}},
		glfw.Key5:         {keyDesc{0xF7FE, 0357}},
		glfw.Key0:         {keyDesc{0xEFFE, 0376}},
		glfw.Key9:         {keyDesc{0xEFFE, 0375}},
		glfw.Key8:         {keyDesc{0xEFFE, 0373}},
		glfw.Key7:         {keyDesc{0xEFFE, 0367}},
		glfw.Key6:         {keyDesc{0xEFFE, 0357}},
		glfw.KeyBackspace: {keyDesc{0xFEFE, 0376}, keyDesc{0xEFFE, 0376}},
		glfw.KeyLeft:      {keyDesc{0xFEFE, 0376}, keyDesc{0xF7FE, 0357}},
		glfw.KeyRight:     {keyDesc{0xFEFE, 0376}, keyDesc{0xEFFE, 0373}},
		glfw.KeyUp:        {keyDesc{0xFEFE, 0376}, keyDesc{0xEFFE, 0367}},
		glfw.KeyDown:      {keyDesc{0xFEFE, 0376}, keyDesc{0xEFFE, 0357}},
	}
)

func keydown(data keyDesc) {
	currentPortData := port.GetIn(data.port)
	port.SetIn(data.port, currentPortData&data.mask)
}

func keyup(data keyDesc) {
	currentPortData := port.GetIn(data.port)
	invertMask := data.mask ^ 0xFF
	port.SetIn(data.port, currentPortData|invertMask)
}

func keyaction(dataList []keyDesc, action glfw.Action) {
	for _, data := range dataList {
		if action > 0 {
			keydown(data)
		} else {
			keyup(data)
		}
	}
}

func keyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	data, found := keyMap[key]
	keyaction(data, action)
}
