package main

import (
	V2 "github.com/Shnifer/flierproto1/v2"
	MNT "github.com/Shnifer/flierproto1/mnt"
	"math/rand"
	"strconv"
)

//Диаметр галактики в абсолютных космических единицах
const GalaxyRadius float32 = 20000
const NumStars = 1500

func randomStar(num int) *MNT.Star{
	const minRadius = 20
	const maxRadius = 60
	const maxSpeed = 0
	const maxMass = 600
	const minMass = 100

	sizeK:=rand.Float32()

	return &MNT.Star{
		Pos: V2.RandomInCircle(GalaxyRadius),
		Dir: V2.RandomInCircle(maxSpeed),
		ColRad: sizeK*(maxRadius-minRadius)+minRadius,
		Mass: sizeK*(maxMass-minMass)+minMass,
		ObservText: "IDN: "+strconv.Itoa(num),
	}
}

func GenerateRandomGalaxy() {
	MNT.GalaxyData = make([]*MNT.Star, NumStars)
	for i:=0;i<NumStars;i++{
		Star:=randomStar(i)
		MNT.GalaxyData[i] = Star
	}
}
