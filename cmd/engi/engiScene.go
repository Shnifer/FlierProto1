package main

import (
	"github.com/Shnifer/flierproto1/control"
	MNT "github.com/Shnifer/flierproto1/mnt"
	"github.com/Shnifer/flierproto1/scene"
	"github.com/veandco/go-sdl2/sdl"
	"log"
	"strconv"
	"fmt"
)

type EngiScene struct {
	*scene.Scene
	BSP *MNT.BaseShipParameters
	SSS *MNT.ShipSystemsState

	FuelStock   float32
	LifeStock   float32

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

const smallOff = 5
const ShowAtribN = 2

var AName = [...]string{"Горючее", "Жизнееобеспечение"}

func (S *EngiScene) Init() {
	const vrtH = 20
	//Скопируем из глобальной не разделяя указатель
	b := BSP
	S.BSP = &b

	S.FuelStock = S.BSP.FuelStock
	S.LifeStock = S.BSP.LifeStock

	for i := 0; i < MNT.SystemsCount; i++ {
		rect := VirtualRect(50*(i%2)+smallOff, (i/2)*(vrtH+smallOff)+smallOff, 40, vrtH)
		SSD := NewSystemStateDisplay(rect,"SSD"+strconv.Itoa(i),fmt.Sprintf("SYSTEM #%v %v",i,MNT.SName[i]),
			S.GetValState(i))
		S.SSDs = append(S.SSDs, SSD)
		S.AddObject(SSD)
	}

	for i := 0; i < ShowAtribN; i++ {
		rect := VirtualRect(50*(i%2)+smallOff, (MNT.SystemsCount/2+i/2)*(vrtH+smallOff)+smallOff, 40, vrtH)
		SSD := NewSystemStateDisplay(rect,"ATR"+strconv.Itoa(i),fmt.Sprintf("Параметр #%v %v",i,AName[i]),
			S.GetValAttr(i))
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
				default:
					log.Printf("ЧОТ НАЖАЛИ, ОНО ОТКЛИКНУЛОСЬ, А ЧТО НЕ ПОНЯТНА! %T\n", clicked)
				}
			}
		}
	}
}

func (S *EngiScene) GetValState(i int) func()float32 {
	return func()float32{
		return S.SSS.Systems[i]
	}
}

func (S *EngiScene) GetValAttr(i int) func()float32 {
	return func()float32{
		switch i {
			case 0:return S.FuelStock/S.BSP.FuelStock
			case 1:return S.LifeStock/S.BSP.LifeStock
		default:
			log.Panicln("GetValAttr: param #",i,"not supported")
		}
		return 0
	}
}