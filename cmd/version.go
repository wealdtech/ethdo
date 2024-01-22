// Copyright © 2019 - 2024 Weald Technology Trading.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ReleaseVersion is the release version of the codebase.
// Usually overridden by tag names when building binaries.
var ReleaseVersion = "local build (latest release 1.35.2)"

// versionCmd represents the version command.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version of ethdo",
	Long: `Obtain the version of ethdo.  For example:

    ethdo version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(ReleaseVersion)
		if viper.GetBool("verbose") {
			if info, ok := debug.ReadBuildInfo(); ok {
				for _, setting := range info.Settings {
					if setting.Key == "vcs.revision" {
						fmt.Printf("Commit hash: %s\n", setting.Value)
						break
					}
				}
			}

			buildInfo, ok := debug.ReadBuildInfo()
			if ok {
				fmt.Printf("Package: %s\n", buildInfo.Path)
				fmt.Println("Dependencies:")
				for _, dep := range buildInfo.Deps {
					for dep.Replace != nil {
						dep = dep.Replace
					}
					fmt.Printf("\t%v %v\n", dep.Path, dep.Version)
				}
			}
		}
		os.Exit(_exitSuccess)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
