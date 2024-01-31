package resigif

import "runtime"

type Option func(*option)

type option struct {
	aspectRatio   aspectRatioOption
	resizeFunc    ImageResizeFunc
	parallelLimit int
}

var defaultOption = &option{
	aspectRatio:   Maintain,
	resizeFunc:    defaultImageResizeFunc,
	parallelLimit: runtime.NumCPU(),
}

type aspectRatioOption int

const (
	// Ignore ignores aspect ratio
	Ignore aspectRatioOption = iota
	// Maintain maintains aspect ratio
	Maintain
)

// WithAspectRatio sets aspect ratio option
//
//	default: Maintain
func WithAspectRatio(aspectRatio aspectRatioOption) Option {
	return func(o *option) {
		o.aspectRatio = aspectRatio
	}
}

// WithImageResizeFunc sets image resize function
//
//	default: using draw.CatmullRom
func WithImageResizeFunc(resizeFunc ImageResizeFunc) Option {
	return func(o *option) {
		o.resizeFunc = resizeFunc
	}
}

// WithParallel sets limit of parallel processing threads
//
//	ignores limit if limit <= 0
//	default: runtime.NumCPU()
func WithParallel(limit int) Option {
	return func(o *option) {
		if limit > 0 {
			o.parallelLimit = limit
		}
	}
}
