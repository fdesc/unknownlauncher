package auth

import (
	"bytes"
	"image"
	"golang.org/x/image/draw"
	"image/png"

	"fdesc/unknownlauncher/util/downloadutil"
)

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

var DefaultSkinIcon = false

func CropSkinImage(skinUrl string) image.Image {
	imageData,err := downloadutil.GetData(skinUrl)
	if err != nil {
		DefaultSkinIcon = true
		return nil
	}
	decodedImage,err := png.Decode(bytes.NewReader(imageData))
	if err != nil {
		DefaultSkinIcon = true
		return nil
	}
	rect := image.Rect(8,8,16,16)
	croppedImg := decodedImage.(SubImager).SubImage(rect)
	newResized := image.NewRGBA(image.Rect(0,0,croppedImg.Bounds().Max.X+48,croppedImg.Bounds().Max.Y+48))
	draw.NearestNeighbor.Scale(newResized,newResized.Rect,croppedImg,croppedImg.Bounds(),draw.Over,nil)
	DefaultSkinIcon = false
	return newResized
}
