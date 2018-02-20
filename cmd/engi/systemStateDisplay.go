package main

import (
	"fmt"
	"github.com/Shnifer/flierproto1/scene"
	"github.com/Shnifer/flierproto1/texture"
	"github.com/veandco/go-sdl2/sdl"
)

type SystemStateDisplay struct {
	*scene.TextUI
	Caption *scene.TextUI
	Rect    sdl.Rect

	getVal func() float32
	//Если 0 - показываем getVal как процент,
	//если больше нуля, как долю getVal/baseVal
	baseVal float32
}

func (ssd *SystemStateDisplay) IsClicked(x, y int32) bool {
	p := sdl.Point{x, y}
	return p.InRect(&ssd.Rect)
}

func NewSystemStateDisplay(Rect sdl.Rect, id string, caption string, getVal func() float32, baseVal float32) *SystemStateDisplay {
	f := texture.Cache.GetFont(DEFVAL.SSDFontName, DEFVAL.SSDFontSize)

	Caption := scene.NewTextUI(caption, f, scene.WHITE, scene.Z_BACKGROUND, scene.FROM_ANGLE)
	res := SystemStateDisplay{
		TextUI:  scene.NewTextUI("", f, scene.WHITE, scene.Z_HUD, scene.FROM_CENTER),
		Rect:    Rect,
		Caption: Caption,
		getVal:  getVal,
		baseVal: baseVal,
	}

	res.Caption.X = res.Rect.X + smallOff
	res.Caption.Y = res.Rect.Y + smallOff
	res.Caption.SetID(id + "~caption")

	res.TextUI.X = res.Rect.X + res.Rect.W/2
	res.TextUI.Y = res.Rect.Y + res.Rect.H/2
	res.TextUI.SetID(id)

	return &res
}

func (ssd *SystemStateDisplay) Init(s *scene.BScene) {
	s.AddObject(ssd.Caption)
	ssd.Caption.Init(s)
	ssd.TextUI.Init(s)
}

func (ssd *SystemStateDisplay) Update(dt float32) {
	ssd.TextUI.Update(dt)
}

func (ssd *SystemStateDisplay) Draw(r *sdl.Renderer) (res scene.RenderReqList) {
	rectreq := scene.NewRectReq(ssd.Rect, scene.RED, scene.Z_BACKGROUND)
	res = append(res, rectreq)

	Val := ssd.getVal()
	if ssd.baseVal == 0 {
		ssd.TextUI.ChangeText(fmt.Sprintf("%.0f%%", (Val * 100)))
	} else {
		ssd.TextUI.ChangeText(fmt.Sprintf("%v%% ( %v / %v )", int(Val/ssd.baseVal*100), Val, ssd.baseVal))
	}
	ress := ssd.TextUI.Draw(r)
	res = append(res, ress...)
	return res
}

func (ssd *SystemStateDisplay) GetID() string {
	return ssd.TextUI.GetID()
}
