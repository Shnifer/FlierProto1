//Модуль обрабоки двумерных вкторов
package V2

import (
	"math"
	"math/rand"
)

//двумерный вектор. Тип float32 выбрать для совместимости с СДЛ
type V2 struct {
	X, Y float32
}

const Deg2Rad = 2 * math.Pi / 360

//Генераторы

//Возвращает случайный вектор единичной длины
func RandomOrt() V2 {
	a := rand.Float64() * 2 * math.Pi
	return V2{float32(math.Sin(a)), float32(math.Cos(a))}
}

func RandomInCircle(R float32) V2 {
	if R == 0 {
		return V2{}
	}
	ort := RandomOrt()
	dist := float32(math.Sqrt(float64(rand.Float32() * (R * R))))
	return Mul(ort, dist)
}

//возвращает единичный вектор , принимает угол в градусах
//0 - вверх (0,1) , положительно - против часовой
//Предназначен для пересчетов реального мира, для экранных координат применят осторожно
func InDir(angle float32) V2 {
	a := float64(angle * Deg2Rad)
	return V2{-float32(math.Sin(a)), float32(math.Cos(a))}
}

//Операции

//процерурные варианты

func AddMul(a, b V2, t float32) V2 {
	return Add(a, Mul(b, t))
}

func Rotate(V V2, angle float32) V2 {
	a := float64(angle * Deg2Rad)
	sin := float32(math.Sin(a))
	cos := float32(math.Cos(a))
	return V2{
		X: V.X*cos - V.Y*sin,
		Y: V.Y*cos + V.X*sin,
	}
}

//переводит вектор в систему с началом кординат в pos и повернутую на angle
func ApplyOnTransform(V, pos V2, angle float32) V2 {
	return Add(pos, Rotate(V, angle))
}

//складывает два вектора
func Add(a, b V2) V2 {
	return V2{a.X + b.X, a.Y + b.Y}
}

//вычитает a-b
func Sub(a, b V2) V2 {
	return V2{a.X - b.X, a.Y - b.Y}
}

//умножает вектор a на число t
func Mul(a V2, t float32) V2 {
	return V2{a.X * t, a.Y * t}
}

//Длина вектора а
func Len(a V2) float32 {
	return float32(math.Sqrt(float64(a.X*a.X + a.Y*a.Y)))
}

//квадрат длины r
func LenSqr(a V2) float32 {
	return a.X*a.X + a.Y*a.Y
}

//Возвращает нормализованный вектор
func Normed(a V2) V2 {
	if a.X == 0 && a.Y == 0 {
		return a
	}
	K := 1 / Len(a)
	return Mul(a, K)
}

//В виде методов

func (a V2) Add(b V2) V2 {
	return Add(a, b)
}

func (a V2) Sub(b V2) V2 {
	return Sub(a, b)
}

func (a V2) Mul(t float32) V2 {
	return Mul(a, t)
}

func (a V2) Len() float32 {
	return Len(a)
}

func (a V2) LenSqr() float32 {
	return LenSqr(a)
}

func (a V2) Normed() V2 {
	return Normed(a)
}

func (a *V2) DoNorm() {
	*a = Normed(*a)
}

func (a V2) Rotate(angle float32) V2 {
	return Rotate(a, angle)
}

func (v V2) ApplyOnTransform(pos V2, angle float32) V2 {
	return ApplyOnTransform(v, pos, angle)
}

func (a V2) AddMul(b V2, t float32) V2 {
	return AddMul(a, b, t)
}

func (a *V2) DoAddMul(b V2, t float32) {
	*a = AddMul(*a, b, t)
}
