package mnt

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Shnifer/flierproto1/v2"
	"log"
	"strconv"
	"strings"
)

//Базовые параметры корабля для 100% состояния систем,
//загружаются перед стартом и не изменяются
type BaseShipParameters struct {
	MaxThrust   float32
	MaxMomentum float32
	ScanRange   float32
	ScanSpeed   float32
	FuelStock   float32
	LifeStock   float32
	TotalMass   float32
	FuelMass    float32
}

func (sbp *BaseShipParameters) Encode() []byte {
	exbuf, err := json.Marshal(sbp)
	if err != nil {
		log.Panicln(err)
	}
	return exbuf
}

func (sbp *BaseShipParameters) Decode(str []byte) {
	log.Println(string(str))
	err := json.Unmarshal(str, sbp)
	if err != nil {
		log.Panicln(err)
	}
}

const SystemsCount = 8
const (
	SMarch  = "S_March"
	SHyper  = "S_Hyper"
	SManeur = "S_Maneur"
	SRadar  = "S_Radar"
	SSonar  = "S_Sonar"
	SFuel   = "S_Fuel"
	SLife   = "S_Life"
	SShield = "S_Shield"
)
const (
	PLifeStock = "P_Lifestock"
	PFuelStock = "P_Fuelstock"
)
const ParamCount = 2

var SNames = [SystemsCount]string{SMarch, SHyper, SManeur, SRadar, SSonar, SFuel, SLife, SShield}
var StrSName = [SystemsCount]string{"Маршевый двигатель", "Пузырь Алькубьерре", "Маневровые двигатели", "Курсовой радар",
	"Сонар", "Топливный бак", "Жизнеобеспечение", "Щиты"}
var PNames = [ParamCount]string{PLifeStock, PFuelStock}
var StrPName = [ParamCount]string{"Запас кислорода", "Запас топлива"}

func IsSystemName(name string) bool {
	for _, v := range SNames {
		if v == name {
			return true
		}
	}
	return false
}

//Состояние систем корабля от 0(уничтожена) до 1(штатно)
//Синхронизируется по сети
type ShipSystemsState map[string]float32

func NewShipSystemsState() ShipSystemsState {
	res := make(ShipSystemsState)
	for i, name := range SNames {
		//TODO: вернуть res[name] = 1, текущая формула - для проверки
		res[name] = 0.1*float32(i) + 0.1
	}
	return res
}

func (sss *ShipSystemsState) Encode() string {
	exbuf, err := json.Marshal(sss)
	if err != nil {
		log.Panicln(err)
	}
	return string(exbuf)
}

func (sss *ShipSystemsState) Decode(str string) {
	err := json.Unmarshal([]byte(str), sss)
	if err != nil {
		log.Panicln(err)
	}
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

func DownloadShipBaseParameters(sbp *BaseShipParameters) {
	res, err := Client.CommandResult(CMD_GETBSP, RES_BSP)
	if err != nil {
		log.Panicln(err)
	}
	_, msg := SplitMsg(res)
	sbp.Decode([]byte(msg))
}
