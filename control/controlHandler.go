package control

import (
	"github.com/veandco/go-sdl2/sdl"
	"sync"
)

//Низкий уровень собирателя данных и состяния устройств ввода
//Разбирает сообщения SDL.
type Handler struct {
	//т.к. в перспективе это пойдёт в Апдейт и может быть параллельно -- мутекс
	mu sync.RWMutex
	//Нажатые прямо сейчас, состояние, опрашивается GetKey
	pressedKeys map[sdl.Scancode]bool
	//те, которых нажимали за этот цикл UpdateControl
	//опрашиваются WasKey
	//функция UpdateWasPressedKey() фиксирует текущее состяние и начинает копить в новый пустой буффер
	wasPressedKeys    map[sdl.Scancode]bool
	bufWasPressedkeys map[sdl.Scancode]bool

	joystick                *sdl.Joystick
	pressedJoybuttons       map[uint8]bool
	wasPressedJoybuttons    map[uint8]bool
	bufWasPressedJoybuttons map[uint8]bool

	axisX, axisY float32
}

func NewControlHandler(joystick *sdl.Joystick) *Handler {
	return &Handler{
		pressedKeys:             make(map[sdl.Scancode]bool),
		wasPressedKeys:          make(map[sdl.Scancode]bool),
		bufWasPressedkeys:       make(map[sdl.Scancode]bool),
		pressedJoybuttons:       make(map[uint8]bool),
		wasPressedJoybuttons:    make(map[uint8]bool),
		bufWasPressedJoybuttons: make(map[uint8]bool),
		joystick:                joystick,
	}
}

func (ch *Handler) HandleSDLEvent(event sdl.Event) {
	switch ev := event.(type) {
	case *sdl.KeyboardEvent:
		ch.handleSDLKeyboardEvent(ev)
	case *sdl.JoyButtonEvent:
		ch.handleJoyButtonEvent(ev)

		//ПОЗЖЕ МЫШЬ и КНОПКИ ДЖОЙСТИКА
	//case *sdl.MouseMotionEvent:
	default: //не наше событие
	}
}

//Запускается в главном цикле по тикам IO
//Фиксирует состояние положение осей
//запускается ~50 раз в секунду
func (ch *Handler) IOUpdate() {
	ch.mu.Lock()
	ch.updateJoystickAxis()
	ch.mu.Unlock()
}

//Запускается в главном цикле по тикам Physic
//Фиксирует состояние нажатых клавиш
//запускается ~250 раз в секунду
func (ch *Handler) BeforeUpdate() {
	ch.mu.Lock()
	ch.wasPressedJoybuttons = ch.bufWasPressedJoybuttons
	ch.wasPressedKeys = ch.bufWasPressedkeys
	ch.bufWasPressedJoybuttons = make(map[uint8]bool)
	ch.bufWasPressedkeys = make(map[sdl.Scancode]bool)
	ch.mu.Unlock()
}

func (ch *Handler) HasJoystick() bool {
	return (ch.joystick != nil)
}

func (ch *Handler) handleSDLKeyboardEvent(ev *sdl.KeyboardEvent) {
	ch.mu.Lock()
	scan := ev.Keysym.Scancode
	switch ev.Type {
	case sdl.KEYDOWN:
		if !ch.pressedKeys[scan] {
			ch.bufWasPressedkeys[scan] = true
		}
		ch.pressedKeys[scan] = true
	case sdl.KEYUP:
		ch.pressedKeys[scan] = false
	}
	ch.mu.Unlock()
}

func (ch *Handler) handleJoyButtonEvent(ev *sdl.JoyButtonEvent) {
	ch.mu.Lock()
	button := ev.Button
	switch ev.Type {
	case sdl.JOYBUTTONDOWN:
		if !ch.pressedJoybuttons[button] {
			ch.bufWasPressedJoybuttons[button] = true
		}
		ch.pressedJoybuttons[button] = true
	case sdl.JOYBUTTONUP:
		ch.pressedJoybuttons[button] = false
	}
	ch.mu.Unlock()
}

func (ch *Handler) GetKey(scancode sdl.Scancode) bool {
	ch.mu.RLock()
	v := ch.pressedKeys[scancode]
	ch.mu.RUnlock()
	return v
}

func (ch *Handler) WasKey(scancode sdl.Scancode) bool {
	ch.mu.RLock()
	v := ch.wasPressedKeys[scancode]
	ch.mu.RUnlock()
	return v
}

func (ch *Handler) GetJoybutton(button uint8) bool {
	ch.mu.RLock()
	v := ch.pressedJoybuttons[button]
	ch.mu.RUnlock()
	return v
}

func (ch *Handler) WasJoybutton(button uint8) bool {
	ch.mu.RLock()
	v := ch.wasPressedJoybuttons[button]
	ch.mu.RUnlock()
	return v
}

func (ch *Handler) AxisX() float32 {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return ch.axisX
}

func (ch *Handler) AxisY() float32 {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return ch.axisY
}

const JoyAxisZerozone = 5000

//Mutex снаружи вызова
func (ch *Handler) updateJoystickAxis() {
	if ch.joystick == nil {
		return
	}

	//Mutex снаружи вызова
	//ch.mu.Lock()
	x := ch.joystick.GetAxis(0)
	y := ch.joystick.GetAxis(1)

	zz := int16(JoyAxisZerozone)
	if x > -zz && x < zz {
		x = 0
	}
	if y > -zz && y < zz {
		y = 0
	}

	ch.axisX = float32(x) / 32768
	ch.axisY = -float32(y) / 32768
	//ch.mu.Unlock()
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
