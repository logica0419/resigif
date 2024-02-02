// Package resigif is an Animated GIF resizing library w/o cgo nor any third-party Libraries
package resigif

import (
	"context"
	"image/gif"
)

// Resize resizes an animated GIF image
//
//	Returned error is either context error or error from ImageResizeFunc
func Resize(ctx context.Context, src *gif.GIF, width, height int, opts ...Option) (*gif.GIF, error) {
	p := newProcessor(opts...)

	return p.resize(ctx, src, width, height)
}
