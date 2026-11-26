package files

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"testing"

	"telesrv/internal/domain"
)

func TestEncodeStrippedThumbInflatesToDecodableJPEG(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 80, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 80; x++ {
			src.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y * 4), B: 90, A: 255})
		}
	}
	packed, w, h, err := encodeStrippedThumb(src)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if packed[0] != strippedThumbMarker || w != 40 || h != 25 || packed[1] != byte(h) || packed[2] != byte(w) {
		t.Fatalf("packed header = %v size %dx%d, want marker 1 and 40x25", packed[:3], w, h)
	}
	jpg, err := inflateStrippedThumb(packed)
	if err != nil {
		t.Fatalf("inflate: %v", err)
	}
	decoded, err := jpeg.Decode(bytes.NewReader(jpg))
	if err != nil {
		t.Fatalf("jpeg.Decode inflated thumb: %v", err)
	}
	if decoded.Bounds().Dx() != w || decoded.Bounds().Dy() != h {
		t.Fatalf("decoded size = %dx%d, want %dx%d", decoded.Bounds().Dx(), decoded.Bounds().Dy(), w, h)
	}
	size, ok := avatarStrippedSize(src)
	if !ok || size.Type != "i" || size.Kind != domain.PhotoSizeKindStripped || !bytes.Equal(size.Bytes, packed) {
		t.Fatalf("avatarStrippedSize ok=%v size=%+v", ok, size)
	}
}

func TestCreateAvatarFromBytesIncludesOfficialStrippedThumb(t *testing.T) {
	media := newFakeMediaStore()
	blobs, err := NewLocalFS(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	svc := NewService(media, blobs, 2)
	photo, err := svc.CreateAvatarFromBytes(context.Background(), testJPEG(t, 160, 160))
	if err != nil {
		t.Fatal(err)
	}
	stripped := domain.StrippedFromSizes(photo.Sizes)
	if len(stripped) < 4 || stripped[0] != strippedThumbMarker {
		t.Fatalf("avatar photo missing stripped thumb, sizes=%+v", photo.Sizes)
	}
	jpg, err := inflateStrippedThumb(stripped)
	if err != nil {
		t.Fatalf("inflate stored thumb: %v", err)
	}
	if _, err := jpeg.Decode(bytes.NewReader(jpg)); err != nil {
		t.Fatalf("stored stripped thumb is not a valid official thumbnail: %v", err)
	}
}
