package main

import (
	"fmt"
	MNT "github.com/Shnifer/flierproto1/mnt"
	"github.com/veandco/go-sdl2/sdl"
	"log"
	"math/rand"
	"time"
)

//Инициализация SDL, загрузка файлов среды, установление сетевого соединения, загрузка галактики
func InitSomeShit() (deferMe func(), r *sdl.Renderer, j *sdl.Joystick) {
	rand.Seed(time.Now().Unix())
	LoadDefVals(ClientDataPath)
	log.Println("Connecting to Server...")
	if err := MNT.ConnectClientToServer(DEFVAL.ServerName, DEFVAL.tcpPort); err != nil {
		log.Panicln(err)
	}

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		log.Panicln(err)
	}

	var mode sdl.DisplayMode
	if err := sdl.GetDesktopDisplayMode(0, &mode); err != nil {
		log.Panic(err)
	}

	var winmode uint32 = sdl.WINDOW_SHOWN
	//Для полного экрана
	if DEFVAL.FullScreen {
		winH = mode.H
		winW = mode.W
		winmode = sdl.WINDOW_FULLSCREEN
	} else {
		winH = DEFVAL.WinH
		winW = DEFVAL.WinW
	}

	window, err := sdl.CreateWindow("COSMO FLIER", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, winW, winH, winmode)
	if err != nil {
		log.Panicln(err)
	}

	var ACCELERATED uint32
	if DEFVAL.RENDERER_ACCELERATED {
		ACCELERATED = sdl.RENDERER_ACCELERATED
	} else {
		ACCELERATED = sdl.RENDERER_SOFTWARE
	}
	renderer, err := sdl.CreateRenderer(window, -1, ACCELERATED)
	if err != nil {
		log.Panicln(err)
	}
	//Параметр сглаживания массштабирования
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	//Создаём кэш текстур В ГЛОБАЛЬНУЮ ПЕРЕМЕННУЮ
	TCache = newTexCache(renderer)

	//Joystick 0 initialize
	var Joystick *sdl.Joystick
	if sdl.NumJoysticks() > 0 {
		Joystick = sdl.JoystickOpen(sdl.JoystickID(0))
		log.Println("Joystick detected")
	} else {
		log.Println("Nojoystick")
	}
	deferMe = func() {
		if Joystick != nil {
			Joystick.Close()
		}
		renderer.Destroy()
		window.Destroy()
		sdl.Quit()
	}

	log.Println("login")
	MNT.LoginToServer(MNT.RoomName, MNT.ROLE_PILOT)
	MNT.DownloadGalaxy()

	fmt.Println("returned", renderer, Joystick)
	return deferMe, renderer, Joystick
}
