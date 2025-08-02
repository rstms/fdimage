package image

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"log"
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

func TestEFICreate(t *testing.T) {
	efiImage := filepath.Join("testdata", "efi.img")
	efiBootFile := filepath.Join("testdata", "bootx64.efi")
	efiName := "BOOTX64.EFI"
	extraFiles := []string{filepath.Join("testdata", "autoexec.ipxe")}

	err := CreateEFIImage(efiImage, efiBootFile, efiName, extraFiles)
	require.Nil(t, err)
}

func TestISOList(t *testing.T) {
	isoImage := filepath.Join("testdata", "netboot.xyz.iso")
	files, err := ListImageFiles(isoImage)
	require.Nil(t, err)
	require.IsType(t, []string{}, files)
	for _, file := range files {
		require.IsType(t, "", file)
		log.Println(file)
	}
}

func TestIMGList(t *testing.T) {
	imageFile := filepath.Join("testdata", "esp.img")
	files, err := ListImageFiles(imageFile)
	require.Nil(t, err)
	require.IsType(t, []string{}, files)
	for _, file := range files {
		require.IsType(t, "", file)
		log.Println(file)
	}
}

func mkTestDir(t *testing.T, name string) string {
	testDir := filepath.Join("testdata", name)
	stat, err := os.Stat(testDir)
	if err == nil {
		if stat.IsDir() {
			err := os.RemoveAll(testDir)
			require.Nil(t, err)
		}
	}
	err = os.Mkdir(testDir, 0700)
	require.Nil(t, err)
	return testDir
}

func TestISOExtract(t *testing.T) {
	imageFile := filepath.Join("testdata", "netboot.xyz.iso")
	destDir := mkTestDir(t, "isofiles")
	err := ExtractImageFiles(imageFile, destDir)
	require.Nil(t, err)
}

func TestIMGExtract(t *testing.T) {
	imageFile := filepath.Join("testdata", "esp.img")
	destDir := mkTestDir(t, "imgfiles")
	err := ExtractImageFiles(imageFile, destDir)
	require.Nil(t, err)
}

func TestImageInfo(t *testing.T) {
	isoImage := filepath.Join("testdata", "netboot.xyz.iso")
	name, size, err := ImageInfo(isoImage)
	require.Nil(t, err)
	log.Printf("iso=%s name=%s size=%d\n", isoImage, name, size)
}

func TestISOCreate(t *testing.T) {
	outputImage := filepath.Join("testdata", "output.iso")
	sourceImage := filepath.Join("testdata", "netboot.xyz.iso")
	autoexecFile := filepath.Join("testdata", "autoexec.ipxe")
	err := CreateISOImage(outputImage, sourceImage, autoexecFile)
	require.Nil(t, err)
}
