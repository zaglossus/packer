package common

import (
	"context"
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

// StepCreateCD will create a CD disk with the given files.
type StepCreateCD struct {
	// Files can be either files or directories. Any files provided here will
	// be written to the root of the CD. Directories will be written to the
	// root of the CD as well, but will retain their subdirectory structure.
	Files []string
	Label string

	CDPath string

	FilesAdded map[string]bool
}

func (s *StepCreateCD) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if len(s.Files) == 0 {
		log.Println("No CD files specified. CD disk will not be made.")
		return multistep.ActionContinue
	}

	if s.Label == "" {
		s.Label = "packer"
	} else {
		log.Printf("CD label is set to %s", s.Label)
	}

	s.FilesAdded = make(map[string]bool)

	ui := state.Get("ui").(packer.Ui)
	ui.Say("Creating CD disk...")

	// Create a temporary file to be our CD drive
	CDF, err := tmp.File("packer*.iso")
	// Set the path so we can remove it later
	s.CDPath = CDF.Name()
	CDF.Close()
	os.Remove(s.CDPath)
	if err != nil {
		state.Put("error",
			fmt.Errorf("Error creating temporary file for CD: %s", err))
		return multistep.ActionHalt
	}

	log.Printf("CD path: %s", s.CDPath)

	var diskSize int64
	diskSize = 10 * 1024 * 1024 * 1024 // 10 MB hardcoded for now.
	mydisk, err := diskfs.Create(s.CDPath, diskSize, diskfs.Raw)
	if err != nil {
		state.Put("error", fmt.Errorf("Error creating CD image: %s", err))
		return multistep.ActionHalt
	}
	defer CDF.Close()

	// ISOs may have logical block sizes only of 2048, 4096, or 8192.
	mydisk.LogicalBlocksize = 2048

	// Disk has been created; now create filesystem.
	fspec := disk.FilesystemSpec{
		Partition:   0,
		FSType:      filesystem.TypeISO9660,
		VolumeLabel: s.Label,
	}

	fs, err := mydisk.CreateFilesystem(fspec)
	if err != nil {
		state.Put("error", fmt.Errorf("Error creating filesystem on CD image: %s", err))
		return multistep.ActionHalt
	}

	for _, file := range s.Files {
		log.Printf("Adding %s to CDrom", file)
		err := s.Add(fs, file)
		if err != nil {
			state.Put("error", fmt.Errorf("Error adding file to CD: %s", err))
			return multistep.ActionHalt
		}
	}

	iso, ok := fs.(*iso9660.FileSystem)
	if !ok {
		if err != nil {
			state.Put("error", fmt.Errorf("Created CD is not an iso filesystem: %s", err))
			return multistep.ActionHalt
		}
	}

	err = iso.Finalize(iso9660.FinalizeOptions{})
	if err != nil {
		state.Put("error", fmt.Errorf("Error finalizing iso system: %s", err))
		return multistep.ActionHalt
	}
	ui.Message("Done copying paths from CD_dirs")

	// Set the path to the CD so it can be used later
	state.Put("cd_path", s.CDPath)

	return multistep.ActionContinue
}

func (s *StepCreateCD) Add(fs filesystem.FileSystem, src string) error {
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

func (s *StepCreateCD) Cleanup(multistep.StateBag) {
	if s.CDPath != "" {
		log.Printf("Deleting CD disk: %s", s.CDPath)
		os.Remove(s.CDPath)
	}
}
