package resigif

import (
	"image"

	"golang.org/x/image/draw"
)

var defaultImageResizeFunc = FromDrawScaler(draw.CatmullRom)

// ImageResizeFunc resizes a single image
//
//	ImageResizeFunc can assume that src image is aligned to (0, 0) point
type ImageResizeFunc func(src *image.NRGBA, width, height int) (dst *image.NRGBA, err error)

// FromDrawScaler converts draw.Scaler to ImageResizeFunc
//
//	draw.Interpolator / *draw.Kernel can be used as draw.Scaler
func FromDrawScaler(scaler draw.Scaler) ImageResizeFunc {
	return func(src *image.NRGBA, width, height int) (dst *image.NRGBA, err error) {
		dst = image.NewNRGBA(image.Rect(0, 0, width, height))
		scaler.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Src, nil)
		return
	}
}
