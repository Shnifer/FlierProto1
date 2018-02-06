package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"sync"
)

//Низкий уровень собирателя данных и состяния устройств ввода
//Разбирает сообщения SDL. Не привязан к сцене (ПОКА)
type controlHandler struct{
	//т.к. в перспективе это пойдёт в Апдейт и может быть параллельно -- мутекс
	mu sync.RWMutex
	pressedKeys map[sdl.Scancode]bool
}

func newControlHandler() *controlHandler{
	return &controlHandler{pressedKeys: make(map[sdl.Scancode]bool)}
}

func (ch *controlHandler) handleSDLKeyboardEvent(ev *sdl.KeyboardEvent){
	ch.mu.Lock()
	scan:=ev.Keysym.Scancode
	switch ev.Type{
		case sdl.KEYDOWN:
			ch.pressedKeys[scan]=true
	case sdl.KEYUP:
		ch.pressedKeys[scan]=false
	}
	ch.mu.Unlock()
}

func (ch *controlHandler) GetKey(scancode sdl.Scancode) bool{
	ch.mu.RLock()
	v:=ch.pressedKeys[scancode]
	ch.mu.RUnlock()
	return v
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