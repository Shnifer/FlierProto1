package mnt

import (
	"github.com/Shnifer/flierproto1/v2"
	"strconv"
	"fmt"
	"strings"
	"errors"
)

//Базовые параметры корабля для 100% состояния систем,
//загружаются перед стартом и не изменяются
type ShipBaseParameters struct{
	MaxThrust float32
	MaxMomentum float32
	ScanRange float32
	ScanTime float32
}

const (
	SEngine int = iota
	SScaner
	systemsCount
)
//Состояние систем корабля от 0(уничтожена) до 1(штатно)
//Синхронизируется по сети
type ShipSystemsState [systemsCount]float32
func NewShipSystemsState () ShipSystemsState{
	res:=ShipSystemsState{}
	for i:=range res{
		res[i]=1
	}
	return res
}

type ShipPosData struct {
	Pos        V2.V2
	Speed      V2.V2
	Angle      float32
	AngleSpeed float32
}

func DecodeShipPos(param string) (*ShipPosData, error) {
	parts := strings.SplitN(param, " ", 6)
	if len(parts) < 6 {
		return nil, errors.New("DecodeShipPos less than 6 params")
	}
	fparts := make([]float32, 6)
	for i := 0; i < 6; i++ {
		val, err := strconv.ParseFloat(parts[i], 32)
		if err != nil {
			return nil, err
		}
		fparts[i] = float32(val)
	}
	return &ShipPosData{
		Pos:        V2.V2{fparts[0], fparts[1]},
		Speed:      V2.V2{fparts[2], fparts[3]},
		Angle:      fparts[4],
		AngleSpeed: fparts[5],
	}, nil
}

func EncodeShipPos(data ShipPosData) string {
	return fmt.Sprintf("%f %f %f %f %f %f",
		data.Pos.X, data.Pos.Y, data.Speed.X, data.Speed.Y, data.Angle, data.AngleSpeed)
}

