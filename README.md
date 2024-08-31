# resigif

[![CI Pipeline](https://github.com/logica0419/resigif/actions/workflows/ci.yml/badge.svg)](https://github.com/logica0419/resigif/actions/workflows/ci.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/logica0419/resigif.svg)](https://pkg.go.dev/github.com/logica0419/resigif) [![license](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/logica0419/resigif/blob/main/LICENSE)

Animated GIF resizing library w/o cgo nor any third-party Libraries

## Installation

You can install resigif with the `go get` command

```sh
go get -u github.com/logica0419/resigif
```

## Quick Start

The only API of this library is `Resize()` function.  
You can easily resize an animated GIF image by passing `*gif.GIF` struct and the target size.

Here's a simple example:

```go
package main

import (
  "context"
  "image/gif"
  "os"

  "github.com/logica0419/resigif"
)

func main() {
  ctx := context.Background()

  src, err := os.Open("image.gif")
  if err != nil {
    panic(err)
  }
  defer src.Close()

  srcImg, err := gif.DecodeAll(src)
  if err != nil {
    panic(err)
  }

  width := 480
  height := 360

  dstImg, err := resigif.Resize(ctx, srcImg, width, height)
  if err != nil {
    panic(err)
  }

  dst, err := os.OpenFile("resized.gif", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
  if err != nil {
    panic(err)
  }
  defer dst.Close()

  err = gif.EncodeAll(dst, dstImg)
  if err != nil {
    panic(err)
  }
}

```

## Customization

- Aspect Ratio Preservation
  - You can choose from `Ignore` or `Maintain`

```go
dstImg, err := resigif.Resize(
  ctx,
  srcImg,
  width,
  height,
  resigif.WithAspectRatio(resigif.Maintain),
)

dstImg, err := resigif.Resize(
  ctx,
  srcImg,
  width,
  height,
  resigif.WithAspectRatio(resigif.Ignore),
)
```

- Resizing algorithm
  - You can use you own resizing algorithm by implementing `ImageResizeFunc` interface and passing it to `WithImageResizeFunc()`
  - If you want to use `golang.org/x/image/draw.Scaler`, you can use `FromDrawScaler()` to convert it to `ImageResizeFunc`

```go
dstImg, err := resigif.Resize(
  ctx,
  srcImg,
  width,
  height,
  resigif.WithImageResizeFunc(resigif.FromDrawScaler(draw.BiLinear)),
)
```

- Parallelism
  - You can control the number of goroutine used for resizing by passing `WithParallel()`

```go
dstImg, err := resigif.Resize(
  ctx,
  srcImg,
  width,
  height,
  resigif.WithParallel(3),
)
```
