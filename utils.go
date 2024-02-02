package resigif

import (
	"image"
)

func deepCopyImage(src *image.NRGBA) *image.NRGBA {
	dst := &image.NRGBA{
		Pix:    make([]uint8, len(src.Pix)),
		Stride: src.Stride,
		Rect:   src.Rect,
	}
	copy(dst.Pix, src.Pix)

	return dst
}
