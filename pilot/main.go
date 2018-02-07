package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"log"
	"runtime"
	"time"
	"fmt"
	MNT "github.com/Shnifer/flierproto1/mnt"
)

//Константы экрана
//TODO:вынести параметры экрана во внешний файл конфигурации
var winW int32
var winH int32

const ResourcePath = "res/"
const ClientDataPath = ResourcePath + "pilot/"

func ListenAndShowFPS() chan<- int {
	c := make(chan int)
	lastFrameN := 0
	go func() {
		for frame := range c {
			log.Println("FPS:", frame-lastFrameN)
			lastFrameN = frame
		}
	}()
	return c
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

	fmt.Println("got",renderer,Joystick)

	MIN_FRAME_MS := DEFVAL.MIN_FRAME_MS

	ControlHandler := newControlHandler(Joystick)

	PilotScene := NewPilotScene(renderer, ControlHandler)
	PilotScene.Init()
	//Проверочный показывать фпс, он же заглушка на систему каналов
	fpsTick := time.Tick(1000 * time.Millisecond)
	cfps := ListenAndShowFPS()
	defer close(cfps)

	lastFrame := time.Now()

	frameN := 0
loop:
	for {
		frameN++
		//Отмеряем время кадра
		deltaTime := float32(time.Since(lastFrame).Seconds())
		_ = deltaTime
		lastFrame = time.Now()

		//Проверяем и хэндлим события СДЛ. Выход -- обязательно, а то не закроется
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch ev := event.(type) {
			case *sdl.QuitEvent:
				log.Println("quit")
				break loop
			case *sdl.KeyboardEvent:
				//Кнопку выходи обрабатываем здесь, чтобы порвать главный цикл
				scan := ev.Keysym.Scancode
				if scan == sdl.SCANCODE_ESCAPE {
					break loop
				}
				//Остальное в хэндлере
				ControlHandler.handleSDLKeyboardEvent(ev)
			case *sdl.MouseMotionEvent:
			case *sdl.JoyButtonEvent:
				ControlHandler.handleJoyButtonEvent(ev)
			default:

			}
		}
		ControlHandler.updateJoystickAxis()

		//Цикл обработки каналов и сообщений в главном треде,
		// выбирает все, поэтому должен быть быстрым
	selectloop:
		for {
			select {
			case <-fpsTick:
				cfps <- frameN
			default:
				break selectloop
			}
		}

		//Опрашиваем состояние мыши (в перспективе и других контроллеров)
		//currMouseState := getMouseState()

		//Цикл апдейта всех компонент
		//for _, b := range Balloons {
		//	b.update(deltaTime,currMouseState,prevMouseState,chunk)
		//}

		//Сортировка и прочая проверка
		//Коллизии?
		//sort.Stable(Balloons)

		//TODO: разделить запуск апдейта и прорисовки по тикерам разной частоты
		PilotScene.Update(deltaTime)

		//Цикл рендера всех компонент
		renderer.Clear()

		//Показ Кадра
		PilotScene.Draw()

		renderer.Present()

		//Досыпка времени до минимума
		elapsed := uint32(time.Since(lastFrame).Seconds() * 1000)
		//log.Println("ms pr frame: ", elapsed)
		if elapsed < MIN_FRAME_MS {
			sdl.Delay(MIN_FRAME_MS - elapsed)
		}

		//обновления состояния мыши
		//prevMouseState = currMouseState
	}
}

func timeCheck(caption string) func() {
	Start := time.Now()
	return func() {
		log.Println(caption, time.Since(Start).Seconds()*1000, "ms")
	}
}