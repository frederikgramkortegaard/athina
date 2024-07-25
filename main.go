package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/sergi/go-diff/diffmatchpatch"
)

var config AthinaConfig

func isDeltaDiffEmpty(delta string) bool {

	for _, c := range delta {
		if c == '\t' || c == '+' {
			return false
		}
	}

	return true

}

func AthinaUpdateAllFiles(log bool) error {

	for change := range AthinaLookForFileChanges() {
		switch change.action {
		case AthinaFileChangeActionAdd:
			_ = AthinaAddFile(change.filename)
			if log {
				fmt.Println("New file: " + change.filename)
			}

		case AthinaFileChangeActionDelete:
			_ = AthinaDeleteFile(change.filename)
			if log {
				fmt.Println("Deleted file: " + change.filename)
			}

		case AthinaFileChangeActionModify:
			_ = AthinaUpdateFile(change.filename)
			if log {
				fmt.Println("Modified file: " + change.filename)
			}

		case AthinaFileChangeActionError:
			fmt.Println("Error: " + change.err.Error())
			return change.err
		case AthinaFileChangeActionNone:
			if log {
				fmt.Println("No changes detected")
			}

		}
	}
	return nil
}

func AthinaUpdateFile(filename string) error {

	// If the file does not exist but there exists a AthinaFile object for the file, we want to mark that the file has been deleted
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		if _, err := os.Stat(ATHINA_PATH_TO_OBJECTS + filename); !os.IsNotExist(err) {
			err := AthinaDeleteFile(filename)
			if err != nil {
				fmt.Println(err)
				return err
			}
		}

		return nil
	}

	// If there is no AthinaFile object for this file, that means that this is a new file and we want to create one
	if _, err := os.Stat(ATHINA_PATH_TO_OBJECTS + filename); os.IsNotExist(err) {
		err := AthinaAddFile(filename)
		if err != nil {
			fmt.Println(err)
			return err
		}

		return nil
	}

	// Otherwise, if there exists a AthinaFile object for this file, and the file exists in the current directory, we want to compare the two
	athinafile, err := loadAthinaFileObject(filename)
	if err != nil {
		fmt.Println(err)
		return err
	}

	diffs, err := diffAthinaFileObjectAndFile(athinafile, filename)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// If there are changes, we want to update the AthinaFile object with the diffs
	if !isDeltaDiffEmpty(diffs) {

		filediff := newFilediff(newFileDiffOptions{deleted: false, change: AthinaFileChangeActionModify, delta: diffs})

		AthinaModifyFile(athinafile, []Filediff{filediff})

	} else {
		fmt.Println("No changes detected in file: " + filename)
	}

	return nil

}

func AthinaAddFile(filename string) error {

	athinafile, err := createInitialAthinaFileObject(filename)
	if err != nil {
		fmt.Println(err)
		return err
	}

	athinafile.Save()
	return nil
}

func AthinaDeleteFile(filename string) error {

	// Load the Athina object
	athinafile, err := loadAthinaFileObject(filename)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Make a new Filediff object
	filediff := newFilediff(newFileDiffOptions{deleted: true, change: AthinaFileChangeActionDelete})
	athinafile.Diffs = append(athinafile.Diffs, filediff)
	athinafile.Save()

	return nil
}

func AthinaModifyFile(athinafile AthinaFile, diffs []Filediff) error {

	// Append the new diffs to the Athina object
	athinafile.Diffs = append(athinafile.Diffs, diffs...)
	athinafile.Save()
	return nil
}

func AthinaDetectFileChange(filename string) (AthinaFileChange, error) {

	// If the file does not exist in the .athina/objects folder
	if _, err := os.Stat(ATHINA_PATH_TO_OBJECTS + filename); os.IsNotExist(err) {

		// And the file does not exist in the current directory
		if _, err := os.Stat(filename); os.IsNotExist(err) {

			// Then there must be an error
			return AthinaFileChange{action: AthinaFileChangeActionError, err: err}, err
		}

		// Otherwise if it doesn't exist in the .athina/objects folder, but it does exist in the current directory, then it must be a new file
		return AthinaFileChange{action: AthinaFileChangeActionAdd, filename: filename}, nil

		// If the file does exist in the .athina/objects folder
	} else {

		// But it does not exist in the current directory
		if _, err := os.Stat(filename); os.IsNotExist(err) {

			// Then the file has been deleted
			return AthinaFileChange{action: AthinaFileChangeActionDelete, filename: filename}, nil

			// Otherwise, if it does exist in the .athina/objects folder and it does exist in the current directory, we check it for any modifications since the last update
		} else {

			// Load the Athina object
			athinafile, err := loadAthinaFileObject(filename)
			if err != nil {
				fmt.Println(err)
				return AthinaFileChange{action: AthinaFileChangeActionError, err: err}, err
			}

			// Open the file in the current directory
			diff, err := diffAthinaFileObjectAndFile(athinafile, filename)
			if err != nil {
				fmt.Println(err)
				return AthinaFileChange{action: AthinaFileChangeActionError, err: err}, err
			}

			// If there are changes, print out the changes
			if !isDeltaDiffEmpty(diff) {
				fmt.Println("Changes detected in file: " + filename)
				return AthinaFileChange{action: AthinaFileChangeActionModify, file: athinafile, filename: filename, diffs: athinafile.Diffs}, nil
			} else {
				return AthinaFileChange{action: AthinaFileChangeActionNone}, nil

			}
		}
	}
}

