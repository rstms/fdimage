package image

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

var imageFile string

func initTestConfig(t *testing.T) {
	viper.SetConfigFile("testdata/config.yaml")
	err := viper.ReadInConfig()
	require.Nil(t, err)
}

func TestCreate(t *testing.T) {
	imageFile = filepath.Join("testdata", "image")
	initTestConfig(t)
	err := CreateFdImage(imageFile, "test", "fdimage")
	require.Nil(t, err)
}

func TestOpen(t *testing.T) {
	TestCreate(t)
	fd, err := OpenFdImage(imageFile)
	require.Nil(t, err)
	require.IsType(t, &FdImage{}, fd)
	err = fd.Close()
	require.Nil(t, err)
}

func TestMkdir(t *testing.T) {
	TestCreate(t)
	err := Mkdir(imageFile, "foo")
	require.Nil(t, err)
	err = Mkdir(imageFile, "bar")
	require.Nil(t, err)
	err = Mkdir(imageFile, "baz")
	require.Nil(t, err)

	list, err := List(imageFile, "", false)
	require.Nil(t, err)
	require.Equal(t, []string{"FOO/", "BAR/", "BAZ/"}, list)
}

func TestNestedMkdir(t *testing.T) {

	TestCreate(t)
	err := Mkdir(imageFile, "foo")
	require.Nil(t, err)

	//err = Mkdir(imageFile, "foo/bar")
	//require.Nil(t, err)

	//err = Mkdir(imageFile, "foo/bar/baz")
	//require.Nil(t, err)

	// write a file to the root
	err = writeFile("sample", "/sample")
	require.Nil(t, err)

	// write a file under the foo directory
	//err = writeFile("sample", "/foo/sample1")
	//require.Nil(t, err)

	// write a file under the foo/bar directory
	//err = writeFile("sample", "/foo/bar/sample2")
	//require.Nil(t, err)

	// write a file under the foo/bar/baz directory
	//err = writeFile("sample", "/foo/bar/baz/sample3")
	//require.Nil(t, err)

	//err = readFile("/foo/bar/baz/sample3", "sample3", "sample")
	//require.Nil(t, err)

	//err = readFile("/foo/bar/sample2", "sample2", "sample")
	//require.Nil(t, err)

	//err = readFile("/foo/sample1", "sample1", "sample")
	//require.Nil(t, err)
}

func writeFile(src, target string) error {
	data, err := os.ReadFile(filepath.Join("testdata", src))
	if err != nil {
		return err
	}
	count, err := WriteFile(imageFile, target, data)
	if int64(len(data)) != count {
		return fmt.Errorf("write count mismatch: %d %d\n", len(data), count)
	}
	return nil
}

func readFile(src, target, reference string) error {

	data, err := ReadFile(imageFile, src)
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
	TestCreate(t)
	data, err := os.ReadFile("testdata/sample")
	require.Nil(t, err)
	count, err := WriteFile(imageFile, "sample", data)
	require.Nil(t, err)
	require.Equal(t, int64(len(data)), count)
}

func TestReadFile(t *testing.T) {
	TestWriteFile(t)
	fdata, err := ReadFile(imageFile, "sample")
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
