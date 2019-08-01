package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
)

func maybeIsPNG(uri string, d []byte) bool {
	ext := strings.ToLower(filepath.Ext(uri))
	if ext == ".png" {
		return true
	}
	// TODO: sniff png header in d
	return false
}

func maybeIsJPEG(uri string, d []byte) bool {
	ext := strings.ToLower(filepath.Ext(uri))
	if ext == ".jpeg" || ext == ".jpg" {
		return true
	}
	// TODO: sniff png header in d
	return false
}

func downloadImage(uri string) (image.Image, error) {
	resp, err := http.Get(uri)
	if err != nil {
		return nil, fmt.Errorf("http.Get('%s') failed with %s", uri, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("http.Get() failed with status '%s'", resp.Status)
	}
	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	triedPNG := false
	triedJPEG := false
	if maybeIsPNG(uri, d) {
		img, err := png.Decode(bytes.NewBuffer(d))
		if err == nil {
			return img, nil
		}
		triedPNG = true
	}
	if maybeIsJPEG(uri, d) {
		img, err := jpeg.Decode(bytes.NewBuffer(d))
		if err == nil {
			return img, nil
		}
		triedJPEG = true
	}
	if !triedPNG {
		img, err := png.Decode(bytes.NewBuffer(d))
		if err == nil {
			return img, nil
		}
	}
	if !triedJPEG {
		img, err := jpeg.Decode(bytes.NewBuffer(d))
		if err == nil {
			return img, nil
		}
	}
	return nil, fmt.Errorf("'%s' is not a valid PNG or JPEG image", uri)
}

func saveImageAsPNG(dst string, img image.Image) error {
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	err = png.Encode(f, img)
	// an error during Close is very unlikely but not impossible
	err2 := f.Close()
	if err == nil && err2 == nil {
		return nil
	}

	// in case of failure, don't leave invalid file on disk
	os.Remove(dst)
	if err != nil {
		return err
	}
	return err2
}

// :show start
func resize(src image.Image, dstSize image.Point) *image.RGBA {
	srcRect := src.Bounds()
	dstRect := image.Rectangle{
		Min: image.Point{0, 0},
		Max: dstSize,
	}
	dst := image.NewRGBA(dstRect)
	draw.CatmullRom.Scale(dst, dstRect, src, srcRect, draw.Over, nil)
	return dst
}

// :show end

func getProportionalY(p image.Point, x int) int {
	res := (int64(p.Y) * int64(x)) / int64(p.X)
	return int(res)
}

func main() {
	img, err := downloadImage("https://www.programming-books.io/covers/Go.png")
	if err != nil {
		log.Fatalf("downloadImage() failed with '%s'\n", err)
	}
	size := img.Bounds().Size()
	x := 140
	y := getProportionalY(size, x)
	fmt.Printf("resizing %d x %d => %d x %d\n", size.X, size.Y, x, y)
	resizedImg := resize(img, image.Point{x, y})
	err = saveImageAsPNG("go-resized.png", resizedImg)
	if err != nil {
		log.Fatalf("saveImageAsPNG() failed with '%s'\n", err)
	}
}
