package rpc

import (
	"context"
	"testing"

	"github.com/iamxvbaba/td/tg"

	"telesrv/internal/domain"
)

// A client derives the requested photo size type from its own ladder, not from
// what the server advertised. Asking for a type that was never stored used to
// fail with LOCATION_INVALID, which leaves the image on a permanent spinner in
// DrKLO. The fallback must answer with a real rendition instead.
func TestUploadGetFileFallsBackToStoredPhotoSize(t *testing.T) {
	ctx := context.Background()
	files := &fakeFiles{
		getFileFound:      true,
		getFileChunk:      domain.FileChunk{Bytes: []byte("pixels"), Total: 6},
		getFileMissingFor: map[string]bool{"photo:9001:y": true},
	}
	r := &Router{deps: Deps{Files: files}}

	_, err := r.onUploadGetFile(WithUserID(ctx, 1001), &tg.UploadGetFileRequest{
		Location: &tg.InputPhotoFileLocation{ID: 9001, AccessHash: 7, ThumbSize: "y"},
		Limit:    1024,
	})
	if err != nil {
		t.Fatalf("getFile for a missing size type = %v, want a fallback rendition", err)
	}
	if files.getFileRequest.LocationKey == "photo:9001:y" {
		t.Fatalf("fallback kept the missing location %q", files.getFileRequest.LocationKey)
	}
	if start := "photo:9001:"; len(files.getFileRequest.LocationKey) <= len(start) ||
		files.getFileRequest.LocationKey[:len(start)] != start {
		t.Fatalf("fallback key = %q, want a photo size of the same photo", files.getFileRequest.LocationKey)
	}
}

// When no rendition of the photo exists at all, the original error must survive.
func TestUploadGetFileKeepsLocationInvalidWhenNoPhotoSizeExists(t *testing.T) {
	ctx := context.Background()
	files := &fakeFiles{
		getFileFound: false,
		getFileMissingFor: map[string]bool{
			"photo:9002:s": true, "photo:9002:m": true, "photo:9002:x": true,
			"photo:9002:y": true, "photo:9002:w": true,
		},
	}
	r := &Router{deps: Deps{Files: files}}

	_, err := r.onUploadGetFile(WithUserID(ctx, 1001), &tg.UploadGetFileRequest{
		Location: &tg.InputPhotoFileLocation{ID: 9002, AccessHash: 7, ThumbSize: "y"},
		Limit:    1024,
	})
	if err == nil {
		t.Fatal("expected LOCATION_INVALID when the photo has no stored rendition")
	}
}
