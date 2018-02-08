package main

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"log"
	"runtime"
	"time"
	"github.com/Shnifer/flierproto1/fps"
)

//Константы экрана
//TODO:вынести параметры экрана во внешний файл конфигурации
var winW int32
var winH int32

const ResourcePath = "res/"
const ClientDataPath = ResourcePath + "pilot/"

type GameState byte

const (
	//TODO: Экран перезагрузки
	state_Login GameState = iota
	state_PilotSpace
	state_NaviSpace
)

func main() {

	runtime.LockOSThread()

	deferMe, renderer, Joystick := InitSomeShit()
	defer deferMe()

	fmt.Println("got", renderer, Joystick)

	ControlHandler := newControlHandler(Joystick)

	PilotScene := NewPilotScene(renderer, ControlHandler)
	PilotScene.Init()

	//Проверочный показывать фпс, он же заглушка на систему каналов
	initFPS:=fps.InitStruct{
		MIN_FRAME_MS:           DEFVAL.MIN_FRAME_MS,
		MIN_PHYS_MS:            DEFVAL.MIN_PHYS_MS,
		MAX_FRAME_MS:           DEFVAL.MAX_FRAME_MS,
		MAX_PHYS_MS:           DEFVAL.MAX_PHYS_MS,
		FPS_UPDATE_MS: DEFVAL.FPS_UPDATE_MS,
		TickerBalancerOverhead: DEFVAL.TickerBalancerOverhead,
	}
	ShowFpsTick, fpsControl:=fps.Start(initFPS)
	defer close(fpsControl)

	lastPhysFrame := time.Now()

	graphFrameN, physFrameN, ioFrameN := 0, 0, 0
	var maxDt, maxGraphT, maxPhysT float32

	breakMainLoop := make(chan bool, 1)

loop:
	for {
		select {
		case <-breakMainLoop:
			break loop
		//Время показать фпс
		case <-ShowFpsTick:
			fpsControl <- fps.FpsData{graphFrameN, physFrameN, ioFrameN,
				maxDt, maxGraphT, maxPhysT}
			maxDt = 0.0
			maxGraphT = 0.0
			maxPhysT = 0.0
		case <-fps.PTick:
			deltaTime := float32(time.Since(lastPhysFrame).Seconds())
			if deltaTime > maxDt {
				maxDt = deltaTime
			}
			lastPhysFrame = time.Now()
			physFrameN++
			PilotScene.Update(deltaTime)
			T := float32(time.Since(lastPhysFrame).Seconds())
			if T > maxPhysT {
				maxPhysT = T
			}
		default:
			select {
			case <-fps.GTick:
				graphFrameN++
				start := time.Now()
				renderer.Clear()
				PilotScene.Draw()
				renderer.Present()
				T := float32(time.Since(start).Seconds())
				if T > maxGraphT {
					maxGraphT = T
				}
			default:
				ioFrameN++
				DoMainLoopIO(breakMainLoop, ControlHandler)
			}
		}
	}
}

func DoMainLoopIO(breakMainLoop chan bool, handler *controlHandler) {
	//Проверяем и хэндлим события СДЛ. Выход -- обязательно, а то не закроется
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch ev := event.(type) {
		case *sdl.QuitEvent:
			log.Println("quit")
			breakMainLoop <- true
		case *sdl.KeyboardEvent:
			//Кнопку выходи обрабатываем здесь, чтобы порвать главный цикл
			scan := ev.Keysym.Scancode
			if scan == sdl.SCANCODE_ESCAPE {
				breakMainLoop <- true
			}
			//Остальное в хэндлере
			handler.handleSDLKeyboardEvent(ev)
		case *sdl.MouseMotionEvent:
		case *sdl.JoyButtonEvent:
			handler.handleJoyButtonEvent(ev)
		default:

		}
	}
	handler.updateJoystickAxis()
}

func timeCheck(caption string) func() {
	Start := time.Now()
	return func() {
		log.Println(caption, time.Since(Start).Seconds()*1000, "ms")
	}
}
