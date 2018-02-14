package main

import (
	"github.com/Shnifer/flierproto1/scene"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/Shnifer/flierproto1/control"
	"log"
)

type EngiScene struct {
	*scene.Scene
}

func NewEngiScene(r *sdl.Renderer, ch *control.Handler) *EngiScene {
	return &EngiScene{
		Scene:         scene.NewScene(r, ch, winW, winH),
	}
}

func (S *EngiScene) Init() {
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