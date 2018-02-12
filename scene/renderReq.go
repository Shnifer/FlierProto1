package scene

import "github.com/veandco/go-sdl2/sdl"

//Слои вывода объектов на рендер
type ZLayer int

const (
	//Статический фон, обычно один и занимает весь экран
	Z_STAT_BACKGROUND ZLayer = iota * 100
	//Динамические изменения на фоне, например координатная сетка
	Z_BACKGROUND
	//Подложка под игровым объектом, например кружок выделения
	Z_UNDER_OBJECT
	//Сами игровые объекты
	Z_GAME_OBJECT
	//Сверху игровых объектов, надписи или гизмосы объектов
	Z_ABOVE_OBJECT
	//не привязанное к координатам игрового мира, например системы управления
	Z_HUD
	//обычно одна картинка с прозрачным центром отрисовывающая красивые края
	Z_STAT_HUD
)

type RenderReq interface {
	GetZ() ZLayer
}

type RenderCopyReq struct {
	tex       *sdl.Texture
	src, dest *sdl.Rect
	z         ZLayer
	angle     float64
	pivot     *sdl.Point
	flip      sdl.RendererFlip
}

func (r RenderCopyReq) GetZ() ZLayer {
	return r.z
}

type RenderDrawLinesReq struct {
	color  sdl.Color
	points []sdl.Point
	z      ZLayer
}

func (r RenderDrawLinesReq) GetZ() ZLayer {
	return r.z
}

type RenderFilledCircleReq struct {
	x, y, rad int32
	color     sdl.Color
	z         ZLayer
}

func (r RenderFilledCircleReq) GetZ() ZLayer {
	return r.z
}

type RenderFilledPieReq struct {
	x, y       int32
	rad, inrad int32
	start, end int32
	color      sdl.Color
	z          ZLayer
}

func (r RenderFilledPieReq) GetZ() ZLayer {
	return r.z
}

func NewRenderReq(tex *sdl.Texture, src, dest *sdl.Rect, z ZLayer, angle float64, pivot *sdl.Point, flip sdl.RendererFlip) RenderCopyReq {
	return RenderCopyReq{
		tex:   tex,
		src:   src,
		dest:  dest,
		z:     z,
		angle: angle,
		pivot: pivot,
		flip:  flip,
	}
}

func NewRenderReqSimple(tex *sdl.Texture, src, dest *sdl.Rect, z ZLayer) RenderCopyReq {
	return RenderCopyReq{
		tex:   tex,
		src:   src,
		dest:  dest,
		z:     z,
		angle: 0,
		pivot: nil,
		flip:  sdl.FLIP_NONE,
	}
}

func NewRenderDrawLinesReq(points []sdl.Point, color sdl.Color, z ZLayer) RenderDrawLinesReq {
	return RenderDrawLinesReq{
		points: points,
		color:  color,
		z:      z,
	}
}

func NewFilledCircleReq(x, y, rad int32, color sdl.Color, z ZLayer) RenderFilledCircleReq {
	return RenderFilledCircleReq{
		x:     x,
		y:     y,
		rad:   rad,
		color: color,
		z:     z,
	}
}

func NewFilledPieReq(x, y, rad, inrad, start, end int32, color sdl.Color, z ZLayer) RenderFilledPieReq {
	return RenderFilledPieReq{
		x:     x,
		y:     y,
		rad:   rad,
		inrad: inrad,
		start: start,
		end:   end,
		color: color,
		z:     z,
	}
}

type RenderReqList []RenderReq

func (r RenderReqList) Len() int {
	return len(r)
}
func (r RenderReqList) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r RenderReqList) Less(i, j int) bool {
	return r[i].GetZ() < r[j].GetZ()
}
