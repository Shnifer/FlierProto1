package main

import (
	"encoding/json"
	MNT "github.com/Shnifer/flierproto1/mnt"
	V2 "github.com/Shnifer/flierproto1/v2"
	"io/ioutil"
	"log"
	"math/rand"
	"strconv"
)

//Диаметр галактики в абсолютных космических единицах

func randomStar(num int) *MNT.Star {
	GalaxyRadius := DEFVAL.GalaxyRadius
	minRadius := DEFVAL.Gen_minRadius
	maxRadius := DEFVAL.Gen_maxMass
	maxSpeed := DEFVAL.Gen_maxSpeed
	maxMass := DEFVAL.Gen_maxMass
	minMass := DEFVAL.Gen_minMass

	sizeK := rand.Float32()

	return &MNT.Star{
		Pos:        V2.RandomInCircle(GalaxyRadius),
		Dir:        V2.RandomInCircle(maxSpeed),
		ColRad:     sizeK*(maxRadius-minRadius) + minRadius,
		Mass:       sizeK*(maxMass-minMass) + minMass,
		ObservText: "IDN: " + strconv.Itoa(num),
	}
}

func GenerateRandomGalaxy() {
	NumStars := DEFVAL.NumStars

	MNT.GalaxyData = make([]*MNT.Star, NumStars)
	for i := 0; i < NumStars; i++ {
		Star := randomStar(i)
		MNT.GalaxyData[i] = Star
	}
}

//Возвращает коэффициент нормальной дистрибуций
//сигма в процентах devProcent
//68% попадут в (100-devProcent, 100+devProcent)
//95% попадут в (100-2*devProcent, 100+2*devProcent)
//Отклонения больше 3 сигма ограничиваются
func kDev(devProcent float32) float32 {
	r := float32(rand.NormFloat64())
	if r > 3 {
		r = 3
	}
	if r < (-3) {
		r = -3
	}
	r = 1 + r*devProcent/100
	if r < 0 {
		r = 0.00001
	}
	return r
}

func LoadGalaxyFromFile() {
	type fileData struct {
		ID, Parent  string
		Diameter    float32
		Distance    float32
		Mass        float32
		OrbitPeriod float32
		Color       MNT.Color
		Count       int
		//начальный угол, если объект 1
		StartAng float32
		//отклонения от базовых значений в процентах, если объектов много
		RadMassDev     float32
		PeriodOrbitDev float32
		TexName string
	}
	exData := make([]fileData, 2)
	exBuf, err := json.Marshal(exData)
	if err != nil {
		log.Panicln(err)
	}
	if err := ioutil.WriteFile(serverDataPath+"example_"+DEFVAL.LoadGalaxyFile, exBuf, 0); err != nil {
		log.Panicln("LoadGalaxyFromFile: can't write even example for you! ", err)
	}

	buf, err := ioutil.ReadFile(serverDataPath + DEFVAL.LoadGalaxyFile)
	if err != nil {
		log.Panicln("LoadGalaxyFromFile: can't read file", err)
	}
	var ReadData []fileData
	json.Unmarshal(buf, &ReadData)
	Len := 0
	for _, v := range ReadData {
		Len += v.Count
	}
	log.Println("loaded", Len, "objects of", len(ReadData), "kind")
	MNT.GalaxyData = make([]*MNT.Star, Len)

	parentsPos := make(map[string]V2.V2)

	Load_K_orbSpeed := DEFVAL.Load_K_OrbSpeed
	Load_K_radius := DEFVAL.Load_K_Radius
	n := 0
	for _, v := range ReadData {

		var pos V2.V2
		var pp V2.V2
		if v.Parent != "" {
			p, ok := parentsPos[v.Parent]
			if !ok {
				log.Panicln("object", v.ID, "can't fing parent", v.Parent, "before him")
			}
			pp = p
		}

		texName := v.TexName
		if texName == "" {
			texName = DEFVAL.Load_DefTexName
		}

		for i := 0; i < v.Count; i++ {
			id := v.ID
			orbDist := float32(0)
			orbSpeed := float32(0)
			if v.Parent != "" {
				orbDist = v.Distance
				orbSpeed = 360 / v.OrbitPeriod * Load_K_orbSpeed
			}
			radius := v.Diameter / 2 * Load_K_radius
			mass := v.Mass
			angle := v.StartAng

			if v.Count == 1 {
				//Если объект одиночный то в фале всё написано
				pos = pp.AddMul(V2.InDir(v.StartAng), v.Distance)
				if v.ID != "" {
					parentsPos[v.ID] = pos
				}
			} else {
				//Генерим случайного по отклонениям
				id = id + "-" + strconv.Itoa(i)

				//Если отклоняется дальше то крутиться медленнее
				kPeriodOrbit := kDev(v.PeriodOrbitDev)
				orbDist *= kPeriodOrbit
				orbSpeed /= kPeriodOrbit
				kRadMass := kDev(v.RadMassDev)
				radius *= kRadMass
				mass *= kRadMass

				angle = rand.Float32() * 360
				pos = pp.AddMul(V2.InDir(angle), orbDist)
			}

			Star := MNT.Star{
				ID:       id,
				Pos:      pos,
				Parent:   v.Parent,
				OrbDist:  orbDist,
				OrbSpeed: orbSpeed,
				Angle:    angle,
				ColRad:   radius,
				Mass:     mass,
				Color:    v.Color,
				TexName:  texName,
			}
			log.Println("created star", Star)
			MNT.GalaxyData[n] = &Star
			n++
		}
	}
}
