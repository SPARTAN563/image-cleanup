# image-cleanup
**Organize your image folders**

image-cleanup is a tool used to quickly and easily organize your image directories by renaming
images based on their EXIF metadata. By default it attempts to de-duplicate names of images in
a consistent, predictable, fashion which allows you to quickly identify duplicates while
avoiding difficult to order `DSC_0110 (1).png` filenames.

## Usage
```
image-cleanup [command]

Available Commands:
  help        Help about any command
  remove      Removes images which are present in another directory tree.
  rename      Renames images in the tree based on a template which may use EXIF tag fields.

Flags:
      --config string   config file (default is $HOME/.image-cleanup.yaml)
  -d, --dry-run         Perform a dry-run of the application without mutating files
  -h, --help            help for image-cleanup

Use "image-cleanup [command] --help" for more information about a command.
```

### `image-cleanup rename`
Scans a directory structure, extracting EXIF data for each image and renaming them according to a provided template function.

```
image-cleanup rename [flags]

Flags:
  -h, --help              help for rename
      --target string     The directory from which to remove files
      --template string   The template used to generate the new filename (default "{{ .FileName }}{{ .Extension }}")

Global Flags:
      --config string   config file (default is $HOME/.image-cleanup.yaml)
  -d, --dry-run         Perform a dry-run of the application without mutating files
```

#### Examples

This is my personal template which results in timestamped images with the device that took them listed (if present in the EXIF data) and the
original filename (commonly `DSC_0001`).

```
image-cleanup rename --template="{{ .DateTime }}{{ if .Model }}_{{ .Model }}{{ end }}_{{ .FileNameClean }}{{ .Extension }}" --target ./
```


### `image-cleanup remove`
Scans a candidate directory structure to identify which images should be removed from a target directory structure.

```
image-cleanup remove [flags]

Flags:
      --candidate stringArray   The directory holding the images to be removed from the target
  -h, --help                    help for remove
      --target string           The directory from which to remove files

Global Flags:
      --config string   config file (default is $HOME/.image-cleanup.yaml)
  -d, --dry-run         Perform a dry-run of the application without mutating files
```