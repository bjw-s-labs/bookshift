package util

import (
	"os"
	"testing"
)

// TestFileWriter_Write ensures NewFileWriter writes bytes to the underlying
// writer and that the file size reflects the number of bytes written.
func TestFileWriter_Write(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "fw-")
	if err != nil {
		t.Fatalf("temp: %v", err)
	}
	defer f.Close()

	w := NewFileWriter(f, 0, false)
	n, err := w.Write([]byte("hello"))
	if err != nil || n != 5 {
		t.Fatalf("write n=%d err=%v", n, err)
	}

	// ensure content was written
	st, _ := f.Stat()
	if st.Size() != 5 {
		t.Fatalf("size=%d", st.Size())
	}
}
