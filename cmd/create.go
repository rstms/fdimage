/*
Copyright © 2025 Matt Krueger <mkrueger@rstms.net>
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

 1. Redistributions of source code must retain the above copyright notice,
    this list of conditions and the following disclaimer.

 2. Redistributions in binary form must reproduce the above copyright notice,
    this list of conditions and the following disclaimer in the documentation
    and/or other materials provided with the distribution.

 3. Neither the name of the copyright holder nor the names of its contributors
    may be used to endorse or promote products derived from this software
    without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.
*/
package cmd

import (
	"fmt"
	"github.com/rstms/fdimage/image"
	"os"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create IMAGE_FILE EFI_FILE EFI_NAME [EXTRA_FILE ...]",
	Short: "create EFI floppy image for a bootable ISO",
	Long: `
Create a FAT formatted floppy disk image file in IMAGE_FILE.  Copy EFI_FILE
into the boot image as /EFI/BOOT/{EFI_NAME}
Copy files named by EXTRA_FILE arguments into the image root directory.
`,
	Args: cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		imageFile := args[0]
		efiFile := args[1]
		efiName := args[2]
		force := ViperGetBool("create.force")
		if force {
			err := os.Remove(imageFile)
			cobra.CheckErr(err)
		}
		if IsFile(imageFile) {
			cobra.CheckErr(fmt.Errorf("file exists: %s", imageFile))
		}
		count := len(args) - 3
		extraFiles := make([]string, count)
		for i := 0; i < count; i++ {
			extraFiles[i] = args[3+i]
		}
		err := image.CreateEFIImage(imageFile, efiFile, efiName, extraFiles)
		cobra.CheckErr(err)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	OptionSwitch(createCmd, "force", "f", "bypass confirmation prompt")
}
