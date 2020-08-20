package common

import (
	"context"
	"log"
	"os"

	"github.com/hashicorp/packer/helper/multistep"
)

// StepCreateCD will create a CD disk with the given files.
type StepCreateCD struct {
	// Files can be either files or directories. Any files provided here will
	// be written to the root of the CD. Directories will be written to the
	// root of the CD as well, but will retain their subdirectory structure.
	Files []string
	Label string

	CDPath string
}

func (s *StepCreateCD) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	cd_path, err := RunCreateCD(s.Files, s.Label, state)
	s.CDPath = cd_path
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepCreateCD) Cleanup(multistep.StateBag) {
	if s.CDPath != "" {
		log.Printf("Deleting CD disk: %s", s.CDPath)
		os.Remove(s.CDPath)
	}
}
