package main

import (
	"github.com/Shnifer/flierproto1/control"
	"github.com/Shnifer/flierproto1/fps"
	MNT "github.com/Shnifer/flierproto1/mnt"
	"github.com/veandco/go-sdl2/sdl"
	"log"
	"runtime"
	"time"
)

//Константы экрана
var winW int32
var winH int32

const ResourcePath = "res/"

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

	EngiScene := NewEngiScene(renderer, ControlHandler)
	EngiScene.Init()

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

	graphFrameN, physFrameN, ioFrameN, netFrameN := 0, 0, 0, 0
	var maxDt, maxGraphT, maxPhysT float32

	breakMainLoop := make(chan bool, 1)

	IOTick := time.Tick(15 * time.Millisecond)
	NetTick := time.Tick(20 * time.Millisecond)

	//считаем сами для показа
//	lastFrame := 0
loop:
	for {
		select {
		//команда на выход
		case <-breakMainLoop:
			break loop
			//Время передать фпс
		case <-ShowFpsTick:
			fpsControl <- fps.FpsData{graphFrameN, physFrameN, ioFrameN, netFrameN,
				maxDt, maxGraphT, maxPhysT}
//			NaviScene.showFps(strconv.Itoa((graphFrameN - lastFrame) * 1000 / DEFVAL.FPS_UPDATE_MS))
//			lastFrame = graphFrameN
			maxDt = 0.0
			maxGraphT = 0.0
			maxPhysT = 0.0

			//ПРИОРИТЕТ 1: тик ФИЗИЧЕСКОГО движка
		case <-fps.PTick:
			//МЫ ВЕДОМЫЕ, пока не олучили первое ненулевое значение из вне -- не трогаемся.
			//это же флаг паузы показа
			deltaTime := float32(time.Since(lastPhysFrame).Seconds())
			if deltaTime > maxDt {
				maxDt = deltaTime
			}
			lastPhysFrame = time.Now()
			physFrameN++

			ControlHandler.BeforeUpdate()
			EngiScene.Update(deltaTime)
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
				EngiScene.Draw()
				renderer.Present()
				T := float32(time.Since(start).Seconds())
				if T > maxGraphT {
					maxGraphT = T
				}

				//ПРИОРИТЕТ 3: снятие состояния УПРАВЛЕНИЯ
			case <-IOTick:
				ioFrameN++
				DoMainLoopIO(breakMainLoop, ControlHandler, EngiScene)
				//ПРИОРИТЕТ 3: обновление состояния СЕТИ
			case <-NetTick:
				netFrameN++
				DoMainLoopNet(EngiScene)
			}
		}

	}
}

func DoMainLoopIO(breakMainLoop chan bool, handler *control.Handler, EngiScene *EngiScene) {
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
	//По сути обрабатываем OnClick одним обработчиком Сцены
	clicks := handler.TakeMouseClicks()
	EngiScene.UpdateClicks(clicks)
}

func DoMainLoopNet(scene *EngiScene) {
loop:
	for {
		//Слушаем пока канал готов сразу отдать
		var msg string
		select {
		case v := <-MNT.Client.Recv:
			msg = v
		default:
			break loop
		}

		cmd, param := MNT.SplitMsg(msg)
		if cmd == MNT.IN_MSG {
			msgtype, param := MNT.SplitMsg(param)
			_ = param
			switch msgtype {
			/*

		case MNT.SHIP_POS:
			data, err := MNT.DecodeShipPos(param)
			if err != nil {
				log.Panicln(err)
			}
			ProcShipData(scene, data)
		case MNT.SESSION_TIME:
			t, err := strconv.ParseFloat(param, 32)
			if err != nil {
				log.Panicln(err)
			}
			scene.NetSyncTime = float32(t)
			}
		*/
			}
		}
	}
}


func timeCheck(caption string) func() {
	Start := time.Now()
	return func() {
		log.Println(caption, time.Since(Start).Seconds()*1000000, "micro s")
	}
}