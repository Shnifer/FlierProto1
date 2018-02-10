package main

import (
	"github.com/Shnifer/flierproto1/control"
	"github.com/Shnifer/flierproto1/fps"
	"github.com/veandco/go-sdl2/sdl"
	"log"
	"runtime"
	"time"
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

	ControlHandler := control.NewControlHandler(Joystick)

	PilotScene := NewPilotScene(renderer, ControlHandler)
	PilotScene.Init()

	//Проверочный показывать фпс, он же заглушка на систему каналов
	initFPS := fps.InitStruct{
		MIN_FRAME_MS:           DEFVAL.MIN_FRAME_MS,
		MIN_PHYS_MS:            DEFVAL.MIN_PHYS_MS,
		MAX_FRAME_MS:           DEFVAL.MAX_FRAME_MS,
		MAX_PHYS_MS:            DEFVAL.MAX_PHYS_MS,
		FPS_UPDATE_MS:          DEFVAL.FPS_UPDATE_MS,
		TickerBalancerOverhead: DEFVAL.TickerBalancerOverhead,
	}
	ShowFpsTick, fpsControl := fps.Start(initFPS)
	defer close(fpsControl)

	lastPhysFrame := time.Now()

	graphFrameN, physFrameN, ioFrameN := 0, 0, 0
	var maxDt, maxGraphT, maxPhysT float32

	breakMainLoop := make(chan bool, 1)

	IOTick := time.Tick(20 * time.Millisecond)

loop:
	for {
		select {
		//команда на выход
		case <-breakMainLoop:
			break loop
		//Время передать фпс
		case <-ShowFpsTick:
			fpsControl <- fps.FpsData{graphFrameN, physFrameN, ioFrameN,
				maxDt, maxGraphT, maxPhysT}
			maxDt = 0.0
			maxGraphT = 0.0
			maxPhysT = 0.0

		//ПРИОРИТЕТ 1: тик ФИЗИЧЕСКОГО движка
		case <-fps.PTick:
			deltaTime := float32(time.Since(lastPhysFrame).Seconds())
			if deltaTime > maxDt {
				maxDt = deltaTime
			}
			lastPhysFrame = time.Now()
			physFrameN++

			ControlHandler.BeforeUpdate()
			PilotScene.Update(deltaTime)
			T := float32(time.Since(lastPhysFrame).Seconds())
			if T > maxPhysT {
				maxPhysT = T
			}
		default:
			select {
			//ПРИОРИТЕТ 2: тик ГРАФИЧЕСКОГО движка
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

			//ПРИОРИТЕТ 3: снятие состояния УПРАВЛЕНИЯ
			case <-IOTick:
				ioFrameN++
				DoMainLoopIO(breakMainLoop, ControlHandler)
			}
		}
	}
}

func DoMainLoopIO(breakMainLoop chan bool, handler *control.Handler) {
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
		default:
		}
		handler.HandleSDLEvent(event)
	}
	handler.IOUpdate()
}

func timeCheck(caption string) func() {
	Start := time.Now()
	return func() {
		log.Println(caption, time.Since(Start).Seconds()*1000000, "micro s")
	}
}
