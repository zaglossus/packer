// +build !windows

package common

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer/tmp"

	diskfs "github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/disk"
	"github.com/diskfs/go-diskfs/filesystem"
	"github.com/diskfs/go-diskfs/filesystem/iso9660"
)

func RunCreateCD(files []string, label string, state multistep.StateBag) (string, error) {
	if len(files) == 0 {
		log.Println("No CD files specified. CD disk will not be made.")
		return "", nil
	}

	if label == "" {
		label = "packer"
	} else {
		log.Printf("CD label is set to %s", label)
	}

	ui := state.Get("ui").(packer.Ui)
	ui.Say("Creating CD disk...")

	// Create a temporary file to be our CD drive
	CDF, err := tmp.File("packer*.iso")
	// Set the path so we can remove it later
	CDPath := CDF.Name()
	CDF.Close()
	os.Remove(CDPath)
	if err != nil {
		return "", fmt.Errorf("Error creating temporary file for CD: %s", err)
	}

	log.Printf("CD path: %s", CDPath)

	diskSize := int64(10 * 1024 * 1024 * 1024) // 10 MB hardcoded for now.
	mydisk, err := diskfs.Create(CDPath, diskSize, diskfs.Raw)
	if err != nil {
		return "", fmt.Errorf("Error creating CD image: %s", err)
	}
	defer CDF.Close()

	// ISOs may have logical block sizes only of 2048, 4096, or 8192.
	mydisk.LogicalBlocksize = 2048

	// Disk has been created; now create filesystem.
	fspec := disk.FilesystemSpec{
		Partition:   0,
		FSType:      filesystem.TypeISO9660,
		VolumeLabel: label,
	}

	fs, err := mydisk.CreateFilesystem(fspec)
	if err != nil {
		return "", fmt.Errorf("Error creating filesystem on CD image: %s", err)
	}

	for _, file := range files {
		log.Printf("Adding %s to CDrom", file)
		err := AddFile(fs, file)
		if err != nil {
			return "", fmt.Errorf("Error adding file to CD: %s", err)
		}
	}

	iso, ok := fs.(*iso9660.FileSystem)
	if !ok {
		if err != nil {
			return "", fmt.Errorf("Created CD is not an iso filesystem: %s", err)
		}
	}

	err = iso.Finalize(iso9660.FinalizeOptions{})
	if err != nil {
		return "", fmt.Errorf("Error finalizing iso system: %s", err)
	}
	ui.Message("Done copying paths from CD_dirs")

	// Set the path to the CD so it can be used later
	state.Put("cd_path", CDPath)

	return CDPath, nil
}

func AddFile(fs filesystem.FileSystem, src string) error {
	finfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("Error adding path to CD: %s", err)
	}

	// add a file
	if !finfo.IsDir() {
		inputF, err := os.Open(src)
		if err != nil {
			return err
		}
		defer inputF.Close()

		rw, err := fs.OpenFile(finfo.Name(), os.O_CREATE|os.O_RDWR)
		if err != nil {
			return fmt.Errorf("Error opening file for copy %s to CD", src)
		}
		nBytes, err := io.Copy(rw, inputF)
		if err != nil {
			return fmt.Errorf("Error copying %s to CD", src)
		}
		log.Printf("Wrote %d bytes to %s", nBytes, finfo.Name())
		return err
	}

	// Add a directory and its subdirectories
	visit := func(pathname string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// add a file
		if !fi.IsDir() {
			inputF, err := os.Open(src)
			if err != nil {
				return err
			}
			defer inputF.Close()

			rw, err := fs.OpenFile(fi.Name(), os.O_CREATE|os.O_RDWR)
			if err != nil {
				return fmt.Errorf("Error opening file %s on CD", src)
			}
			nBytes, err := io.Copy(inputF, rw)
			if err != nil {
				return fmt.Errorf("Error copying %s to CD", src)
			}
			log.Printf("Wrote %d bytes to %s", nBytes, finfo.Name())
			return err
		}

		if fi.Mode().IsDir() {
			// create the directory on the CD, continue walk.
			err := fs.Mkdir(pathname)
			if err != nil {
				err = fmt.Errorf("error creating new directory %s: %s", pathname, err)
			}
			return err
		}
		return err
	}

	return filepath.Walk(src, visit)
}
