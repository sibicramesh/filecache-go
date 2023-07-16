package filecache

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

// Tar gzips and tars encoded data retrieved from
// the filecaches and writes it to w.
func Tar(w io.Writer, fileCaches ...Queue) error {

	gzipWriter := gzip.NewWriter(w)
	defer gzipWriter.Close() // nolint: errcheck

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close() // nolint: errcheck

	for _, f := range fileCaches {

		partReader, partSize, err := f.EncodeToReader()
		if err != nil {
			if err == ErrNilType {
				continue
			}

			return err
		}
		closeReader := func() {
			if partCloser, ok := partReader.(io.Closer); ok {
				if errClose := partCloser.Close(); errClose != nil {
					zap.L().Debug("failed to close file during tar for "+f.Name(), zap.Error(errClose))
				}
			}
		}

		header := &tar.Header{
			Name:    f.Name(),
			Size:    partSize,
			Mode:    0666,
			ModTime: time.Now(),
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			closeReader()
			return err
		}

		if _, err = io.Copy(tarWriter, partReader); err != nil {
			closeReader()
			return err
		}

		closeReader()
	}

	return nil
}

// Untar reads data from the reader r and writes it
// to the destDir as independent files. It is upto
// the caller to make sure directory exists.
func Untar(destDir string, r io.Reader) error {

	gzipReader, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzipReader.Close() // nolint: errcheck

	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()
		switch err {
		case nil:
			// Do nothing.
		case io.EOF:
			return nil
		default:
			return err
		}

		dest := filepath.Join(destDir, header.Name)
		file, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, header.FileInfo().Mode())
		if err != nil {
			return err
		}

		if _, err = io.Copy(file, tarReader); err != nil {
			file.Close() // nolint: errcheck
			return err
		}
		file.Close() // nolint: errcheck
	}
}
