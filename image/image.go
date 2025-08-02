package image

import (
	"fmt"
	diskfs "github.com/rstms/go-diskfs"
	diskpkg "github.com/rstms/go-diskfs/disk"
	"github.com/rstms/go-diskfs/filesystem"
	"github.com/rstms/go-diskfs/filesystem/iso9660"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	EFI_IMAGE_SIZE         = 1024 * 1440
	ISO_PAD_BYTES          = 1024
	ISO_LOGICAL_BLOCK_SIZE = 2048
)

func CreateEFIImage(imageFilename, efiFilename, efiName string, extraFiles []string) error {
	fmt.Printf("CreateFdImage(%s, %s, %s, %v)\n", imageFilename, efiFilename, efiName, extraFiles)

	disk, err := diskfs.Create(imageFilename, EFI_IMAGE_SIZE, diskfs.Raw)
	if err != nil {
		return err
	}
	log.Printf("disk: %+v\n", disk)
	spec := diskpkg.FilesystemSpec{FSType: filesystem.TypeFat32}

	dfs, err := disk.CreateFilesystem(spec)
	if err != nil {
		return err
	}
	log.Printf("dfs: %+v\n", dfs)
	err = dfs.Mkdir("/EFI/BOOT")
	if err != nil {
		return err
	}
	log.Println("mkdir success")

	err = copyFileToImage(dfs, "/EFI/BOOT/"+efiName, efiFilename)
	if err != nil {
		return err
	}

	for _, extraFile := range extraFiles {
		_, name := filepath.Split(extraFile)
		err = copyFileToImage(dfs, "/"+name, extraFile)
		if err != nil {
			return err
		}
	}
	return nil
}

func copyFileToImage(imageFS filesystem.FileSystem, dstPath string, srcPath string) error {
	log.Printf("copyFileToImage(%s %s)\n", dstPath, srcPath)
	ifp, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer ifp.Close()
	ofp, err := imageFS.OpenFile(dstPath, os.O_CREATE|os.O_RDWR)
	if err != nil {
		return err
	}
	defer ofp.Close()
	_, err = io.Copy(ofp, ifp)
	if err != nil {
		return err
	}
	return nil
}

func copyFileFromImage(imageFS filesystem.FileSystem, dstPath string, srcPath string) error {
	//srcPath = strings.TrimLeft(srcPath, "/")
	log.Printf("copyFileFromImage(%s %s)\n", dstPath, srcPath)
	ifp, err := imageFS.OpenFile(srcPath, os.O_RDONLY)
	if err != nil {
		return err
	}
	defer ifp.Close()
	log.Printf("opened src: %v\n", ifp)
	ofp, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer ofp.Close()
	log.Printf("opened dst: %v\n", ifp)
	_, err = io.Copy(ofp, ifp)
	if err != nil {
		return err
	}
	return nil
}

func copyFileInterImage(dstFS filesystem.FileSystem, dstPath string, srcFS filesystem.FileSystem, srcPath string) error {
	log.Printf("copyFileInterImage(%s %s)\n", dstPath, srcPath)
	ifp, err := srcFS.OpenFile(srcPath, os.O_RDONLY)
	if err != nil {
		return err
	}
	defer ifp.Close()
	ofp, err := dstFS.OpenFile(dstPath, os.O_CREATE|os.O_RDWR)
	if err != nil {
		return err
	}
	defer ofp.Close()
	_, err = io.Copy(ofp, ifp)
	if err != nil {
		return err
	}
	return nil
}

func openImageFS(imageFilename string) (filesystem.FileSystem, error) {
	log.Printf("openImageFS(%s)\n", imageFilename)
	disk, err := diskfs.Open(imageFilename)
	if err != nil {
		return nil, err
	}
	log.Printf("opened disk: %+v\n", disk)

	fs, err := disk.GetFilesystem(0)
	if err != nil {
		return nil, err
	}
	log.Printf("opened filesystem: %+v\n", fs)

	return fs, nil
}

func walkFS(fs filesystem.FileSystem, dir string) ([]string, error) {
	entries, err := fs.ReadDir(dir)
	if err != nil {
		return []string{}, err
	}
	//fmt.Printf("walkFS: %s %v\n", dir, entries)
	files := []string{}
	for _, entry := range entries {
		//fmt.Printf("entry Name=%s isDir=%v %+v\n", entry.Name(), entry.IsDir(), entry)
		if entry.Name() == "NO NAME" {
			continue
		}
		name := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			if entry.Name() != "." && entry.Name() != ".." {
				files = append(files, name+"/")
				dirFiles, err := walkFS(fs, filepath.Join(dir, entry.Name()))
				if err != nil {
					return []string{}, err
				}
				for _, dirFile := range dirFiles {
					if dirFile != name {
						files = append(files, dirFile)
					}
				}
			}
		} else {
			files = append(files, name)
		}
	}
	return files, nil
}

func ListImageFiles(imageFilename string) ([]string, error) {

	fs, err := openImageFS(imageFilename)
	if err != nil {
		return []string{}, err
	}
	files, err := walkFS(fs, "/")
	if err != nil {
		return []string{}, err
	}
	return files, nil
}

