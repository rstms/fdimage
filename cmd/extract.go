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

var extractCmd = &cobra.Command{
	Use:   "extract IMAGE_FILE DEST_DIR",
	Short: "extract files from image",
	Long: `
Extract all files from a disk image to directory DEST_DIR.
`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		imageFile := args[0]
		destDir := args[1]
		force := ViperGetBool("extract.force")
		if force {
			if IsDir(destDir) {
				err := os.RemoveAll(destDir)
				cobra.CheckErr(err)
			}
			err := os.Mkdir(destDir, 0700)
			cobra.CheckErr(err)
		}
		if !IsDir(destDir) {
			cobra.CheckErr(fmt.Errorf("not a directory: %s", destDir))
		}
		err := image.ExtractImageFiles(imageFile, destDir)
		cobra.CheckErr(err)
	},
}

func init() {
	rootCmd.AddCommand(extractCmd)
	OptionSwitch(extractCmd, "force", "f", "bypass confirmation prompt")
}
