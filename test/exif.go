package main

import (
	"fmt"
	"log"
	"os"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
)

func main() {
	fname := "/media/psf/Home/Desktop/DSC_2063.jpg"

	f, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}

	// Optionally register camera makenote data parsing - currently Nikon and
	// Canon are supported.
	exif.RegisterParsers(mknote.All...)

	x, err := exif.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	if camModel, err := x.Get(exif.Model); err == nil {
		fmt.Println(camModel.StringVal())
	}

	if focal, err := x.Get(exif.FocalLength); err == nil {
		numer, denom, err := focal.Rat2(0) // retrieve first (only) rat. value
		if err == nil {
			fmt.Println("focal:", numer, denom)
		}
	}

	// Two convenience functions exist for date/time taken and GPS coords:
	if tm, err := x.DateTime(); err == nil {
		fmt.Println("Taken: ", tm)
	}

	if lat, long, err := x.LatLong(); err == nil {
		fmt.Println("lat, long: ", lat, ", ", long)
	}
}
