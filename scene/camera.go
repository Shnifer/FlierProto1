package scene

import (
	V2 "github.com/Shnifer/flierproto1/v2"
	"github.com/veandco/go-sdl2/sdl"
)

//СОГЛАШЕНИЯ ПО КООРДИНАТАМ И УГЛАМ:
//в мировых координатах Y вверх, X вправо, углы против часовой, 0 угол -- вверх

type Сamera struct {
	//ЦентрКамеры в мировых координатах
	CameraCenter V2.V2
	//Масштаб пикселей/единицу мировых координат,
	//увеличение текстур там, где это надо
	CameraScale float32
	CameraAngle float32

	//Размер поля камеры в пикселях экрана
	CamW, CamH int32
}

func newCamera(CamW, CamH int32, CameraScale float32) Сamera {
	return Сamera{
		CameraScale: CameraScale,
		CamW:        CamW,
		CamH:        CamH,
	}
}

//Прямоугольник во float32, для реальных координат
type f32Rect struct {
	//Храним как центр и полуоси
	Center V2.V2
	HW, HH float32

	//НЕ обрабатывается при расчете целевых координат, попадание в экран и т.д.
	//Служит для переноса угла между преобразовании
	//при отрисовки вынуть и передать в поле CopyEx
	//angle float32
}

//Создаёт новый прямоугольник по центру и полуосям
func NewF32Rect(center V2.V2, hW, hH float32) f32Rect {
	return f32Rect{
		Center: center,
		HW:     hW,
		HH:     hH}
}

//Создаёт новый квадрат по центру и радиусу
func NewF32Sqr(center V2.V2, rad float32) f32Rect {
	return f32Rect{
		Center: center,
		HW:     rad,
		HH:     rad}
}

//Преобразует координаты из реавльного вектора в координаты камеры сцены
func (cam Сamera) CameraTransformV2(v V2.V2) (x, y int32) {
	w := V2.Sub(v, cam.CameraCenter).Rotate(cam.CameraAngle).Mul(cam.CameraScale)
	x = cam.CamW/2 + int32(w.X)
	y = cam.CamH/2 - int32(w.Y)
	return
}

//Преобразует координаты из координат экрана в реальный вектор
func (cam Сamera) CameraScrTransformV2(x, y int32) V2.V2 {
	x = x - cam.CamW/2
	y = cam.CamH/2 - y
	V := V2.V2{float32(x), float32(y)}.Mul(1 / cam.CameraScale).Rotate(-cam.CameraAngle).Add(cam.CameraCenter)
	return V
}

func (cam Сamera) InCamera(rect sdl.Rect) bool {
	return !(rect.X+rect.W < 0 || rect.X > cam.CamW || rect.Y > cam.CamH || rect.Y+rect.H < 0)
}

//Преобразует прямоугольник из реальных координат в sdl.Rect
//определяет, лежит ли полученный Rect в границах экрана камеры
func (cam Сamera) CameraTransformRect(r f32Rect) (camRect *sdl.Rect, inCamera bool) {
	//TODO: понять нужен ли воможный пивот не по центру, пока считаем от центра

	//Центральный пиксель после поворота
	x, y := cam.CameraTransformV2(r.Center)
	//полувысота и ширина в экранных координатах после скейла
	hW := int32(r.HW * cam.CameraScale)
	hH := int32(r.HH * cam.CameraScale)

	res := sdl.Rect{
		x - hW,
		y - hH,
		hW * 2,
		hH * 2}

	inCamera = cam.InCamera(res)

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

	inCamera = cam.InCamera(res)

	return &res, inCamera
}

func (cam Scene) CameraRectByCenterAndScreenWH(center V2.V2, W,H int32) (camRect *sdl.Rect, inCamera bool) {
	x, y := cam.CameraTransformV2(center)
	res := sdl.Rect{
		x - W/2,
		y - H/2,
		W,
		H}

	inCamera = cam.InCamera(res)

	return &res, inCamera
}