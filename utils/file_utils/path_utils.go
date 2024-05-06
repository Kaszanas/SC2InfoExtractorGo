package file_utils

import (
	"io/fs"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// ListFiles creates a slice of filepaths from a give input directory
// based filtering supplied fileExtension
func ListFiles(
	inputPath string,
	filterFileExtension string,
) ([]string, error) {

	log.WithFields(log.Fields{
		"inputPath":           inputPath,
		"filterFileExtension": filterFileExtension}).
		Info("Entered ListFiles()")

	// files, err := os.ReadDir(inputPath)
	// if err != nil {
	// 	log.WithField("error", err).Error("Error reading directory")
	// 	return nil, err
	// }

	var listOfFiles []string
	if filterFileExtension == "" {
		listOfFiles, err := getAllFiles(inputPath)
		if err != nil {
			log.WithField("error", err).Error("Error getting list of files")
			return nil, err
		}
		return listOfFiles, nil
	}

	listOfFiles, err := getFilesByExtension(inputPath, filterFileExtension)
	if err != nil {
		log.WithField("error", err).Error("Error getting list of files")
		return nil, err
	}

	log.WithField("n_files", len(listOfFiles)).Info("Finished ListFiles()")
	return listOfFiles, nil
}

// ExistingFilesSet creates a set of existing files in a directory.
func ExistingFilesSet(
	inputPath string,
	fiterFileExtension string,
) (map[string]struct{}, error) {

	log.WithFields(log.Fields{
		"inputPath":           inputPath,
		"filterFileExtension": fiterFileExtension}).
		Info("Entered ExistingFilesSet()")

	// List the files in the selected directory:
	listOfFiles, err := ListFiles(inputPath, fiterFileExtension)
	if err != nil {
		log.WithField("error", err).Error("Error getting list of files")
		return nil, err
	}

	// Convert from a slice (list) to a set (map),
	// by default files should not be duplicates:
	existingFilesSet := make(map[string]struct{})
	for _, file := range listOfFiles {
		existingFilesSet[file] = struct{}{}
	}

	return existingFilesSet, nil
}

func getFilesByExtension(
	inputPath string,
	filterFileExtension string,
) ([]string, error) {
	var listOfFiles []string

	err := filepath.WalkDir(
		inputPath,
		func(path string, dirEntry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !dirEntry.IsDir() &&
				filepath.Ext(dirEntry.Name()) == filterFileExtension {
				listOfFiles = append(listOfFiles, path)
			}
			return nil
		})

	if err != nil {
		return nil, err
	}

	return listOfFiles, nil
}

func getAllFiles(inputPath string) ([]string, error) {
	var listOfFiles []string

	err := filepath.WalkDir(
		inputPath,
		func(path string, dirEntry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !dirEntry.IsDir() {
				listOfFiles = append(listOfFiles, path)
			}
			return nil
		})

	if err != nil {
		return nil, err
	}

	return listOfFiles, nil
}
