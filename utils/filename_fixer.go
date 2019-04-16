package utils

import "regexp"

type fixer struct {
	re   *regexp.Regexp
	repl string
}

type FilenameFixer struct {
	fixers []*fixer
}

func NewFilenameFixer() *FilenameFixer {
	return &FilenameFixer{
		fixers: []*fixer{
			&fixer{
				re:   regexp.MustCompile(` \(\d+\)$`),
				repl: "",
			},
			&fixer{
				re:   regexp.MustCompile(`(DSC_\d+)_\d+$`),
				repl: "$1",
			},
		},
	}
}

func (f *FilenameFixer) Fix(filename string) string {
	for _, fixer := range f.fixers {
		filename = fixer.re.ReplaceAllString(filename, fixer.repl)
	}

	return filename
}
