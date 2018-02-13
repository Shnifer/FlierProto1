package main

import (
	"github.com/Shnifer/flierproto1/scene"
	"github.com/Shnifer/flierproto1/texture"
	"github.com/Shnifer/flierproto1/v2"
	"github.com/veandco/go-sdl2/sdl"
	"log"
	"sync"

	"math"
)

type metaballBase struct {
	star   *StarGameObject
	weight float32
}

//База будет вынесена в интерфейс, так что получения позиции - метод
func (b *metaballBase) Pos() V2.V2 {
	return b.star.Pos
}

//База отчитывается от максимальном радиусе, в котором может быть результат
//Из квадратов ограничивающих базы будут построены границы небулы
func (b metaballBase) maxR() float32 {
	return b.weight
}

//Расстояние (кв) на котором гарантируется область вокруг точки
func (b metaballBase) garantR2(totalW float32) float32 {
	return (1 - float32(math.Sqrt(float64(totalW)))) * b.weight * b.weight
}

//База возвращает потенциал для точки
//Должна работать для любой точки, сама отслеживая ограничения если нужно
func (b metaballBase) calcW(point V2.V2) float32 {
	v := b.Pos().Sub(point)
	if abs(v.X) > b.weight || abs(v.Y) > b.weight {
		return 0
	}
	dist2 := v.LenSqr() / b.weight / b.weight
	x := 1 - dist2
	if x <= 0 {
		return 0
	}
	return (x * x)
}

type Nebula struct {
	base   []metaballBase
	totalW float32

	effect string

	scene *scene.Scene
	tex   *sdl.Texture
}

func NewNebula(stars []*StarGameObject) *Nebula {
	res := Nebula{effect: "Nebula effect string", totalW: 0.7}
	for _, star := range stars {
		res.base = append(res.base, metaballBase{star: star, weight: 250})
	}
	return &res
}

func (n *Nebula) Init(s *scene.Scene) {
	n.scene = s
}

func (n *Nebula) Update(dt float32) {

}

func (n Nebula) isInside(point V2.V2) bool {
	var sum float32
	for _, base := range n.base {
		sum += base.calcW(point)
		if sum > n.totalW {
			return true
		}
	}
	return false
}

func (n *Nebula) Draw(r *sdl.Renderer) (res scene.RenderReqList) {
	//stop := timeCheck("nebuladraw")
	//defer stop()

	baseRects, totalRect, pixels := n.calcPixels()
	if len(pixels) == 0 {
		//Ничего нет, расходимся
		return
	}

	if n.tex != nil {
		n.tex.Destroy()
	}

	basesMaxReq := scene.NewRectsReq(baseRects, scene.WHITE, scene.Z_HUD)
	totalReq := scene.NewRectReq(totalRect, scene.RED, scene.Z_HUD)

	tex, err := texture.PixelsToTexture(n.scene.R, pixels, int(totalRect.W), int(totalRect.H))
	if err != nil {
		log.Panicln(err)
	}
	n.tex = tex
	req := scene.NewRenderReqSimple(n.tex, nil, &totalRect, scene.Z_UNDER_OBJECT)
	res = append(res, basesMaxReq, totalReq, req)
	return res
}

func (n *Nebula) calcPixels() (bases []sdl.Rect, total sdl.Rect, pixels []byte) {
	//var bases []sdl.Rect
	//var total sdl.Recк
	baseIndNebulaInd := make(map[int]int)
	for nInd, base := range n.base {

		fsqr := scene.NewF32Sqr(base.Pos(), base.maxR())

		camrect, inCamera := n.scene.CameraTransformRect(fsqr)
		if inCamera {
			bases = append(bases, *camrect)
			total = total.Union(camrect)
			baseIndNebulaInd[len(bases)-1] = nInd
		}
	}

	if total.Empty() {
		//Вся небула не попадает в экран -- больше не считаем
		return bases, total, []byte{}
	}

	//TODO: распараллелить!
	pixels = make([]byte, total.W*total.H*4)
	wg := sync.WaitGroup{}
	tX, tY := total.X, total.Y //угол total в экранных координатах
	tW := total.W
	const maxInTimeN = 3
	wcontrol := make(chan bool, maxInTimeN)
	log.Println(len(bases))
	for baseInd, base := range bases {
		wg.Add(1)
		go func(baseInd int, base sdl.Rect) {
			wcontrol <- true
			//======
			garantR2 := n.base[baseIndNebulaInd[baseInd]].garantR2(n.totalW)
			bpoint:=n.base[baseIndNebulaInd[baseInd]].Pos()
			var intRects []sdl.Rect
			var myInd int
			for sInd:=0;sInd<len(bases);sInd++{
				if base.HasIntersection(&bases[sInd]) {
					intRects = append(intRects,bases[sInd])
					if sInd==baseInd{
						myInd = len(intRects)-1
					}
				}
			}
			bX, bY := base.X-tX, base.Y-tY //угол base относительно начала total для расчёта индекса
			for y := int32(0); y < base.W; y++ {
			loop:
				for x := int32(0); x < base.H; x++ {
					//Вдруг точка есть в более ранних прямоугольниках?
					aX, aY := x+base.X, y+base.Y //координаты точки в экранных координатах
					isMoreRects:=false
					for r := 0; r < len(intRects); r++ {
						if r == myInd {
							continue
						}
						cr := &intRects[r]
						if cr.X <= aX && cr.Y <= aY &&
							cr.X+cr.W > aX && cr.Y+cr.H > aY {
							if r < myInd {
								continue loop
							} else {
								isMoreRects = true
								break
							}
						}
					}

					point:=n.scene.CameraScrTransformV2(aX, aY)
					D2:=(point.X-bpoint.X)*(point.X-bpoint.X)+(point.Y-bpoint.Y)*(point.Y-bpoint.Y)
					draw:=false
					if D2<=garantR2 {
						draw=true
					} else if !isMoreRects && garantR2>0{
						continue
					} else {
						draw=n.isInside(point)
					}

					if draw {
						//ПИШЕМ БЕЗ МУТЕКСОВ, как пидоры!
						ind := (bY+y)*(tW*4) + (bX+x)*4
						pixels[ind+0] = 255
						pixels[ind+1] = 255
						pixels[ind+2] = 255
						pixels[ind+3] = 255

					}
				}
			}
			<-wcontrol
			wg.Done()
			//======
		}(baseInd, base)
	}
	wg.Wait()
	return bases, total, pixels
}

func (n Nebula) GetID() string {
	return ""
}