func AthinaLookForFileChanges() <-chan AthinaFileChange {

	ch := make(chan AthinaFileChange)

	go func() {

		// Go through each file in the .athina/objects folder, and compare it to the current file in the directory
		// If there are changes, print out the changes
		// If there are no changes, do nothing

		// Get all the files in the .athina/objects folder
		files, err := os.ReadDir(ATHINA_PATH_TO_OBJECTS)
		if err != nil {
			fmt.Println(err)
			ch <- AthinaFileChange{action: AthinaFileChangeActionError, err: err}
		}

		// For each file in the .athina/objects folder, check if there is a corresponding file in the current directory
		for _, file := range files {

			if _, err := os.Stat(file.Name()); errors.Is(err, os.ErrNotExist) {
				// File exists in the .athina/objects folder, but not in the current directory
				// This means that the file has been deleted
				ch <- AthinaFileChange{action: AthinaFileChangeActionDelete, filename: file.Name()}

			} else {

				athinafile, err := loadAthinaFileObject(file.Name())
				if err != nil {
					fmt.Println(err)
					ch <- AthinaFileChange{action: AthinaFileChangeActionError, err: err}
				}

				diff, err := diffAthinaFileObjectAndFile(athinafile, file.Name())
				if err != nil {
					fmt.Println(err)
					ch <- AthinaFileChange{action: AthinaFileChangeActionError, err: err}
				}

				if !isDeltaDiffEmpty(diff) {
					ch <- AthinaFileChange{action: AthinaFileChangeActionModify, file: athinafile, filename: file.Name(), delta: diff}

				}
			}
		}

		// Now we want to fo through all the files in the current directry that are not in the .athina/objects folder
		// If there are any, we want to add them to the .athina/objects folder
		currentfiles, err := os.ReadDir(".")
		if err != nil {
			fmt.Println(err)
			ch <- AthinaFileChange{action: AthinaFileChangeActionError, err: err}
		}

		for _, currentfile := range currentfiles {

			if config.IsIgnored(currentfile.Name()) {
				continue
			}

			if currentfile.IsDir() {
				continue
			}

			if _, err := os.Stat(ATHINA_PATH_TO_OBJECTS + currentfile.Name()); errors.Is(err, os.ErrNotExist) {
				// File exists in the current directory, but not in the .athina/objects folder, this means that the file is new and should be added
				ch <- AthinaFileChange{action: AthinaFileChangeActionAdd, filename: currentfile.Name()}

			}
		}
		close(ch)
	}()

	return ch
}

func createInitialAthinaFileObject(filename string) (AthinaFile, error) {

	// Create a new file object
	var file AthinaFile

	// Set the filename
	file.Filename = filename

	// read the files content
	lf, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return AthinaFile{}, err
	}

	defer lf.Close()

	// Read the content of file 'test.txt'
	fileinfo, err := lf.Stat()
	if err != nil {
		fmt.Println(err)
		return AthinaFile{}, err
	}

	filesize := fileinfo.Size()
	filecontent := make([]byte, filesize)
	_, err = lf.Read(filecontent)
	if err != nil {
		fmt.Println(err)
		return AthinaFile{}, err
	}

	// Convert the content of file 'test.txt' to string
	file.Origin = string(filecontent)
	filediff := newFilediff(newFileDiffOptions{added: true, change: AthinaFileChangeActionAdd})
	file.Diffs = append(file.Diffs, filediff)

	return file, nil

}

func AthinaRevertFileByHash(filename string, hash string) error {

	// Load the Athina object
	athinafile, err := loadAthinaFileObject(filename)
	if err != nil {
		fmt.Println(err)
		return err
	}

	AthinaRevertFileObjectByHash(athinafile, hash)

	return nil

}

func AthinaRevertFileObjectByHash(athinafile AthinaFile, hash string) error {

	// First we check if there exists a filediff with this hash

	var found bool = false
	for _, filediff := range athinafile.Diffs {
		if filediff.Hash == hash {
			found = true
			break

		}
	}

	if !found {
		return errors.New("no such hash found")
	}

	origin := athinafile.Origin
	dmp := diffmatchpatch.New()
	for _, filediff := range athinafile.Diffs {

		diff, _ := dmp.DiffFromDelta(origin, filediff.Delta)
		origin = dmp.DiffText2(diff)

	}

	// Add a new diff to the Athinafile object
	diff, _ := diffAthinaFileObjectAndString(athinafile, origin)
	athinafile.Origin = origin
	filediff := newFilediff(newFileDiffOptions{deleted: false, change: AthinaFileChangeActionRevert, delta: diff})

	AthinaModifyFile(athinafile, []Filediff{filediff})

	//@TODO: Modify the file in the directory

	// Write to the file
	file, err := os.Create(athinafile.Filename)
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer file.Close()

	_, err = file.WriteString(origin)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func main() {

	// Initialize the .athina folder
	initializeAthinaFolder()

	// Load the config file
	loadAthinaConfig()

	args := os.Args[1:]

	// If no arguments are passed, we default to looking for file changes
	if len(args) < 1 {

		for change := range AthinaLookForFileChanges() {
			switch change.action {
			case AthinaFileChangeActionAdd:
				fmt.Println("New file: " + change.filename)
			case AthinaFileChangeActionDelete:
				fmt.Println("Deleted file: " + change.filename)
			case AthinaFileChangeActionModify:
				fmt.Println("Modified file: " + change.filename)
			case AthinaFileChangeActionError:
				fmt.Println("Error: " + change.err.Error())
			case AthinaFileChangeActionNone:
				fmt.Println("No changes detected")
			}
		}
		return
	}

	// If arguments are passed, we send it to handleCLI
	handleCLI(args)

}
