package resigif_test

import (
	"context"
	"flag"
	"image"
	"image/gif"
	"os"
	"path/filepath"
	"testing"

	"github.com/logica0419/resigif"
	"github.com/logica0419/resigif/testdata"
	"github.com/youta-t/its"
	"golang.org/x/image/draw"
)

var overWrite = flag.Bool("overwrite", false, "overwrite test file")

func mustOpenGif(t *testing.T, name string) *gif.GIF {
	t.Helper()

	file, err := testdata.FS.Open(name)
	its.Nil[error]().Match(err).OrFatal(t)

	defer func() {
		_ = file.Close()
	}()

	image, err := gif.DecodeAll(file)
	its.Nil[error]().Match(err).OrFatal(t)

	return image
}

func mustEncodeGif(t *testing.T, name string, image *gif.GIF) {
	t.Helper()

	file, err := os.OpenFile(filepath.Clean("testdata/"+name), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	its.Nil[error]().Match(err).OrFatal(t)

	defer func() {
		_ = file.Close()
	}()

	err = gif.EncodeAll(file, image)
	its.Nil[error]().Match(err).OrFatal(t)
}

func TestResize(t *testing.T) {
	t.Parallel()

	flag.Parse()

	type args struct {
		ctx    context.Context
		src    string
		width  int
		height int
		opts   []resigif.Option
	}

	tests := []struct {
		name       string
		args       args
		want       string
		errMatcher its.Matcher[error]
	}{
		{
			name: "success (mushroom: square / small)",
			args: args{
				ctx:    context.Background(),
				src:    "mushroom.gif",
				width:  256,
				height: 256,
				opts: []resigif.Option{
					resigif.WithAspectRatio(resigif.Ignore),
				},
			},
			want:       "mushroom_resized.gif",
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (bicycle: oblong)",
			args: args{
				ctx:    context.Background(),
				src:    "bicycle.gif",
				width:  256,
				height: 0,
				opts: []resigif.Option{
					resigif.WithAspectRatio(resigif.WidthFirst),
				},
			},
			want:       "bicycle_width_first_resized.gif",
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (bicycle: oblong )",
			args: args{
				ctx:    context.Background(),
				src:    "bicycle.gif",
				width:  0,
				height: 256,
				opts: []resigif.Option{
					resigif.WithAspectRatio(resigif.HeightFirst),
				},
			},
			want:       "bicycle_height_first_resized.gif",
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (tooth: square、DisposalBackground)",
			args: args{
				ctx:    context.Background(),
				src:    "tooth.gif",
				width:  256,
				height: 256,
				opts: []resigif.Option{
					resigif.WithParallel(1),
				},
			},
			want:       "tooth_resized.gif",
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (new_year: landscape)",
			args: args{
				ctx:    context.Background(),
				src:    "new_year.gif",
				width:  256,
				height: 256,
				opts: []resigif.Option{
					resigif.WithImageResizeFunc(resigif.FromDrawScaler(draw.BiLinear)),
				},
			},
			want:       "new_year_resized.gif",
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (miku: portrait / optimized)",
			args: args{
				ctx:    context.Background(),
				src:    "miku.gif",
				width:  256,
				height: 256,
				opts:   nil,
			},
			want:       "miku_resized.gif",
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (frog: portrait / DisposalBackground + wrong background color)",
			args: args{
				ctx:    context.Background(),
				src:    "frog.gif",
				width:  256,
				height: 256,
				opts:   nil,
			},
			want:       "frog_resized.gif",
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (surprised: square / empty global color table)",
			args: args{
				ctx:    context.Background(),
				src:    "surprised.gif",
				width:  256,
				height: 256,
				opts:   nil,
			},
			want:       "surprised_resized.gif",
			errMatcher: its.Nil[error](),
		},
		{
			name: "fail (error from ImageResizeFunc)",
			args: args{
				ctx:    context.Background(),
				src:    "surprised.gif",
				width:  256,
				height: 256,
				opts: []resigif.Option{resigif.WithImageResizeFunc(func(_ *image.NRGBA, _, _ int) (*image.NRGBA, error) {
					return nil, os.ErrInvalid
				})},
			},
			want:       "surprised_resized.gif",
			errMatcher: its.Error(os.ErrInvalid),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := resigif.Resize(tt.args.ctx, mustOpenGif(t, tt.args.src), tt.args.width, tt.args.height, tt.args.opts...)
			if err == nil && overWrite != nil && *overWrite {
				mustEncodeGif(t, tt.want, got)
			}

			tt.errMatcher.Match(err).OrError(t)

			if err != nil {
				return
			}

			its.DeepEqual(mustOpenGif(t, tt.want)).Match(got).OrError(t)
		})
	}
}
