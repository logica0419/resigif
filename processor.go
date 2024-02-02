package resigif

import (
	"context"
	"image"
	"image/color"
	"image/gif"
	"math"
	"runtime"

	"golang.org/x/image/draw"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type processor struct {
	sem *semaphore.Weighted

	aspectRatio   aspectRatioOption
	resizeFunc    ImageResizeFunc
	parallelLimit int
}

func newProcessor(opts ...Option) *processor {
	proc := &processor{
		sem:           nil,
		aspectRatio:   Maintain,
		resizeFunc:    FromDrawScaler(draw.CatmullRom),
		parallelLimit: runtime.NumCPU(),
	}

	for _, opt := range opts {
		opt(proc)
	}

	proc.sem = semaphore.NewWeighted(int64(proc.parallelLimit))

	return proc
}

func (p *processor) resize(ctx context.Context, src *gif.GIF, width, height int) (*gif.GIF, error) {
	srcWidth, srcHeight := src.Config.Width, src.Config.Height

	// Calculate resize ratio & update width and height
	widthRatio, heightRatio := p.calculateRatio(srcWidth, srcHeight, &width, &height)

	var (
		// Canvas to pile up frames
		// 	When resizing a frame with transparent pixels, the edge pixels may be
		//  mixed with the transparent color and produce black jagged noise
		// 	To avoid the noise, pile up frames on this canvas before resizing
		tempCanvas = image.NewNRGBA(image.Rect(0, 0, srcWidth, srcHeight))

		// Uniform image of background color
		// nolint:forcetypeassert
		bgColorUniform = image.NewUniform(src.Config.ColorModel.(color.Palette)[src.BackgroundIndex])

		// Destination GIF image
		dst = &gif.GIF{
			Image:     make([]*image.Paletted, len(src.Image)),
			Delay:     src.Delay,
			LoopCount: src.LoopCount,
			Disposal:  src.Disposal,
			Config: image.Config{
				ColorModel: src.Config.ColorModel,
				Width:      width,
				Height:     height,
			},
			BackgroundIndex: src.BackgroundIndex,
		}
	)

	var eg *errgroup.Group
	eg, ctx = errgroup.WithContext(ctx)

	for i, srcFrame := range src.Image {
		// Frame size and position
		//	This may be different from the size of the GIF image
		srcBounds := srcFrame.Bounds()
		// Calculate resized frame size and position
		destBounds := image.Rect(
			int(math.Round(float64(srcBounds.Min.X)*widthRatio)),
			int(math.Round(float64(srcBounds.Min.Y)*heightRatio)),
			int(math.Round(float64(srcBounds.Max.X)*widthRatio)),
			int(math.Round(float64(srcBounds.Max.Y)*heightRatio)),
		)

		// If the disposal method is "none", pile up the frame to tempCanvas before passing it to a goroutine
		if src.Disposal[i] == gif.DisposalNone {
			draw.Draw(tempCanvas, srcBounds, srcFrame, srcBounds.Min, draw.Over)
			eg.Go(
				p.resizeFrame(ctx, i, deepCopyImage(tempCanvas), destBounds, srcFrame.Palette, dst),
			)

			continue
		}

		// Pile up the frame to tempCanvas and resize it in a goroutine
		eg.Go(
			p.pileAndResizeFrame(ctx, i, srcFrame, deepCopyImage(tempCanvas), destBounds, srcFrame.Palette, dst),
		)

		// If the disposal method is "background", fill tempCanvas with the background color
		if src.Disposal[i] == gif.DisposalBackground {
			// If the transparent color is in the frame palette, use it as the background color
			r, g, b, a := srcFrame.Palette[srcFrame.Palette.Index(color.Transparent)].RGBA()
			if r == 0 && g == 0 && b == 0 && a == 0 {
				draw.Draw(tempCanvas, srcBounds, image.Transparent, image.Point{X: 0, Y: 0}, draw.Src)
			} else {
				draw.Draw(tempCanvas, srcBounds, bgColorUniform, image.Point{X: 0, Y: 0}, draw.Src)
			}
		}
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return dst, nil
}

func (p *processor) calculateRatio(srcWidth, srcHeight int, width, height *int) (float64, float64) {
	// Calculate resize ratio
	floatSrcWidth, floatSrcHeight, floatWidth, floatHeight := float64(srcWidth), float64(srcHeight), float64(*width), float64(*height)
	widthRatio := floatWidth / floatSrcWidth
	heightRatio := floatHeight / floatSrcHeight

	// If aspect ratio should be maintained, use the smaller ratio
	//  Update width and height accordingly
	if p.aspectRatio == Maintain {
		if widthRatio < heightRatio {
			heightRatio = widthRatio
			*height = int(math.Round(floatSrcHeight * heightRatio))
		} else {
			widthRatio = heightRatio
			*width = int(math.Round(floatSrcWidth * widthRatio))
		}
	}

	return widthRatio, heightRatio
}

func (p *processor) pileAndResizeFrame(
	ctx context.Context,
	index int,
	frame *image.Paletted,
	tempCanvas *image.NRGBA,
	destBounds image.Rectangle,
	srcPalette color.Palette,
	dst *gif.GIF,
) func() error {
	return func() error {
		draw.Draw(tempCanvas, frame.Bounds(), frame, frame.Bounds().Min, draw.Over)

		return p.resizeFrame(ctx, index, tempCanvas, destBounds, srcPalette, dst)()
	}
}

func (p *processor) resizeFrame(
	ctx context.Context,
	index int,
	frame *image.NRGBA,
	destBounds image.Rectangle,
	srcPalette color.Palette,
	dst *gif.GIF,
) func() error {
	return func() error {
		// Acquire semaphore
		if err := p.sem.Acquire(ctx, 1); err != nil {
			return err
		}

		defer p.sem.Release(1)

		// Resize image
		fittedImage, err := p.resizeFunc(frame, dst.Config.Width, dst.Config.Height)
		if err != nil {
			return err
		}

		// Crop image to fit destBounds and put it into dst
		dst.Image[index] = image.NewPaletted(destBounds, srcPalette)
		draw.Draw(dst.Image[index], destBounds, fittedImage.SubImage(destBounds), destBounds.Min, draw.Src)

		return nil
	}
}
