// +build windows

package common

import (
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
)

func RunCreateCD(files []string, label string, state multistep.StateBag) (string, error) {
	return "", fmt.Errorf("CDROM support not implemented for windows")
}
