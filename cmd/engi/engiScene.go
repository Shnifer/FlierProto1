package main

import (
	"fmt"
	"github.com/Shnifer/flierproto1/control"
	MNT "github.com/Shnifer/flierproto1/mnt"
	"github.com/Shnifer/flierproto1/scene"
	"github.com/veandco/go-sdl2/sdl"
	"log"
)

type EngiScene struct {
	*scene.Scene
	BSP *MNT.BaseShipParameters
	SSS MNT.ShipSystemsState

	SSDs []*SystemStateDisplay
}

func NewEngiScene(r *sdl.Renderer, ch *control.Handler) *EngiScene {
	return &EngiScene{
		Scene: scene.NewScene(r, ch, winW, winH),
		SSS:   MNT.NewShipSystemsState(),
		SSDs:  make([]*SystemStateDisplay, 0),
	}
}

//Переводит виртуальные координаты "проценты экрана" в экранный прямоугольник
func VirtualRect(x, y, w, h int) sdl.Rect {
	return sdl.Rect{
		X: winW * int32(x) / 100,
		Y: winH * int32(y) / 100,
		W: winW * int32(w) / 100,
		H: winH * int32(h) / 100,
	}
}

const smallOff = 3

func (S *EngiScene) Init() {
	const vrtH = 15
	//Скопируем из глобальной не разделяя указатель
	b := BSP
	S.BSP = &b

	S.SSS[MNT.PFuelStock] = S.BSP.FuelStock
	S.SSS[MNT.PLifeStock] = S.BSP.LifeStock

	for i := 0; i < MNT.SystemsCount; i++ {
		rect := VirtualRect(50*(i%2)+smallOff, (i/2)*(vrtH+smallOff)+smallOff, 40, vrtH)
		name := MNT.SNames[i]
		SSD := NewSystemStateDisplay(rect, name, fmt.Sprintf("SYSTEM #%v %v", i, MNT.StrSName[i]),
			S.GetValState(name), 0)
		S.SSDs = append(S.SSDs, SSD)
		S.AddObject(SSD)
	}

	for i := 0; i < MNT.ParamCount; i++ {
		rect := VirtualRect(50*(i%2)+smallOff, (MNT.SystemsCount/2+i/2)*(vrtH+smallOff)+smallOff, 40, vrtH)
		name := MNT.PNames[i]
		SSD := NewSystemStateDisplay(rect, name, fmt.Sprintf("Параметр #%v %v", i, MNT.StrPName[i]),
			S.GetValState(name), S.getBaseVal(name))
		S.SSDs = append(S.SSDs, SSD)
		S.AddObject(SSD)
	}

	background := scene.NewStaticImage("engiBackground.jpg", scene.Z_STAT_BACKGROUND)
	S.AddObject(background)

	S.Scene.Init()
}

func (S *EngiScene) Update(dt float32) {
	S.Scene.Update(dt)
}

func (S *EngiScene) Draw() {
	S.Scene.Draw()
}

//Обрабатываем по частоте IOtick~50 в секунду все события кликов мышки
func (s *EngiScene) UpdateClicks(clicks []*control.MouseClick) {
	for _, click := range clicks {
		//здесь маскируем клики в HUD и прочие скрытые элементы
		//Либо собираем всех, кто откликнулся на факт клика и анализируем
		for _, obj := range s.Objects {
			Clickable, ok := obj.(scene.Clickable)
			if !ok {
				continue
			}
			if Clickable.IsClicked(click.X, click.Y) {
				switch clicked := obj.(type) {
				case *SystemStateDisplay:
					name := clicked.GetID()
					switch click.But {
					case sdl.BUTTON_RIGHT:
						if MNT.IsSystemName(name) {
							s.BreakSystem(name, 0.1)
						}
					case sdl.BUTTON_LEFT:
						if MNT.IsSystemName(name) {
							s.RepairSystem(name, 0.1)
						}
					}
				default:
					log.Printf("ЧОТ НАЖАЛИ, ОНО ОТКЛИКНУЛОСЬ, А ЧТО НЕ ПОНЯТНА! %T\n", clicked)
				}
			}
		}
	}
}

func (S *EngiScene) GetValState(name string) func() float32 {
	return func() float32 {
		return S.SSS[name]
	}
}

func (S *EngiScene) getBaseVal(name string) float32 {
	switch name {
	case MNT.PFuelStock:
		return S.BSP.FuelStock
	case MNT.PLifeStock:
		return S.BSP.LifeStock
	}
	log.Panicln("getBaseVal WRONG name")
	return 0
}

func (S *EngiScene) BreakSystem(name string, num float32) {
	v := S.SSS[name]
	v = v - num
	if v < 0 {
		v = 0
	}
	S.SSS[name] = v
}

func (S *EngiScene) RepairSystem(name string, num float32) {
	v := S.SSS[name]
	v = v + num
	if v > 1 {
		v = 1
	}
	S.SSS[name] = v
}
