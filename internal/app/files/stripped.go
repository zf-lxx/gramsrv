package files

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"

	"telesrv/internal/domain"

	xdraw "golang.org/x/image/draw"
)

// Official clients inflate photoStrippedSize (type "i") back to a JPEG using a
// fixed 623-byte JFIF prefix, then read height from bytes[1] and width from
// bytes[2]. See https://core.telegram.org/api/files#stripped-thumbnails.
const (
	strippedThumbMarker      = 0x01
	strippedThumbMaxSide     = 40
	strippedThumbJPEGQuality = 30
)

func avatarStrippedSize(src image.Image) (domain.PhotoSize, bool) {
	payload, w, h, err := encodeStrippedThumb(src)
	if err != nil || len(payload) < 4 || w <= 0 || h <= 0 {
		return domain.PhotoSize{}, false
	}
	return domain.PhotoSize{
		Kind:  domain.PhotoSizeKindStripped,
		Type:  "i",
		W:     w,
		H:     h,
		Size:  len(payload),
		Bytes: payload,
	}, true
}

func encodeStrippedThumb(src image.Image) ([]byte, int, int, error) {
	if src == nil {
		return nil, 0, 0, fmt.Errorf("nil image")
	}
	thumb := scaleStrippedThumb(src)
	bounds := thumb.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w <= 0 || h <= 0 || w > 255 || h > 255 {
		return nil, 0, 0, fmt.Errorf("invalid stripped thumb size %dx%d", w, h)
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, thumb, &jpeg.Options{Quality: strippedThumbJPEGQuality}); err != nil {
		return nil, 0, 0, err
	}
	scan, ok := jpegScanData(buf.Bytes())
	if !ok {
		return nil, 0, 0, fmt.Errorf("jpeg scan data not found")
	}
	out := make([]byte, 0, 3+len(scan))
	out = append(out, strippedThumbMarker, byte(h), byte(w))
	out = append(out, scan...)
	return out, w, h, nil
}

func scaleStrippedThumb(src image.Image) image.Image {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w <= 0 || h <= 0 {
		return src
	}
	longSide := w
	if h > longSide {
		longSide = h
	}
	if longSide <= strippedThumbMaxSide {
		return src
	}
	tw := maxInt(1, w*strippedThumbMaxSide/longSide)
	th := maxInt(1, h*strippedThumbMaxSide/longSide)
	dst := image.NewRGBA(image.Rect(0, 0, tw, th))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, bounds, xdraw.Src, nil)
	return dst
}

func jpegScanData(data []byte) ([]byte, bool) {
	sos := -1
	for i := 0; i < len(data)-1; i++ {
		if data[i] == 0xFF && data[i+1] == 0xDA {
			sos = i
			break
		}
	}
	if sos < 0 || sos+4 >= len(data) {
		return nil, false
	}
	headerLen := int(data[sos+2])<<8 | int(data[sos+3])
	start := sos + 2 + headerLen
	end := len(data)
	if end >= 2 && data[end-2] == 0xFF && data[end-1] == 0xD9 {
		end -= 2
	}
	if start >= end {
		return nil, false
	}
	return data[start:end], true
}

func inflateStrippedThumb(packed []byte) ([]byte, error) {
	if len(packed) < 4 || packed[0] != strippedThumbMarker || len(strippedJPEGHeader) < 168 {
		return nil, fmt.Errorf("invalid stripped thumb")
	}
	out := make([]byte, 0, len(strippedJPEGHeader)+len(packed))
	out = append(out, strippedJPEGHeader[:164]...)
	out = append(out, packed[1], strippedJPEGHeader[165], packed[2])
	out = append(out, strippedJPEGHeader[167:]...)
	out = append(out, packed[3:]...)
	out = append(out, 0xFF, 0xD9)
	return out, nil
}
