package image

import (
	"github.com/rstms/go-fs/fs"
	"os"
)

type Floppy struct {
	file *os.File
	fd   *fs.FileDisk
}

const KB = 1024
const KBSize = 1440
const ImageSize = KBSize * KB

func createBackingFile(filename) error {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	err = file.Truncate(ImageSize)
	if err != nil {
		return err
	}
	return nil
}

func NewImage(filename string) (*FdImage, error) {

	f = Floppy{}
	err := createBackingFile(filename)
	if err != nil {
		return nil, error
	}
	return &f, nil
}
