package main

import (
	"github.com/Shnifer/FlierProto1/V2"
	"github.com/Shnifer/FlierProto1/MNT"
	"math/rand"
)

//Диаметр галактики в абсолютных космических единицах
const GalaxyRadius float32 = 5000
const NumStars = 100

func randomStar() *MNT.Star{
	const minRadius = 20
	const maxRadius = 60
	const maxSpeed = 10
	const maxMass = 1000
	const minMass = 100

	sizeK:=rand.Float32()

	return &MNT.Star{
		Pos: V2.RandomInCircle(GalaxyRadius),
		Dir: V2.RandomInCircle(maxSpeed),
		ColRad: sizeK*(maxRadius-minRadius)+minRadius,
		Mass: sizeK*(maxMass-minMass)+minMass,
		ObservText: "science data text",
	}
}

func GenerateRandomGalaxy() {
	MNT.GalaxyData = make([]*MNT.Star, NumStars)
	for i:=0;i<NumStars;i++{
		Star:=randomStar()
		MNT.GalaxyData[i] = Star
	}
}
