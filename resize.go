package resigif

import (
	"context"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"math"

	"github.com/disintegration/imaging"
	"golang.org/x/sync/errgroup"
)

func Resize(src *gif.GIF, width, height int) (*gif.GIF, error) {
	srcWidth, srcHeight := src.Config.Width, src.Config.Height

	// 元の比率を保つよう調整 & 拡大・縮小比率を計算
	floatSrcWidth, floatSrcHeight, floatWidth, floatHeight := float64(srcWidth), float64(srcHeight), float64(width), float64(height)
	ratio := floatWidth / floatSrcWidth
	if floatSrcWidth/floatSrcHeight > floatWidth/floatHeight {
		ratio = floatWidth / floatSrcWidth
		height = int(math.Round(floatSrcHeight * ratio))
	} else if floatSrcWidth/floatSrcHeight < floatWidth/floatHeight {
		ratio = floatHeight / floatSrcHeight
		width = int(math.Round(floatSrcWidth * ratio))
	}

	destImage := &gif.GIF{
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

	var (
		gifBound = image.Rect(0, 0, srcWidth, srcHeight)
		// フレームを重ねるためのキャンバス
		//	差分最適化されたGIFに対応するための処置
		// 	差分最適化されたGIFでは、1フレーム目以外、周りが透明ピクセルのフレームを
		// 	次々に重ねていくことでアニメーションを表現する
		// 	周りが透明ピクセルのフレームをそのまま縮小すると、周りの透明ピクセルと
		// 	混ざった色が透明色ではなくなってフレームの縁に黒っぽいノイズが入ってしまう
		// 	ため、キャンバスでフレームを重ねてから縮小する
		tempCanvas = image.NewNRGBA(gifBound)
		// 	DisposalBackgroundに対応するための、背景色の画像
		bgColorUniform = image.NewUniform(src.Config.ColorModel.(color.Palette)[src.BackgroundIndex])
		// DisposalPreviousに対応するため、直前のフレームを保持するためのキャンバス
		backupCanvas = image.NewNRGBA(gifBound)

		eg, _ = errgroup.WithContext(context.Background())
	)

	for i, srcFrame := range src.Image {
		// 元のフレームのサイズと位置
		//  差分最適化されたGIFでは、これが元GIFのサイズより小さいことがある
		srcBounds := srcFrame.Bounds()
		// 縮小後のフレームのサイズと位置を計算
		destBounds := image.Rect(
			int(math.Round(float64(srcBounds.Min.X)*ratio)),
			int(math.Round(float64(srcBounds.Min.Y)*ratio)),
			int(math.Round(float64(srcBounds.Max.X)*ratio)),
			int(math.Round(float64(srcBounds.Max.Y)*ratio)),
		)

		// DisposalがPreviousなら、今のキャンバスをDeep Copyしてバックアップ
		if src.Disposal[i] == gif.DisposalPrevious {
			backupCanvas = &image.NRGBA{
				Pix:    append([]uint8{}, tempCanvas.Pix...),
				Stride: tempCanvas.Stride,
				Rect:   tempCanvas.Rect,
			}
		}

		// キャンバスに読んだフレームを重ねる
		draw.Draw(tempCanvas, srcBounds, srcFrame, srcBounds.Min, draw.Over)

		// 拡縮用GoRoutineを起動
		eg.Go(resizeRoutine(frameData{
			index: i,
			tempCanvas: &image.NRGBA{
				Pix:    append([]uint8{}, tempCanvas.Pix...),
				Stride: tempCanvas.Stride,
				Rect:   tempCanvas.Rect,
			}, // tempCanvasはポインタを使い回しているので、Deep Copyする
			resizeWidth:  width,
			resizeHeight: height,
			srcBounds:    srcBounds,
			destBounds:   destBounds,
			srcPalette:   src.Image[i].Palette,
		}, destImage))

		switch src.Disposal[i] {
		case gif.DisposalBackground: // DisposalがBackgroundなら、このフレームの範囲を背景色で塗りつぶす
			// フレームのカラーパレットに透明色が含まれていたら、背景色を透明色とみなす
			r, g, b, a := srcFrame.Palette[srcFrame.Palette.Index(color.Transparent)].RGBA()
			if r == 0 && g == 0 && b == 0 && a == 0 {
				draw.Draw(tempCanvas, srcBounds, image.Transparent, image.Point{}, draw.Src)
			} else {
				draw.Draw(tempCanvas, srcBounds, bgColorUniform, image.Point{}, draw.Src)
			}
		case gif.DisposalPrevious: // DisposalがPreviousなら、直前のフレームを復元
			tempCanvas = backupCanvas
		}
	}

	err := eg.Wait()
	if err != nil {
		return nil, err
	}

	return destImage, nil
}

// GIFのリサイズ時、拡縮用GoRoutineに渡すフレームのデータ
type frameData struct {
	index        int
	tempCanvas   *image.NRGBA
	resizeWidth  int
	resizeHeight int
	srcBounds    image.Rectangle
	destBounds   image.Rectangle
	srcPalette   color.Palette
}

func resizeRoutine(data frameData, destImage *gif.GIF) func() error {
	return func() error {
		// 重ねたフレームを縮小
		fittedImage := imaging.Resize(data.tempCanvas, data.resizeWidth, data.resizeHeight, mks2013Filter)

		// destBoundsに合わせて、縮小されたイメージを切り抜き
		destFrame := image.NewPaletted(data.destBounds, data.srcPalette)
		draw.Draw(destFrame, data.destBounds, fittedImage.SubImage(data.destBounds), data.destBounds.Min, draw.Src)

		destImage.Image[data.index] = destFrame

		return nil
	}
}
