package main

import (
	V2 "github.com/Shnifer/flierproto1/v2"
	"github.com/veandco/go-sdl2/sdl"
)

//СОГЛАШЕНИЯ ПО КООРДИНАТАМ И УГЛАМ:
//в мировых координатах Y вверх, X вправо, углы против часовой, 0 угол -- вверх

type Camera struct {
	//ЦентрКамеры в мировых координатах
	CameraCenter V2.V2
	//Масштаб пикселей/единицу мировых координат,
	//увеличение текстур там, где это надо
	CameraScale float32
	CameraAngle float32
}

func newCamera() Camera {
	return Camera{CameraScale: 10}
}

//Прямоугольник во float32, для реальных координат
type f32Rect struct {
	//Храним как центр и полуоси
	center V2.V2
	hW, hH float32

	//НЕ обрабатывается при расчете целевых координат, попадание в экран и т.д.
	//Служит для переноса угла между преобразовании
	//при отрисовки вынуть и передать в поле CopyEx
	//angle float32
}

//Создаёт новый прямоугольник по центру и полуосям
func newF32Rect(center V2.V2, hW, hH float32) f32Rect {
	return f32Rect{
		center: center,
		hW:     hW,
		hH:     hH}
}

//Создаёт новый квадрат по центру и радиусу
func newF32Sqr(center V2.V2, rad float32) f32Rect {
	return f32Rect{
		center: center,
		hW:     rad,
		hH:     rad}
}

//Преобразует координаты из реавльного вектора в координаты камеры сцены
func (cam Camera) CameraTransformV2(v V2.V2) (x, y int32) {
	//TODO: положение камеры не по центру экрана
	w := V2.Sub(v, cam.CameraCenter).Rotate(cam.CameraAngle).Mul(cam.CameraScale)
	x = winW/2 + int32(w.X)
	y = winH/2 - int32(w.Y)
	return
}

func (cam Camera) inCamera(rect sdl.Rect) bool {
	return !(rect.X+rect.W < 0 || rect.X > winW || rect.Y > winH || rect.Y+rect.H < 0)
}

//Преобразует прямоугольник из реальных координат в sdl.Rect
//определяет, лежит ли полученный Rect в границах экрана камеры
func (cam Camera) CameraTransformRect(r f32Rect) (camRect *sdl.Rect, inCamera bool) {
	//TODO: понять нужен ли воможный пивот не по центру, пока считаем от центра

	//Центральный пиксель после поворота
	x, y := cam.CameraTransformV2(r.center)
	//полувысота и ширина в экранных координатах после скейла
	hW := int32(r.hW * cam.CameraScale)
	hH := int32(r.hH * cam.CameraScale)

	res := sdl.Rect{
		x - hW,
		y - hH,
		hW * 2,
		hH * 2}

	inCamera = cam.inCamera(res)

	return &res, inCamera
}

//Экранный прямоугольник для заданной физической координаты центра И ЭКРАННОГО РАЗМЕРА
func (cam Scene) CameraRectByCenterAndScreenSize(center V2.V2, halfsize int32) (camRect *sdl.Rect, inCamera bool) {
	x, y := cam.CameraTransformV2(center)
	res := sdl.Rect{
		x - halfsize,
		y - halfsize,
		halfsize * 2,
		halfsize * 2}

	inCamera = cam.inCamera(res)

	return &res, inCamera
}
