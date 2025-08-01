package image

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func initTestConfig(t *testing.T) {
	viper.SetConfigFile("testdata/config.yaml")
	err := viper.ReadInConfig()
	require.Nil(t, err)
}

func TestCreate(t *testing.T) {
	initTestConfig(t)
	err := CreateFdImage("testdata/image", "test", "fdimage")
	require.Nil(t, err)
}

func createAndOpen(t *testing.T) *FdImage {
	TestCreate(t)
	fd, err := OpenFdImage("testdata/image")
	require.Nil(t, err)
	return fd
}

func TestOpen(t *testing.T) {
	fd, err := OpenFdImage("testdata/image")
	require.Nil(t, err)
	require.IsType(t, &FdImage{}, fd)
	err = fd.Close()
	require.Nil(t, err)
}

func TestMkdir(t *testing.T) {
	fd := createAndOpen(t)
	defer fd.Close()
	err := fd.Mkdir("foo")
	require.Nil(t, err)
	err = fd.Mkdir("bar")
	require.Nil(t, err)
	err = fd.Mkdir("baz")
	require.Nil(t, err)

	list, err := fd.List("", false)
	require.Nil(t, err)
	require.Equal(t, []string{"FOO/", "BAR/", "BAZ/"}, list)
}

func TestNestedMkdir(t *testing.T) {
	fd := createAndOpen(t)
	err := fd.Mkdir("foo")
	require.Nil(t, err)
	err = fd.Mkdir("foo/bar")
	require.Nil(t, err)
	err = fd.Mkdir("foo/bar/baz")
	require.Nil(t, err)

	// write a file to the root
	err = writeFile(fd, "sample", "/sample")
	require.Nil(t, err)

	// write a file under the foo directory
	err = writeFile(fd, "sample", "/foo/sample1")
	require.Nil(t, err)

	// write a file under the foo/bar directory
	err = writeFile(fd, "sample", "/foo/bar/sample2")
	require.Nil(t, err)

	// write a file under the foo/bar/baz directory
	err = writeFile(fd, "sample", "/foo/bar/baz/sample3")
	require.Nil(t, err)

	err = fd.Close()
	require.Nil(t, err)

	rfd, err := OpenFdImage("testdata/image")

	err = readFile(rfd, "/foo/bar/baz/sample3", "sample3", "sample")
	require.Nil(t, err)

	err = readFile(rfd, "/foo/bar/sample2", "sample2", "sample")
	require.Nil(t, err)

	err = readFile(rfd, "/foo/sample1", "sample1", "sample")
	require.Nil(t, err)

	err = rfd.Close()
	require.Nil(t, err)
}

func writeFile(fd *FdImage, src, target string) error {
	data, err := os.ReadFile(filepath.Join("testdata", src))
	if err != nil {
		return err
	}
	count, err := fd.WriteFile(target, data)
	if int64(len(data)) != count {
		return fmt.Errorf("write count mismatch: %d %d\n", len(data), count)
	}
	return nil
}

func readFile(fd *FdImage, src, target, reference string) error {

	data, err := fd.ReadFile(src)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join("testdata", target), data, 0600)
	if err != nil {
		return err
	}
	err = compareFiles(filepath.Join("testdata", target), filepath.Join("testdata", reference))
	if err != nil {
		return err
	}
	return nil
}

func compareFiles(first, second string) error {
	fdata, err := os.ReadFile(first)
	if err != nil {
		return err
	}
	sdata, err := os.ReadFile(second)
	if err != nil {
		return err
	}
	if len(fdata) != len(sdata) {
		return fmt.Errorf("size mismatch: %d != %d", len(fdata), len(sdata))
	}
	for i, fb := range fdata {
		if fb != sdata[i] {
			return fmt.Errorf("difference at index %d", i)
		}
	}
	return nil
}

//func TestSetDir(t *testing.T) {
//}

func TestWriteFile(t *testing.T) {
	fd := createAndOpen(t)
	defer fd.Close()
	data, err := os.ReadFile("testdata/sample")
	require.Nil(t, err)
	count, err := fd.WriteFile("sample", data)
	require.Nil(t, err)
	require.Equal(t, int64(len(data)), count)
}

func TestReadFile(t *testing.T) {
	TestWriteFile(t)
	fd, err := OpenFdImage("testdata/sample")
	require.Nil(t, err)
	defer fd.Close()
	fdata, err := fd.ReadFile("sample")
	require.Nil(t, err)
	tdata, err := os.ReadFile("testdata/sample")
	require.Nil(t, err)
	require.Equal(t, len(tdata), len(fdata))
	require.Equal(t, tdata, fdata)
}

/*
func Open(filename string) (*FdImage, error) {
func (f *FdImage) Mkdir(path string) error {
func (f *FdImage) Chdir(path string) error {
func (f *FdImage) setRootDir() error {
func (f *FdImage) setDir(filename string) (string, error) {
func (f *FdImage) WriteFile(filename string, data []byte) (int64, error) {
func (f *FdImage) ReadFile(filename string) ([]byte, error) {
func (f *FdImage) List(pathname string, longFlag bool) ([]string, error) {
func (f *FdImage) Close() error {
*/
