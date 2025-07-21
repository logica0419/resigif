package resigif

// Option is option for GIF resizing1
type Option func(*processor)

type aspectRatioOption int

const (
	// Ignore ignores aspect ratio
	Ignore aspectRatioOption = iota
	// Maintain maintains aspect ratio
	Maintain
	// WidthFirst maintains aspect ratio, prioritizing width
	WidthFirst
	// HeightFirst maintains aspect ratio, prioritizing height
	HeightFirst
)

// WithAspectRatio sets aspect ratio option
//
//	default: Maintain
func WithAspectRatio(aspectRatio aspectRatioOption) Option {
	return func(o *processor) {
		o.aspectRatio = aspectRatio
	}
}

// WithImageResizeFunc sets image resize function
//
//	default: using draw.CatmullRom
func WithImageResizeFunc(resizeFunc ImageResizeFunc) Option {
	return func(o *processor) {
		if resizeFunc != nil {
			o.resizeFunc = resizeFunc
		}
	}
}

// WithParallel sets limit of parallel processing threads
//
//	ignores limit if limit <= 0
//	default: runtime.NumCPU()
func WithParallel(limit int) Option {
	return func(o *processor) {
		if limit > 0 {
			o.parallelLimit = limit
		}
	}
}