func ExtractImageFiles(imageFilename string, destDir string) error {
	files, err := ListImageFiles(imageFilename)
	if err != nil {
		return err
	}
	fs, err := openImageFS(imageFilename)
	if err != nil {
		return err
	}
	for _, file := range files {
		if strings.HasSuffix(file, "/") {
			dir := strings.TrimRight(file, "/")
			err := os.Mkdir(filepath.Join(destDir, dir), 0700)
			if err != nil {
				return err
			}
		} else {
			err := copyFileFromImage(fs, filepath.Join(destDir, file), file)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func ImageInfo(imageFile string) (string, int64, error) {
	stat, err := os.Stat(imageFile)
	if err != nil {
		return "", 0, err
	}
	size := stat.Size()
	fs, err := openImageFS(imageFile)
	if err != nil {
		return "", 0, err
	}
	fmt.Printf("%+v\n", fs)
	name := strings.TrimSpace(fs.Label())
	return name, size, nil
}

func CreateISOImage(dstImage, srcImage, autoexec string) error {

	stat, err := os.Stat(autoexec)
	if err != nil {
		return err
	}
	autoexecSize := stat.Size()
	imageName, imageSize, err := ImageInfo(srcImage)

	log.Printf("autoexecSize: %d\n", autoexecSize)
	log.Printf("imageName: %s\n", imageName)
	log.Printf("imageSize: %d\n", imageSize)

	tmpDir, err := os.MkdirTemp("", "isobuild*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	// isoFiles is the list of files in the source ISO
	isoFiles, err := ListImageFiles(srcImage)
	if err != nil {
		return err
	}

	// efiSrcImage is the filename of the EFI boot image in the source ISO
	var efiSrcImage string
	for _, name := range isoFiles {
		log.Println(name)
		if strings.HasSuffix(name, ".img") {
			efiSrcImage = name
		}
	}

	// efiImageName is the basename of the ISO EFI boot image
	_, efiImageName := path.Split(efiSrcImage)

	// efiTmpSrcImage is the temp dir copy of the ISO EFI boot image
	efiTmpSrcImage := filepath.Join(tmpDir, efiImageName+".iso")

	// efiTmpModImage is the temp dir generated EFI boot image
	efiTmpModImage := filepath.Join(tmpDir, efiImageName+".mod")

	// open the source ISO filesystem
	srcFS, err := openImageFS(srcImage)
	if err != nil {
		return err
	}

	// copy the EFI boot image from the source ISO to efiTmpSrcImage
	err = copyFileFromImage(srcFS, efiTmpSrcImage, efiSrcImage)
	if err != nil {
		return err
	}

	// get the list of files in the EFI boot image
	efiFiles, err := ListImageFiles(efiTmpSrcImage)
	if err != nil {
		return err
	}

	// efiBootBin is the EFI boot binary in the EFI boot image
	var efiBootBin string
	for _, name := range efiFiles {
		log.Printf("EFI file: %s\n", name)
		if !strings.HasSuffix(name, "/") && strings.HasPrefix(name, "/EFI/BOOT") {
			efiBootBin = name
			log.Printf("efiBootBin: %s\n", name)
		}
	}

	_, efiBootName := path.Split(efiBootBin)
	log.Printf("efiBootName: %s\n", efiBootName)

	// efiTmpBootBin is the boot binary extracted from the EFI boot image
	efiTmpBootBin := filepath.Join(tmpDir, efiBootBin)
	// copy the EFI boot binary from the EFI boot image to efiTmpBootBin
	efiFS, err := openImageFS(efiTmpSrcImage)
	if err != nil {
		return err
	}

	// copy the EFI boot binary from the extracted EFI boot image
	err = copyFileFromImage(efiFS, efiTmpBootBin, efiBootBin)
	if err != nil {
		return err
	}

	err = CreateEFIImage(efiTmpModImage, efiTmpBootBin, efiBootName, []string{autoexec})
	if err != nil {
		return err
	}

	outputIsoSize := imageSize + autoexecSize*2 + ISO_PAD_BYTES

	isoDisk, err := diskfs.Create(dstImage, outputIsoSize, diskfs.Raw)
	if err != nil {
		return err
	}

	log.Printf("created ISO disk: %+v\n", isoDisk)

	isoDisk.LogicalBlocksize = ISO_LOGICAL_BLOCK_SIZE
	spec := diskpkg.FilesystemSpec{
		FSType:      filesystem.TypeISO9660,
		VolumeLabel: imageName,
	}
	dstFS, err := isoDisk.CreateFilesystem(spec)
	if err != nil {
		return err
	}

	log.Printf("created ISO filesystem: %+v\n", dstFS)

	// copy src ISO files to dest ISO
	for _, file := range isoFiles {
		switch file {
		case "autoexec.ipxe":
			// copy the modified autoexec
			log.Println("writing modified autoexec.ipxe")
			err = copyFileToImage(dstFS, "autoexec.ipxe", autoexec)
			if err != nil {
				return err
			}
		case efiSrcImage:
			// defer until finalize
		case "isolinux.bin":
			// copy to tmp for use by finalize
			err = copyFileFromImage(srcFS, filepath.Join(tmpDir, file), file)
			if err != nil {
				return err
			}
		case "boot.catalog":
			// don't copy (autogenerated)
		default:
			log.Printf("copying: %s\n", file)
			err = copyFileInterImage(dstFS, file, srcFS, file)
			if err != nil {
				return err
			}
		}
	}

	options := iso9660.FinalizeOptions{
		VolumeIdentifier: imageName,
		RockRidge:        true,
		ElTorito: &iso9660.ElTorito{
			Entries: []*iso9660.ElToritoEntry{
				{
					Platform:  iso9660.BIOS,
					Emulation: iso9660.NoEmulation,
					BootFile:  filepath.Join(tmpDir, "isolinux.bin"),
					BootTable: true,
					LoadSize:  4,
				},
				{
					Platform:  iso9660.EFI,
					Emulation: iso9660.NoEmulation,
					BootFile:  efiTmpModImage,
				},
			},
		},
	}
	iso, ok := dstFS.(*iso9660.FileSystem)
	if !ok {
		return fmt.Errorf("filesystem is not iso9660")
	}
	log.Printf("finalizing: %+v\n", options)
	err = iso.Finalize(options)
	if err != nil {
		return err
	}
	log.Println("finalized")
	return nil
}
