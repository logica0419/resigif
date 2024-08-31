package resigif_test

import (
	"context"
	"flag"
	"image/gif"
	"os"
	"path/filepath"
	"testing"

	"github.com/logica0419/resigif"
	"github.com/logica0419/resigif/testdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/image/draw"
)

var overWrite = flag.Bool("overwrite", false, "overwrite test file")

func mustOpenGif(t *testing.T, name string) *gif.GIF {
	t.Helper()

	file, err := testdata.FS.Open(name)
	require.NoError(t, err)

	defer func() {
		_ = file.Close()
	}()

	image, err := gif.DecodeAll(file)
	require.NoError(t, err)

	return image
}

func mustEncodeGif(t *testing.T, name string, image *gif.GIF) {
	t.Helper()

	file, err := os.OpenFile(filepath.Clean("testdata/"+name), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	require.NoError(t, err)

	defer func() {
		_ = file.Close()
	}()

	err = gif.EncodeAll(file, image)
	require.NoError(t, err)
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
		name      string
		args      args
		want      string
		assertion require.ErrorAssertionFunc
	}{
		{
			name: "success (mushroom 正方形、小サイズ)",
			args: args{
				ctx:    context.Background(),
				src:    "mushroom.gif",
				width:  256,
				height: 256,
				opts: []resigif.Option{
					resigif.WithAspectRatio(resigif.Ignore),
				},
			},
			want:      "mushroom_resized.gif",
			assertion: require.NoError,
		},
		{
			name: "success (tooth 正方形、DisposalBackground)",
			args: args{
				ctx:    context.Background(),
				src:    "tooth.gif",
				width:  256,
				height: 256,
				opts: []resigif.Option{
					resigif.WithParallel(1),
				},
			},
			want:      "tooth_resized.gif",
			assertion: require.NoError,
		},
		{
			name: "success (new_year 横長)",
			args: args{
				ctx:    context.Background(),
				src:    "new_year.gif",
				width:  256,
				height: 256,
				opts: []resigif.Option{
					resigif.WithImageResizeFunc(resigif.FromDrawScaler(draw.BiLinear)),
				},
			},
			want:      "new_year_resized.gif",
			assertion: require.NoError,
		},
		{
			name: "success (miku 縦長、差分最適化)",
			args: args{
				ctx:    context.Background(),
				src:    "miku.gif",
				width:  256,
				height: 256,
				opts:   nil,
			},
			want:      "miku_resized.gif",
			assertion: require.NoError,
		},
		{
			name: "success (frog 縦長、DisposalBackground + 背景色不整合)",
			args: args{
				ctx:    context.Background(),
				src:    "frog.gif",
				width:  256,
				height: 256,
				opts:   nil,
			},
			want:      "frog_resized.gif",
			assertion: require.NoError,
		},
		{
			name: "success (surprised 正方形、空のGlobal Color Table)",
			args: args{
				ctx:    context.Background(),
				src:    "surprised.gif",
				width:  256,
				height: 256,
				opts:   nil,
			},
			want:      "surprised_resized.gif",
			assertion: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := resigif.Resize(tt.args.ctx, mustOpenGif(t, tt.args.src), tt.args.width, tt.args.height, tt.args.opts...)

			if overWrite != nil && *overWrite {
				mustEncodeGif(t, tt.want, got)
			}

			tt.assertion(t, err)
			assert.Equal(t, mustOpenGif(t, tt.want), got)
		})
	}
}
