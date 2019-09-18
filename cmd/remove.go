// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
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
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Removes images which are present in another directory tree.",
	Long:  `Scans a candidate directory structure to identify which images should be removed from a target directory structure.`,
	Run: func(cmd *cobra.Command, args []string) {
		log := logrus.WithFields(logrus.Fields{})

		apply, err := cmd.Root().PersistentFlags().GetBool("apply")
		if err != nil {
			log.WithError(err).Error()
			return
		}

		target, err := cmd.PersistentFlags().GetString("target")
		if err != nil {
			log.WithError(err).Error()
			return
		}

		candidates, err := cmd.PersistentFlags().GetStringArray("candidate")
		if err != nil {
			log.WithError(err).Error()
			return
		}

		log = log.WithFields(logrus.Fields{
			"apply":      apply,
			"target":     target,
			"candidates": candidates,
		})

		removals := map[string]struct{}{}

		for _, candidate := range candidates {
			log := log.WithField("candidate", candidate)

			if err := filepath.Walk(candidate, func(path string, info os.FileInfo, err error) error {
				log := log.WithField("path", path)

				if err != nil {
					log.WithError(err).Error()
					return nil
				}

				if info.IsDir() {
					return nil
				}

				base := filepath.Base(info.Name())
				if base == "." {
					return nil
				}

				log.WithField("filename", base).Debug("Added to removal list")
				removals[base] = struct{}{}

				return nil
			}); err != nil {
				log.WithError(err).Error("Failed to enumerate candidate directory")
			}
		}

		if target == "" {
			for filename := range removals {
				log.WithField("filename", filename).Info("Found candidate filename")
			}
		} else {
			if err := filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
				log := log.WithField("path", path)
				if err != nil {
					log.WithError(err).Warning()
					return nil
				}

				if info.IsDir() {
					return nil
				}

				base := filepath.Base(info.Name())
				if base == "." {
					return nil
				}

				log = log.WithField("filename", base)

				if _, ok := removals[base]; ok {
					log.Info("Removing file")
					if apply {
						if err := os.Remove(path); err != nil {
							log.WithError(err).Error("Failed to remove file")
						}
					}
				}

				return nil
			}); err != nil {
				log.WithError(err).Error("Failed to remove files")
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	removeCmd.PersistentFlags().String("target", "", "The directory from which to remove files")
	removeCmd.PersistentFlags().StringArray("candidate", []string{}, "The directory holding the images to be removed from the target")
	removeCmd.MarkPersistentFlagRequired("candidate")
	removeCmd.MarkPersistentFlagFilename("target")
	removeCmd.MarkPersistentFlagFilename("candidate")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// removeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
