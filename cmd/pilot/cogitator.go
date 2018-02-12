package main

import (
	"github.com/Shnifer/flierproto1/control"
	"github.com/Shnifer/flierproto1/scene"
	"github.com/Shnifer/flierproto1/v2"
	"github.com/veandco/go-sdl2/sdl"
	"math"
)

type PlayerInputs struct {
	AxisX, AxisY float32
	But1         bool
}

type CogitatorOutput struct {
	wantedMainThrust float32
	wantedAngThrust  float32
}

type ShipStats struct {
	angle, angleSpeed, maxAngMomentum, mainThrust float32
}

//В массштабе ввода: нажатая кнопка -1,1, ось - (-1,+1)
type Cogitator struct {
	Inputs PlayerInputs
	Stats  ShipStats

	wantedAngle      float32
	wantedMainThrust float32
}

const wantedAngleSpeed = 180
const wantedThrustSpeed = 2

func (c *Cogitator) Cogitate(dt float32) (CO CogitatorOutput) {
	c.wantedAngle += c.Inputs.AxisX * wantedAngleSpeed * dt
	c.wantedAngle = angClamp(c.wantedAngle)

	dAng := angSub(c.wantedAngle, c.Stats.angle)
	if abs(dAng) < 0.1 {
		dAng = 0
	}
	CO.wantedAngThrust = dAng - c.Stats.angleSpeed/5

	c.wantedMainThrust += c.Inputs.AxisY * wantedThrustSpeed * dt
	if c.wantedMainThrust > 1 {
		c.wantedMainThrust = 1
	} else if c.wantedMainThrust < 0 {
		c.wantedMainThrust = 0
	}
	CO.wantedMainThrust = c.wantedMainThrust
	if c.Inputs.But1 {
		CO.wantedMainThrust = 1
	}
	return CO
}

func (c *Cogitator) GetShipStates(ship *PlayerShipGameObject) {
	c.Stats.angle = ship.angle
	c.Stats.angleSpeed = ship.angleSpeed
	c.Stats.maxAngMomentum = ship.maxAngMomentum
	c.Stats.mainThrust = ship.mainThrust
}

//Разбираем управление, Здесь же перемапливаем клавиши разных джойстиков
func (c *Cogitator) GetInputs(CH *control.Handler) {

	var Inputs PlayerInputs
	if CH.HasJoystick() {
		Inputs.AxisX = CH.AxisX() * -1 //физическая ось джойстика вправо, наша логическая влево, как угол поворота
		Inputs.AxisY = CH.AxisY()
		Inputs.But1 = CH.GetJoybutton(0)
	} else {
		if CH.GetKey(sdl.SCANCODE_A) {
			Inputs.AxisX += 1
		}
		if CH.GetKey(sdl.SCANCODE_D) {
			Inputs.AxisX -= 1
		}
		if CH.GetKey(sdl.SCANCODE_W) {
			Inputs.AxisY += 1
		}
		if CH.GetKey(sdl.SCANCODE_S) {
			Inputs.AxisY -= 1
		}
		Inputs.But1 = CH.GetKey(sdl.SCANCODE_X)
	}

	c.Inputs = Inputs
}

//Копия в текстурах
func angClamp(ang float32) float32 {
	switch {
	case ang < 0:
		{
			return angClamp(ang + 360)
		}
	case ang >= 360:
		{
			return angClamp(ang - 360)
		}
	default:
		return ang
	}
}

//Угол между заданными направлениями a-b. Нормированный к (-180,+180)
func angSub(a, b float32) float32 {

	return angClamp(a-b+180) - 180
}

func (c *Cogitator) Draw(r *sdl.Renderer) (res scene.RenderReqList) {

	White := sdl.Color{255, 255, 255, 100}
	Yellow := sdl.Color{255, 255, 0, 255}
	Green := sdl.Color{0, 255, 0, 255}

	center := sdl.Point{winW / 2, winH / 2}
	CompassR := winH * 35 / 100

	cirl1 := psCircle(center, CompassR, 32)
	req1 := scene.NewRenderDrawLinesReq(cirl1, White, scene.Z_HUD)

	angColor := Yellow
	d := angSub(c.wantedAngle, c.Stats.angle)
	const greenDAng = 3
	if d < greenDAng && d > -greenDAng {
		angColor = Green
	}

	wantedAng := psTrigon(center, c.wantedAngle, CompassR+5, CompassR+25, 20)
	req2 := scene.NewRenderDrawLinesReq(wantedAng, angColor, scene.Z_HUD)

	currentAng := psTrigon(center, c.Stats.angle, CompassR-5, CompassR-25, 20)
	req3 := scene.NewRenderDrawLinesReq(currentAng, angColor, scene.Z_HUD)

	thrColor := Yellow
	const greenthrD = 0.03
	if abs(c.wantedMainThrust-c.Stats.mainThrust) < greenthrD {
		thrColor = Green
	}

	const angwidthOfThrust = 30
	w_ang := 90 + angwidthOfThrust*(1-2*c.wantedMainThrust)
	wantedThr := psTrigon(center, w_ang, CompassR+105, CompassR+125, 20)
	req4 := scene.NewRenderDrawLinesReq(wantedThr, thrColor, scene.Z_HUD)

	c_ang := 90 + angwidthOfThrust*(1-2*c.Stats.mainThrust)
	currentThr := psTrigon(center, c_ang, CompassR+95, CompassR+75, 20)
	req5 := scene.NewRenderDrawLinesReq(currentThr, thrColor, scene.Z_HUD)

	res = append(res, req1, req2, req3, req4, req5)

	return res
}

func psCircle(center sdl.Point, radius int32, num int) (res []sdl.Point) {
	for i := 0; i <= num; i++ {
		ang := float32(i) * 360 / float32(num)
		dV := V2.InDir(ang).Mul(float32(radius))
		res = append(res, sdl.Point{center.X + int32(dV.X), center.Y + int32(dV.Y)})
	}
	return res
}

func psTrigon(center sdl.Point, angle float32, rTop, rBot, botW int32) (res []sdl.Point) {
	ort := V2.InDir(angle)
	botM := ort.Mul(float32(rBot))
	dBot := ort.Rotate90().Mul(float32(botW) / 2)
	top := ort.Mul(float32(rTop))
	V1 := sdl.Point{center.X + top.ScX(), center.Y + top.ScY()}
	V2 := sdl.Point{center.X + botM.Add(dBot).ScX(), center.Y + botM.Add(dBot).ScY()}
	V3 := sdl.Point{center.X + botM.Sub(dBot).ScX(), center.Y + botM.Sub(dBot).ScY()}
	res = append(res, V1, V2, V3, V1)
	return res
}

func sign(x float32) float32 {
	switch {
	case x > 0:
		{
			return 1
		}
	case x < 0:
		{
			return -1
		}
	default:
		{
			return 0
		}
	}
}

func abs(x float32) float32 {
	if x < 0 {
		return -x
	} else {
		return x
	}
}

func sqrt(x float32) float32 {
	return float32(math.Sqrt(float64(x)))
}
