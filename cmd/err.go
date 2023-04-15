// Copyright © 2019 Weald Technology Trading
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

	"github.com/spf13/viper"
)

// errCheck checks for an error and quits if it is present.
func errCheck(err error, msg string) {
	if err != nil {
		if !viper.GetBool("quiet") {
			if msg == "" {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			} else {
				fmt.Fprintf(os.Stderr, "%s: %s\n", msg, err.Error())
			}
		}
		os.Exit(1)
	}
}

// errAssert checks a condition and quits if it is false
// func errAssert(condition bool, err error, msg string) {
// 	if !condition {
// 		if err != nil {
// 			if !quiet {
// 				if msg == "" {
// 					fmt.Fprintf(os.Stderr, "%s\n", err.Error())
// 				} else {
// 					fmt.Fprintf(os.Stderr, "%s: %s\n", msg, err.Error())
// 				}
// 			}
// 			os.Exit(1)
// 		}
// 	}
// }

// assert checks a condition and quits if it is false.
func assert(condition bool, msg string) {
	if !condition {
		die(msg)
	}
}

// die prints an error and quits.
func die(msg string) {
	if msg != "" && !viper.GetBool("quiet") {
		fmt.Fprintf(os.Stderr, "%s\n", msg)
	}
	os.Exit(_exitFailure)
}

// warnCheck checks for an error and warns if it is present
// func warnCheck(err error, msg string) {
// 	if err != nil {
// 		if !quiet {
// 			if msg == "" {
// 				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
// 			} else {
// 				fmt.Fprintf(os.Stderr, "%s: %s\n", msg, err.Error())
// 			}
// 		}
// 	}
// }

// check checks a condition and warns if it is false
// func check(condition bool, msg string) {
// 	if !condition {
// 		warn(msg)
// 	}
// }

// Warn prints a warning
// func warn(msg string) {
// 	if !quiet {
// 		fmt.Fprintf(os.Stderr, "%s\n", msg)
// 	}
// }
