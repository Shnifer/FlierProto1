package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"sync"
)

//Низкий уровень собирателя данных и состяния устройств ввода
//Разбирает сообщения SDL. Не привязан к сцене (ПОКА)
type controlHandler struct {
	//т.к. в перспективе это пойдёт в Апдейт и может быть параллельно -- мутекс
	mu          sync.RWMutex
	pressedKeys map[sdl.Scancode]bool

	Joystick          *sdl.Joystick
	pressedJoybuttons map[uint8]bool
	AxisX, AxisY      float32
}

func newControlHandler(joystick *sdl.Joystick) *controlHandler {
	return &controlHandler{
		pressedKeys:       make(map[sdl.Scancode]bool),
		pressedJoybuttons: make(map[uint8]bool),
		Joystick:          joystick,
	}
}

func (ch *controlHandler) handleSDLKeyboardEvent(ev *sdl.KeyboardEvent) {
	ch.mu.Lock()
	scan := ev.Keysym.Scancode
	switch ev.Type {
	case sdl.KEYDOWN:
		ch.pressedKeys[scan] = true
	case sdl.KEYUP:
		ch.pressedKeys[scan] = false
	}
	ch.mu.Unlock()
}

func (ch *controlHandler) handleJoyButtonEvent(ev *sdl.JoyButtonEvent) {
	ch.mu.Lock()
	button := ev.Button
	switch ev.Type {
	case sdl.JOYBUTTONDOWN:
		ch.pressedJoybuttons[button] = true
	case sdl.JOYBUTTONUP:
		ch.pressedJoybuttons[button] = false
	}
	ch.mu.Unlock()
}

func (ch *controlHandler) GetKey(scancode sdl.Scancode) bool {
	ch.mu.RLock()
	v := ch.pressedKeys[scancode]
	ch.mu.RUnlock()
	return v
}

func (ch *controlHandler) GetJoybutton(button uint8) bool {
	ch.mu.RLock()
	v := ch.pressedJoybuttons[button]
	ch.mu.RUnlock()
	return v
}

const JoyAxisZerozone = 5000

func (ch *controlHandler) updateJoystickAxis() {
	if ch.Joystick == nil {
		return
	}
	x := ch.Joystick.GetAxis(0)
	y := ch.Joystick.GetAxis(1)
	ch.mu.Lock()
	zz := int16(JoyAxisZerozone)
	if x > -zz && x < zz {
		x = 0
	}
	if y > -zz && y < zz {
		y = 0
	}

	ch.AxisX = float32(x) / 32768
	ch.AxisY = -float32(y) / 32768
	ch.mu.Unlock()
}

/* UNUSED right now
type MouseState struct{
	leftButton  bool
	rightButton bool
	x, y        int
}

type controlState struct {
	curMouse, prevMouse MouseState
}

func newControlState() *controlState{
	res:= controlState{}
	return &res
}


func(cs *controlState) Update(){

	cs.prevMouse = cs.curMouse

	mouseX, mouseY, mouseButtonState := sdl.GetMouseState()
	leftButton := mouseButtonState & sdl.ButtonLMask()
	rightButton := mouseButtonState & sdl.ButtonRMask()
	cs.curMouse.x = int(mouseX)
	cs.curMouse.y = int(mouseY)
	cs.curMouse.leftButton = (leftButton != 0)
	cs.curMouse.rightButton = (rightButton != 0)
}
*/
