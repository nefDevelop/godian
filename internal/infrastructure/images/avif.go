//go:build avif

package images

import (
	"image"
	"io"

	"github.com/gen2brain/avif"
)

// avifSupported indica si se compiló con soporte AVIF (-tags avif).
const avifSupported = true

func encodeAVIF(w io.Writer, img image.Image) error {
	return avif.Encode(w, img, avif.Options{Quality: 50})
}
