package util

import (
	"io"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
)

type FileWriter struct {
	fileSize    int64
	file        *os.File
	multiWriter io.Writer
}

func NewFileWriter(file *os.File, fileSize int64, showProgressBar bool) *FileWriter {
	mw := io.MultiWriter(file)
	if showProgressBar && ShowProgressBars && fileSize > 0 {
		// Create a progress bar
		bar := progressbar.NewOptions64(
			int64(fileSize),
			progressbar.OptionSetDescription("Progress: "),
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionShowBytes(true),
			progressbar.OptionShowTotalBytes(true),
			progressbar.OptionThrottle(65*time.Millisecond),
			progressbar.OptionShowCount(),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionFullWidth(),
			progressbar.OptionSetRenderBlankState(true),
			progressbar.OptionClearOnFinish(),
		)

		mw = io.MultiWriter(file, bar)
	}

	return &FileWriter{
		fileSize:    fileSize,
		file:        file,
		multiWriter: mw,
	}
}

func (w *FileWriter) Write(b []byte) (int, error) {
	return w.multiWriter.Write(b)
}
