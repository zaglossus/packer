// +build windows
package common

import (
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
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
	state.Put("error", fmt.Errorf("CDROM support not implemented for windows: %s", err))
	return multistep.ActionHalt
}

func (s *StepCreateCD) Cleanup(multistep.StateBag) {}
