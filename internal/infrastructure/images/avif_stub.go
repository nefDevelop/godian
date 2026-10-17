//go:build !avif

package images

import (
	"image"
	"io"
)

// avifSupported indica si se compiló con soporte AVIF (-tags avif).
const avifSupported = false

func encodeAVIF(w io.Writer, img image.Image) error { return nil }
