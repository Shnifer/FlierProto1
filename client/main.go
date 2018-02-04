package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"log"
	"time"
	"math/rand"
	"github.com/Shnifer/FlierProto1/MNT"
	"runtime"
)

//Константы экрана
//TODO:вынести параметры экрана во внешний файл конфигурации
var winW int32 = 800
var winH int32 = 600

const MIN_FRAME_MS = 20

const ResourcePath = "res/"

func ListenAndShowFPS() chan<- int {
	c:=make (chan int)
	lastFrameN:=0
	go func(){
		for frame:=range c{
			log.Println("FPS:",frame-lastFrameN)
			lastFrameN = frame
		}
	}()
	return c
}

type GameState byte
const(
	state_Login GameState = iota
	state_PilotSpace
	state_NaviSpace
)

func main() {

	runtime.LockOSThread()

	rand.Seed(time.Now().Unix())

	log.Println("Connecting to Server...")
	if err:=MNT.ConnectToServer();err!=nil{
		log.Panicln(err)
	}

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		log.Panicln(err)
	}
	defer sdl.Quit()

	var mode sdl.DisplayMode
	if err:=sdl.GetDesktopDisplayMode(0, &mode); err!=nil{
		log.Panic(err)
	}

	var winmode uint32 = sdl.WINDOW_SHOWN
	//Для полного экрана
	//winH = mode.H
	//winW = mode.W
	//winmode = sdl.WINDOW_FULLSCREEN

	window, err := sdl.CreateWindow("COSMO FLIER", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, winW, winH, winmode)
	if err != nil {
		log.Panicln(err)
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		log.Panicln(err)
	}
	defer renderer.Destroy()

	//Создаём кэш текстур В ГЛОБАЛЬНУЮ ПЕРЕМЕННУЮ
	TCache = newTexCache(renderer)

	/*
	Закрыто чтобы не свистело. Имеет смысл включать по необходимости
	if err := mix.OpenAudio(mix.DEFAULT_FREQUENCY, mix.DEFAULT_FORMAT, mix.DEFAULT_CHANNELS, mix.DEFAULT_CHUNKSIZE); err != nil {
		log.Panicln(err)
	}
	defer mix.CloseAudio()
	*/

	//Загрузка Аудио файла
	//chunk, err := mix.LoadWAV("res/explode.wav")
	//if err != nil {
	//	log.Panicln(err)
	//}

	//Параметр сглаживания массштабирования
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	log.Println("login")
	MNT.LoginToServer(MNT.RoomName,MNT.ROLE_PILOT)
	MNT.DownloadGalaxy()

	ControlHandler:=newControlHandler()

	PilotScene := NewPilotScene(renderer, ControlHandler)
	PilotScene.Init()
	//Проверочный показывать фпс, он же заглушка на систему каналов
	fpsTick:=time.Tick(1000*time.Millisecond)
	cfps:=ListenAndShowFPS()
	defer close(cfps)

	lastFrame := time.Now()

	frameN:=0
loop:
	for {
		frameN++
		//Отмеряем время кадра
		deltaTime := float32(time.Since(lastFrame).Seconds())
		_ = deltaTime
		lastFrame = time.Now()

		//Проверяем и хэндлим события СДЛ. Выход -- обязательно, а то не закроется
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch ev:=event.(type) {
			case *sdl.QuitEvent:
				log.Println("quit")
				break loop
			case *sdl.KeyboardEvent:
				//Кнопку выходи обрабатываем здесь, чтобы порвать главный цикл
				scan:=ev.Keysym.Scancode
				if scan==sdl.SCANCODE_ESCAPE {
					break loop
				}
				//Остальное в хэндлере
				ControlHandler.handleSDLKeyboardEvent(ev)
			default:
				//log.Printf("event: %T",ev)
			}
		}

		//Цикл обработки каналов и сообщений в главном треде,
		// выбирает все, поэтому должен быть быстрым
		selectloop:
		for {
			select{
				case <-fpsTick:
					cfps<-frameN
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

func timeCheck(caption string) func(){
	Start:=time.Now()
	return func(){
		log.Println(caption, time.Since(Start).Seconds()*1000, "ms")
	}
}