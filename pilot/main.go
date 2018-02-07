package main

import (
	"fmt"
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

type fpsData struct {
	graph, phys, io     int
	maxDt               float32
	maxGraphT, maxPhysT float32
}

type controlTickerData struct {
	newGraphPeriodms, newPhysPeriodms float32
}

func ListenAndShowFPS() (chan<- fpsData, <-chan controlTickerData) {
	const overhead = 1.2
	inData := make(chan fpsData)
	tickerControl := make(chan controlTickerData, 1)

	lastGraph, lastPhys, lastIO := 0, 0, 0

	go func() {
		defer close(tickerControl)
		for fps := range inData {
			log.Println(
				"Frame/s:", fps.graph-lastGraph,
				"Phys/s:", fps.phys-lastPhys,
				"io/s:", fps.io-lastIO,
				"max dt", fps.maxDt*1000, "ms",
				"maxGraph:", fps.maxGraphT*1000, "ms",
				"maxPhys:", fps.maxPhysT*1000, "ms")
			lastGraph = fps.graph
			lastPhys = fps.phys
			lastIO = fps.io

			newGraphPeriod := fps.maxGraphT * overhead * 1000
			if newGraphPeriod < float32(DEFVAL.MIN_FRAME_MS) {
				newGraphPeriod = float32(DEFVAL.MIN_FRAME_MS)
			}
			if newGraphPeriod < 1 {
				newGraphPeriod = 1
			}
			newPhysPeriod := fps.maxPhysT * overhead * 1000
			if newGraphPeriod < float32(DEFVAL.MIN_PHYS_MS) {
				newGraphPeriod = float32(DEFVAL.MIN_PHYS_MS)
			}
			if newPhysPeriod < 1 {
				newPhysPeriod = 1
			}

			log.Println(newGraphPeriod, newPhysPeriod)

			tickerControl <- controlTickerData{
				newGraphPeriodms: newGraphPeriod,
				newPhysPeriodms:  newPhysPeriod,
			}
		}
	}()
	return inData, tickerControl
}

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

	MIN_FRAME_MS := DEFVAL.MIN_FRAME_MS
	MAX_PHYS_MS := DEFVAL.MIN_PHYS_MS

	ControlHandler := newControlHandler(Joystick)

	PilotScene := NewPilotScene(renderer, ControlHandler)
	PilotScene.Init()
	//Проверочный показывать фпс, он же заглушка на систему каналов
	ShowfpsTick := time.Tick(1000 * time.Millisecond)
	fpsControl, tickerControl := ListenAndShowFPS()
	defer close(fpsControl)

	GraphTick := time.NewTicker(time.Duration(MIN_FRAME_MS) * time.Millisecond)
	PhysTick := time.NewTicker(time.Duration(MAX_PHYS_MS) * time.Millisecond)
	defer GraphTick.Stop()
	defer PhysTick.Stop()

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
		case <-ShowfpsTick:
			fpsControl <- fpsData{graphFrameN, physFrameN, ioFrameN,
				maxDt, maxGraphT, maxPhysT}
			maxDt = 0.0
			maxGraphT = 0.0
			maxPhysT = 0.0
		case tc := <-tickerControl:
			PhysTick.Stop()
			GraphTick.Stop()
			GraphTick = time.NewTicker(time.Duration(tc.newGraphPeriodms) * time.Millisecond)
			PhysTick = time.NewTicker(time.Duration(tc.newPhysPeriodms) * time.Millisecond)

		case <-PhysTick.C:
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
			case <-GraphTick.C:
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
