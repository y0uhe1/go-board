package main

import (
	"bytes"
	"errors"
	"image"
	"image/png"
	"io/ioutil"
	"unicode/utf8"
	"unsafe"

	"github.com/cwchiu/go-winapi"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

func makeBoard() (*Board, error) {
	f, err := ioutil.ReadFile("../font/BIZ-UDGothicB.ttc")

	if err != nil {
		return nil, err
	}

	ft, err := truetype.Parse(f)

	if err != nil {
		return nil, err
	}

	opt := truetype.Options{
		Size:              50,
		DPI:               0,
		Hinting:           0,
		GlyphCacheEntries: 0,
		SubPixelsX:        0,
		SubPixelsY:        0,
	}

	imageWidth := utf8.RuneCountInString(text) * 100
	imageHeight := 100
	textTopMargin := 90

	img := image.NewRGBA(image.Rect(0, 0, imageWidth, imageHeight))

	face := truetype.NewFace(ft, &opt)

	dr := &font.Drawer{
		Dst:  img,
		Src:  image.Black,
		Face: face,
		Dot:  fixed.Point26_6{},
	}

	dr.Dot.X = (fixed.I(imageWidth) - dr.MeasureString(text)) / 2
	dr.Dot.Y = fixed.I(textTopMargin)

	dr.DrawString(text)

	buf := &bytes.Buffer{}
	err = png.Encode(buf, img)

	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	winapi.SystemParametersInfo(winapi.SPI_GETWORKAREA, 0, unsafe.Pointer(&rc), 0)

	hBitmap, err = hBitmapFromImage(img)

	if err != nil {
		return nil, err
	}

	hRgn = toRgn(img)

	return &Board{
		x:  int(rc.Right),
		y:  0,
		w:  bounds.Dx(),
		h:  bounds.Dy(),
		dx: 10,
		dy: 0,
	}, nil
}

func hBitmapFromImage(img image.Image) (winapi.HBITMAP, error) {
	var bi winapi.BITMAPV5HEADER
	bi.BiSize = uint32(unsafe.Sizeof(bi))
	bi.BiWidth = int32(img.Bounds().Dx())
	bi.BiHeight = -int32(img.Bounds().Dy())
	bi.BiPlanes = 1
	bi.BiBitCount = 32
	bi.BiCompression = winapi.BI_BITFIELDS
	bi.BV4RedMask = 0x00FF0000
	bi.BV4GreenMask = 0x0000FF00
	bi.BV4BlueMask = 0x000000FF
	bi.BV4AlphaMask = 0xFF000000

	hdc := winapi.GetDC(0)
	defer winapi.ReleaseDC(0, hdc)

	var bits unsafe.Pointer
	hBitmap := winapi.CreateDIBSection(
		hdc, &bi.BITMAPINFOHEADER, winapi.DIB_RGB_COLORS, &bits, 0, 0)
	switch hBitmap {
	case 0, winapi.ERROR_INVALID_PARAMETER:
		return 0, errors.New("CreateDIBSection failed")
	}

	ba := (*[1 << 30]byte)(unsafe.Pointer(bits))
	i := 0
	for y := img.Bounds().Min.Y; y != img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x != img.Bounds().Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			ba[i+3] = byte(a >> 8)
			ba[i+2] = byte(r >> 8)
			ba[i+1] = byte(g >> 8)
			ba[i+0] = byte(b >> 8)
			i += 4
		}
	}
	return hBitmap, nil
}

func toRgn(img image.Image) winapi.HRGN {
	hRgn := winapi.CreateRectRgn(0, 0, 0, 0)
	for y := img.Bounds().Min.Y; y != img.Bounds().Max.Y; y++ {
		opaque := false
		v := 0
		for x := img.Bounds().Min.X; x != img.Bounds().Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			// combine transparent colors
			if a > 0 {
				if !opaque {
					opaque = true
					v = x
				}
			} else {
				if opaque {
					addMask(hRgn, v, y, x, y+1)
					opaque = false
				}
			}
		}
		if opaque {
			addMask(hRgn, v, y, img.Bounds().Max.X, y+1)
		}
	}
	return hRgn
}

func addMask(hRgn winapi.HRGN, left, top, right, bottom int) {
	mask := winapi.CreateRectRgn(int32(left), int32(top), int32(right), int32(bottom))
	winapi.CombineRgn(hRgn, mask, hRgn, winapi.RGN_OR)
	winapi.DeleteObject(winapi.HGDIOBJ(mask))
}
