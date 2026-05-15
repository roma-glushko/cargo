package sample

import (
	"time"

	"github.com/roma-glushko/cargo/internal/voyage"
)

func ts(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

var HongkongToNewYork = voyage.NewBuilder("0100S", "CNHKG").
	AddMovement("CNHGH", ts(2009, 3, 1), ts(2009, 3, 2)).
	AddMovement("JNTKO", ts(2009, 3, 3), ts(2009, 3, 5)).
	AddMovement("AUMEL", ts(2009, 3, 6), ts(2009, 3, 8)).
	AddMovement("USNYC", ts(2009, 3, 9), ts(2009, 3, 12)).
	Build()

var NewYorkToDallas = voyage.NewBuilder("0200T", "USNYC").
	AddMovement("USCHI", ts(2009, 3, 13), ts(2009, 3, 14)).
	AddMovement("USDAL", ts(2009, 3, 15), ts(2009, 3, 16)).
	Build()

var DallasToHelsinki = voyage.NewBuilder("0300A", "USDAL").
	AddMovement("DEHAM", ts(2009, 3, 17), ts(2009, 3, 19)).
	AddMovement("SESTO", ts(2009, 3, 20), ts(2009, 3, 21)).
	AddMovement("FIHEL", ts(2009, 3, 22), ts(2009, 3, 23)).
	Build()

var DallasToHelsinkiAlt = voyage.NewBuilder("0301S", "USDAL").
	AddMovement("FIHEL", ts(2009, 3, 17), ts(2009, 3, 20)).
	Build()

var HelsinkiToHongkong = voyage.NewBuilder("0400S", "FIHEL").
	AddMovement("NLRTM", ts(2009, 3, 23), ts(2009, 3, 25)).
	AddMovement("CNSHA", ts(2009, 3, 26), ts(2009, 3, 28)).
	AddMovement("CNHKG", ts(2009, 3, 29), ts(2009, 3, 31)).
	Build()

var TestVoyageV100 = voyage.NewBuilder("V100", "CNHKG").
	AddMovement("JNTKO", ts(2009, 3, 3), ts(2009, 3, 5)).
	AddMovement("USNYC", ts(2009, 3, 6), ts(2009, 3, 9)).
	Build()

var TestVoyageV200 = voyage.NewBuilder("V200", "JNTKO").
	AddMovement("USNYC", ts(2009, 3, 6), ts(2009, 3, 8)).
	AddMovement("USCHI", ts(2009, 3, 10), ts(2009, 3, 14)).
	AddMovement("SESTO", ts(2009, 3, 14), ts(2009, 3, 16)).
	Build()

var TestVoyageV300 = voyage.NewBuilder("V300", "JNTKO").
	AddMovement("NLRTM", ts(2009, 3, 8), ts(2009, 3, 11)).
	AddMovement("DEHAM", ts(2009, 3, 11), ts(2009, 3, 12)).
	AddMovement("AUMEL", ts(2009, 3, 14), ts(2009, 3, 18)).
	AddMovement("JNTKO", ts(2009, 3, 19), ts(2009, 3, 21)).
	Build()

var TestVoyageV400 = voyage.NewBuilder("V400", "DEHAM").
	AddMovement("SESTO", ts(2009, 3, 14), ts(2009, 3, 15)).
	AddMovement("FIHEL", ts(2009, 3, 15), ts(2009, 3, 16)).
	AddMovement("DEHAM", ts(2009, 3, 20), ts(2009, 3, 23)).
	Build()

var AllVoyages = []*voyage.Voyage{
	HongkongToNewYork,
	NewYorkToDallas,
	DallasToHelsinki,
	DallasToHelsinkiAlt,
	HelsinkiToHongkong,
	TestVoyageV100,
	TestVoyageV200,
	TestVoyageV300,
	TestVoyageV400,
}
