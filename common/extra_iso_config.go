//go:generate struct-markdown

package common

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/packer/template/interpolate"
)

// An iso (CD) containing custom files can be made available for your build.
//
// By default, no extra CD will be attached. All files listed in this setting
// get placed into the root directory of the CD and the CD is attached as the
// second CD device.
//
// This config exists to work around modern operating systems that have no
// way to mount floppy disks, which was our previous go-to for adding files at
// boot time.
//
// It may not work for all installers. For example, in debian the preseed file
// needs to be remastered inside the installation iso, which is outside the
// scope of this tool
type CDConfig struct {
	// A list of files to place onto a CD that is attached when the VM is
	// booted. This can include either files or directories; any directories
	// will be copied onto the CD recursively, preserving directory structure
	// hierarchy. Symlinks will be ignored.
	CDFiles []string `mapstructure:"cd_files"`
	CDLabel string   `mapstructure:"cd_label"`
}

func (c *CDConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error
	var err error

	if c.CDFiles == nil {
		c.CDFiles = make([]string, 0)
	}

	for _, path := range c.CDFiles {
		if strings.ContainsAny(path, "*?[") {
			_, err = filepath.Glob(path)
		} else {
			_, err = os.Stat(path)
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("Bad CD disk file '%s': %s", path, err))
		}
	}

	return errs
}
