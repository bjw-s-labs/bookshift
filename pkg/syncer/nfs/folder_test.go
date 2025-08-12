package nfs

import (
	"testing"

	"github.com/kha7iq/go-nfs-client/nfs4"
)

// TestNfsFolderFetchFiles_Recurse verifies recursive traversal discovers files in subfolders.
func TestNfsFolderFetchFiles_Recurse(t *testing.T) {
	root := "/root"
	sub := "/root/sub"
	fake := &fakeNfs{
		files: map[string][]nfs4.FileInfo{
			root: {{Name: "sub", IsDir: true}},
			sub:  {{Name: "b.kepub"}},
		},
		reads: map[string]string{sub + "/b.kepub": "data"},
	}
	folder := NewNfsFolder(root, fake)
	files, err := folder.FetchFiles(root, []string{".kepub"}, true)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("want 1, got %d", len(files))
	}
}

// TestNfsFolderFetchFiles ensures FetchFiles filters by extension and returns File abstractions.
func TestNfsFolderFetchFiles(t *testing.T) {
	root := "/root"
	fake := &fakeNfs{
		files: map[string][]nfs4.FileInfo{
			root: {
				{Name: "a.epub"},
				{Name: "skip.txt"},
			},
		},
		reads: map[string]string{"/root/a.epub": "content"},
	}

	folder := NewNfsFolder(root, fake)
	files, err := folder.FetchFiles(root, []string{".epub"}, false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("got %d", len(files))
	}
}
