package image

import (
	"bytes"
	"fmt"
	"github.com/rstms/fdimage/fat"
	"github.com/rstms/fdimage/fs"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type FdImage struct {
	Filename   string
	FileDisk   *fs.FileDisk
	FileSystem *fat.FileSystem
	dir        fs.Directory
	cwd        string
}

func (f *FdImage) Close() error {
	fmt.Printf("Close()\n")
	return f.FileDisk.Close()
}

const KB = 1024
const KBSize = 1440
const ImageSize = KBSize * KB

func isFile(pathname string) bool {
	_, err := os.Stat(pathname)
	return !os.IsNotExist(err)
}

func createBackingFile(filename string) (*os.File, error) {
	fmt.Printf("createBackingFile(%s)\n", filename)
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	err = file.Truncate(ImageSize)
	if err != nil {
		file.Close()
		return nil, err
	}
	return file, nil
}

func CreateFdImage(filename, label, name string) error {
	fmt.Printf("CreateFdImage(%s, %s, %s)\n", filename, label, name)
	file, err := createBackingFile(filename)
	if err != nil {
		return err
	}
	fileDisk, err := fs.NewFileDisk(file)
	if err != nil {
		file.Close()
		return err
	}
	defer fileDisk.Close()
	config := fat.SuperFloppyConfig{
		FATType: fat.FAT12,
		Label:   label,
		OEMName: name,
	}
	err = fat.FormatSuperFloppy(fileDisk, &config)
	if err != nil {
		return err
	}
	return nil
}

func OpenFdImage(filename string) (*FdImage, error) {
	fmt.Printf("OpenFdImage(%s)\n", filename)
	if !isFile(filename) {
		return nil, fmt.Errorf("file not found: %s", filename)
	}
	file, err := os.OpenFile(filename, os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	fileDisk, err := fs.NewFileDisk(file)
	if err != nil {
		file.Close()
		return nil, err
	}
	fileSystem, err := fat.New(fileDisk)
	if err != nil {
		return nil, err
	}
	fd := FdImage{
		Filename:   filename,
		FileDisk:   fileDisk,
		FileSystem: fileSystem,
	}
	err = setRootDir(&fd)
	if err != nil {
		return nil, err
	}
	return &fd, nil
}

func Mkdir(image, path string) error {
	fmt.Printf("Mkdir(%s, %s)\n", image, path)
	fd, err := OpenFdImage(image)
	if err != nil {
		return err
	}
	defer fd.Close()
	name, err := setDir(fd, path)
	if err != nil {
		return err
	}
	fmt.Printf("Adding Directory: %s cwd=%s %+v\n", name, fd.cwd, fd.dir)
	_, err = fd.dir.AddDirectory(name)
	if err != nil {
		return err
	}
	return nil
}

func chdir(fd *FdImage, path string) error {
	fmt.Printf("Chdir(%s)\n", path)
	parts := []string{}
	var tail string
	for path != "" {
		path, tail = filepath.Split(path)
		path = strings.TrimRight(path, string(filepath.Separator))
		if tail != "" {
			parts = append(parts, tail)
		}
	}
	err := setRootDir(fd)
	if err != nil {
		return err
	}
	for i, part := range parts {
		fmt.Printf("part[%d] = %s\n", i, part)
	}
	for i := len(parts) - 1; i >= 0; i-- {
		name := parts[i]
		fmt.Printf("setting dir %d %s\n", i, name)
		entry := fd.dir.Entry(name)
		if entry == nil {
			return fmt.Errorf("not found: %s", name)
		}
		if !entry.IsDir() {
			return fmt.Errorf("not a directory: %s", name)
		}
		dir, err := entry.Dir()
		if err != nil {
			return err
		}
		fd.dir = dir
		if fd.cwd == "/" {
			fd.cwd += entry.Name()
		} else {
			fd.cwd += "/" + entry.Name()
		}
		fmt.Printf("new cwd=%s %+v %+v\n", fd.cwd, entry, dir)
	}
	return nil
}

func setRootDir(fd *FdImage) error {
	rootDir, err := fd.FileSystem.RootDir()
	if err != nil {
		return err
	}
	fd.dir = rootDir
	fd.cwd = "/"
	return nil
}

func setDir(fd *FdImage, filename string) (string, error) {
	fmt.Printf("setDir(%s)\n", filename)
	if filename == "" {
		return "", nil
	}
	if filename == string(filepath.Separator) {
		return "", setRootDir(fd)
	}
	path, name := filepath.Split(filename)
	if path != "" {
		path = strings.TrimRight(path, string(filepath.Separator))
		err := chdir(fd, path)
		if err != nil {
			return "", err
		}
	}
	fmt.Printf("setDir returning %s cwd=%s\n", name, fd.cwd)
	return name, nil
}

func WriteFile(image, filename string, data []byte) (int64, error) {
	fmt.Printf("WriteFile(%s, %s [%d bytes])\n", image, filename, len(data))
	fd, err := OpenFdImage(image)
	if err != nil {
		return 0, err
	}
	defer fd.Close()
	name, err := setDir(fd, filename)
	if err != nil {
		return 0, err
	}
	entry, err := fd.dir.AddFile(name)
	if err != nil {
		return 0, err
	}
	file, err := entry.File()
	if err != nil {
		return 0, err
	}
	buf := bytes.NewBuffer(data)
	count, err := io.Copy(file, buf)
	if err != nil {
		return 0, err
	}
	fmt.Printf("wrote %d bytes to %s\n", count, filename)
	return count, nil
}

func ReadFile(image, filename string) ([]byte, error) {
	fmt.Printf("ReadFile(%s)\n", filename)
	fd, err := OpenFdImage(image)
	if err != nil {
		return []byte{}, err
	}
	defer fd.Close()
	name, err := setDir(fd, filename)
	if err != nil {
		return []byte{}, err
	}
	entry := fd.dir.Entry(name)
	if entry == nil {
		return []byte{}, fmt.Errorf("file not found: %s", filename)
	}

	fileSize, err := entry.FileSize()
	if err != nil {
		return []byte{}, err
	}
	fmt.Printf("fileSize: %d\n", fileSize)

	file, err := entry.File()
	if err != nil {
		return []byte{}, err
	}

	buf := make([]byte, fileSize)
	count, err := file.Read(buf)
	switch err {
	case nil:
	case io.EOF:
	default:
		return []byte{}, err
	}
	fmt.Printf("read %d bytes from %s\n", count, filename)
	if uint32(count) != fileSize {
		return []byte{}, fmt.Errorf("read underrun: expected=%d read=%d\n", fileSize, count)
	}
	return buf, nil
}

/*
const BUFSIZE = 8192

func ReadFile(filename string) ([]byte, error) {
	fmt.Printf("ReadFile(%s)\n", filename)
	name, err := f.setDir(filename)
	if err != nil {
		return []byte{}, err
	}
	entry := f.dir.Entry(name)
	if entry == nil {
		return []byte{}, fmt.Errorf("file not found: %s", filename)
	}

	if entry.IsDir() {
		return []byte{}, fmt.Errorf("attempted read of directory as file: %s", filename)

	}
	fmt.Printf("DirectoryEntry: %+v\n", entry)

	var fatDirectoryEntry *fat.DirectoryEntry
	fatDirectoryEntry = entry.(*fat.DirectoryEntry)
	fmt.Printf("fatDirectoryEntry: %+v\n", fatDirectoryEntry)

	fileSize, err := entry.FileSize()
	if err != nil {
		return []byte{}, err
	}
	fmt.Printf("fileSize: %d\n", fileSize)

	file, err := entry.File()
	if err != nil {
		return []byte{}, err
	}

	fmt.Printf("File: %+v\n", file)

	//var fatFile *fat.File
	//fatFile = file.(*fat.File)
	//fmt.Printf("FileEntry: %+v\n", fatFile.entry)

	var chunkTotal int64
	var buf bytes.Buffer
	for done := false; !done; {
		chunk := make([]byte, BUFSIZE)
		count, err := file.Read(chunk)
		switch err {
		case nil:
		case io.EOF:
			fmt.Printf("EOF returned after reading %d bytes\n", chunkTotal)
			done = true
		default:
			return []byte{}, err
		}
		fmt.Printf("  read chunk of %d bytes\n", count)
		chunkTotal += int64(count)
		wcount, err := buf.Write(chunk[:count])
		if err != nil {
			return []byte{}, err
		}
		if wcount != count {
			fmt.Errorf("chunk write incomplete: %d != %d", wcount, count)
		}
	}
	bufLen := buf.Len()
	data := buf.Bytes()
	dataLen := len(data)
	if dataLen != bufLen {
		fmt.Errorf("buffer length (%d) mismatches data length (%d)", bufLen, dataLen)
	}
	if int64(dataLen) != chunkTotal {
		fmt.Errorf("chunk total (%d) mismatches data length (%d)", chunkTotal, dataLen)
	}
	fmt.Printf("read %d bytes from %s\n", dataLen, filename)
	return data, nil
}
*/

func List(image, pathname string, longFlag bool) ([]string, error) {
	fmt.Printf("List(%s, %s, %v)\n", image, pathname, longFlag)
	names := []string{}
	fd, err := OpenFdImage(image)
	if err != nil {
		return names, err
	}
	defer fd.FileDisk.Close()
	name, err := setDir(fd, pathname)
	if err != nil {
		return names, err
	}
	fmt.Printf("name=%s cwd=%s\n", name, fd.cwd)
	var entries []fs.DirectoryEntry
	if name == "" {
		entries = fd.dir.Entries()
	} else {
		entry := fd.dir.Entry(name)
		if entry == nil {
			return names, fmt.Errorf("not found: %s", pathname)
		}
		if entry.IsDir() {
			fmt.Printf("entry is Dir %+v\n", entry)
			dir, err := entry.Dir()
			if err != nil {
				return names, err
			}
			fd.dir = dir
			entries = dir.Entries()
		} else {
			fmt.Printf("entry is File %+v\n", entry)
			entries = []fs.DirectoryEntry{entry}
		}
	}
	for _, entry := range entries {
		name := entry.Name()
		shortName := entry.ShortName()
		if entry.IsDir() {
			name += "/"
		}
		if longFlag {
			names = append(names, fmt.Sprintf("%s\t%s\t%+v", name, shortName, entry))
		} else {
			names = append(names, name)
		}
	}
	return names, nil
}
