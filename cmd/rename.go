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
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
	"github.com/sirupsen/logrus"
	"github.com/spartan563/image-cleanup/utils"
	"github.com/spf13/cobra"
)

type imageMetaFields map[string]string

func (f imageMetaFields) Walk(name exif.FieldName, tag *tiff.Tag) error {
	if tag.String() == "<no value>" {
		return nil
	}

	f[string(name)] = strings.Trim(tag.String(), "\"")
	return nil
}

// renameCmd represents the rename command
var renameCmd = &cobra.Command{
	Use:   "rename",
	Short: "Renames images in the tree based on a template which may use EXIF tag fields.",
	Long:  `Scans a directory structure, extracting EXIF data for each image and renaming them according to a provided template function.`,
	Run: func(cmd *cobra.Command, args []string) {
		apply, err := cmd.Root().PersistentFlags().GetBool("apply")
		if err != nil {
			fmt.Println(err)
			return
		}

		target, err := cmd.PersistentFlags().GetString("target")
		if err != nil {
			fmt.Println(err)
			return
		}

		filenameTemplate, err := cmd.PersistentFlags().GetString("template")
		if err != nil {
			fmt.Println(err)
			return
		}

		tmpl, err := template.New("filename").Parse(filenameTemplate)
		if err != nil {
			fmt.Println(err)
			return
		}

		filenameFixer := utils.NewFilenameFixer()

		log := logrus.WithFields(logrus.Fields{
			"target":   target,
			"apply":    apply,
			"template": filenameTemplate,
		})

		if err := filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
			log := log.WithField("path", path)

			if err != nil {
				log.WithError(err).Error()
				return nil
			}

			if info.IsDir() {
				log.Debug("Skipping directory")
				return nil
			}

			log.Debug("Opening file")
			f, err := os.Open(path)
			if err != nil {
				log.WithError(err).Error()
				return nil
			}

			log.Debug("Decoding EXIF data")
			fields, err := exif.Decode(f)
			if err != nil {
				log.WithError(err).Warning("Failed to decode EXIF data")
				f.Close()
				return nil
			}

			f.Close()

			context := imageMetaFields{}
			context["Extension"] = strings.ToLower(filepath.Ext(info.Name()))
			context["FileName"] = info.Name()[:len(info.Name())-len(context["Extension"])]
			context["FileNameClean"] = filenameFixer.Fix(context["FileName"])

			fields.Walk(context)

			date, err := fields.DateTime()
			if err == nil {
				context["Date"] = date.Format("2006-01-02")
				context["Time"] = date.Format("15-04-05")
				context["DateTime"] = date.Format("2006-01-02T15-04-05")
			} else {
				context["DateTime"] = strings.Replace(context["DateTime"], ":", "-", -1)
			}

			filename := bytes.NewBufferString("")
			filename.WriteString(filepath.Dir(path))
			filename.WriteRune(filepath.Separator)

			if err := tmpl.Execute(filename, context); err != nil {
				log.WithError(err).Warning("Failed to render template")
				return nil
			}

			log = log.WithField("newPath", filename.String())

			if path != filename.String() {
				log.Info("Renaming file")
				// Move the file
				if apply {
					if err := os.Rename(path, filename.String()); err != nil {
						log.WithError(err).Error("Failed to rename file")
					}
				}
			} else {
				log.Debug("Skipping file, no change necessary")
			}

			return nil
		}); err != nil {
			log.WithError(err).Error("Failed to rename files")
		}
	},
}

func init() {
	rootCmd.AddCommand(renameCmd)
	renameCmd.PersistentFlags().String("target", "", "The directory from which to remove files")
	renameCmd.MarkPersistentFlagFilename("target")

	renameCmd.PersistentFlags().String("template", "{{ .FileName }}{{ .Extension }}", "The template used to generate the new filename")
	renameCmd.MarkPersistentFlagRequired("template")
}
