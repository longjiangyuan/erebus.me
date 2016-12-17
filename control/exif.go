package control

import (
	"fmt"
	"strconv"

	"github.com/rwcarlsen/goexif/exif"
)

type ExifFormatter struct {
	ShutterSpeed string
	Aperture     string
	FocalLength  string
	ISO          string
	Model        string
	LensModel    string
	Taken        string
}

func (info *ExifFormatter) Format(x *exif.Exif) {
	info.LensModel = formatExifLensModel(x)
	info.Model = formatExifModel(x)
	info.Aperture = formatExifFNumber(x)
	info.FocalLength = formatExifFocalLength(x)
	info.ShutterSpeed = formatExifExposeTime(x)
	info.ISO = formatExifISOSpeedRating(x)

	if tm, err := x.DateTime(); err == nil {
		info.Taken = tm.Format("Jan _2, 2006 15:04")
	}
}

func formatExifLensModel(x *exif.Exif) string {
	tag, err := x.Get(exif.LensModel)
	if err != nil {
		return ""
	}
	v, _ := tag.StringVal()
	return v
}

func formatExifISOSpeedRating(x *exif.Exif) string {
	tag, err := x.Get(exif.ISOSpeedRatings)
	if err != nil {
		return ""
	}
	i, err := tag.Int(0)
	if err != nil {
		return ""
	}
	return strconv.Itoa(i)
}

func formatExifFocalLength(x *exif.Exif) string {
	tag, err := x.Get(exif.FocalLength)
	if err != nil {
		return ""
	}
	numer, denom, err := tag.Rat2(0)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%d mm", numer/denom)
}

func formatExifExposeTime(x *exif.Exif) string {
	tag, err := x.Get(exif.ExposureTime)
	if err != nil {
		return ""
	}
	numer, denom, err := tag.Rat2(0)
	if err != nil {
		return ""
	}
	//log.Println("formatExifExposeTime:", numer, denom)

	if numer > denom {
		return fmt.Sprintf("%.1f s", float64(numer)/float64(denom))
	} else {
		return fmt.Sprintf("%d/%d s", numer, denom)
	}
}

func formatExifFNumber(x *exif.Exif) string {
	tag, err := x.Get(exif.FNumber)
	if err != nil {
		return ""
	}
	numer, denom, err := tag.Rat2(0)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("f/%.1f", float64(numer)/float64(denom))
}

func formatExifModel(x *exif.Exif) string {
	tag, err := x.Get(exif.Model)
	if err != nil {
		return ""
	}
	v, _ := tag.StringVal()
	return v
}

func formatExifSoftware(x *exif.Exif) string {
	tag, err := x.Get(exif.Software)
	if err != nil {
		return ""
	}
	v, _ := tag.StringVal()
	return v
}
