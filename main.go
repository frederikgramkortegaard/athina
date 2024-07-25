package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/sergi/go-diff/diffmatchpatch"
)

const GOGIT_FOLDER = ".athina"

type AthinaFileChangeAction string

const (
	AthinaFileChangeActionAdd    AthinaFileChangeAction = "File add"
	AthinaFileChangeActionDelete AthinaFileChangeAction = "File delete"
	AthinaFileChangeActionModify AthinaFileChangeAction = "File modify"

	AthinaFileChangeActionRevert AthinaFileChangeAction = "File revert"

	AthinaFileChangeActionNone  AthinaFileChangeAction = "none"
	AthinaFileChangeActionError AthinaFileChangeAction = "error"
)

type AthinaFile struct {
	Filename string
	Origin   string
	Diffs    []Filediff
}

type Filediff struct {
	Hash    string
	Diffs   []diffmatchpatch.Diff
	Delta   string
	Deleted bool
	Added   bool
	Change  AthinaFileChangeAction
}

type createFilediffOptions struct {
	diffs   []diffmatchpatch.Diff
	deleted bool
	added   bool
	change  AthinaFileChangeAction
	delta   string
}

func isDeltaEmpty(delta string) bool {

	for _, c := range delta {
		if c == '\t' || c == '+' {
			return false
		}
	}

	return true

}

func createFilediff(options createFilediffOptions) Filediff {

	var filediff Filediff
	filediff.Diffs = options.diffs
	filediff.Deleted = options.deleted
	filediff.Added = options.added
	filediff.Hash = sha1Hash(hashFilediff(filediff) + sha1Hash(string(options.change)))
	filediff.Change = options.change
	filediff.Delta = options.delta
	return filediff
}

type AthinaFileChange struct {
	action AthinaFileChangeAction

	file     AthinaFile
	filename string
	diffs    []Filediff
	err      error
	delta    string
}

type AthinaUpdateOptions struct {
	symbol string
}

func initializeAthinaFolder() error {

	// Ensure that there exists a .athina folder in the current directory. If not, create one
	if _, err := os.Stat(".athina"); os.IsNotExist(err) {
		err := os.Mkdir(".athina", 0755)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	// Ensure that there exists a .athina/objects folder in the current directory. If not, create one
	if _, err := os.Stat(".athina/objects"); os.IsNotExist(err) {
		err := os.Mkdir(".athina/objects", 0755)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	return nil
}

func saveAthinaFileObject(f AthinaFile) error {

	// Convert the File object to a Json object
	file, err := os.Create(".athina/objects/" + f.Filename)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(f)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func loadAthinaFileObject(identifier string) (AthinaFile, error) {

	// In '.athina/objects', there should be a file with the name of the identifier
	// If the file does not exist, return an error, otherwise we can load the file in as a File object
	if _, err := os.Stat(".athina/objects/" + identifier); os.IsNotExist(err) {
		return AthinaFile{}, err
	}

	// Load it as a Json object into a File struct
	file, err := os.Open(".athina/objects/" + identifier)
	if err != nil {
		fmt.Println(err)
		return AthinaFile{}, err
	}
	defer file.Close()

	var f AthinaFile
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&f)
	if err != nil {
		fmt.Println(err)
		return AthinaFile{}, err
	}

	return f, nil

}

func emulateDeltaDiffsFromAthinaFileObject(athinafile AthinaFile) (string, error) {

	dmp := diffmatchpatch.New()
	origin := athinafile.Origin
	for _, filediff := range athinafile.Diffs {
		if filediff.Deleted && filediff.Added {
			continue
		}

		if isDeltaEmpty(filediff.Delta) {
			continue
		}

		new_diff, err := dmp.DiffFromDelta(origin, filediff.Delta)
		if err != nil {
			fmt.Println(err)
			return "", err
		}

		origin = dmp.DiffText2(new_diff)
	}

	return origin, nil

}

func diffAthinaFileObjectAndString(athinafile AthinaFile, filestring string) (string, error) {

	// Every AthinaFile has one initial diff, which is the diff between the original file and the original file
	// and then every subsequent diff is stored as a Delta in the Filediff object

	// First, we want to reconstruct the current state of the Athina File
	current_state, err := emulateDeltaDiffsFromAthinaFileObject(athinafile)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	// Then, we want to compare the current state of the Athina File with the current state of the file in the directory
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(current_state, filestring, false)
	delta := dmp.DiffToDelta(diffs)
	return delta, nil
}

func diffAthinaFileAndString(athinafilename string, filestring string) (string, error) {

	athinafile, err := loadAthinaFileObject(athinafilename)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return diffAthinaFileObjectAndString(athinafile, filestring)
}

func diffAthinaFileObjectAndFile(athinafile AthinaFile, filename string) (string, error) {
	// Load the regular file

	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	defer file.Close()

	// Read the content of the file
	fileinfo, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	filesize := fileinfo.Size()
	filecontent := make([]byte, filesize)
	_, err = file.Read(filecontent)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	// Convert the content of the file to a string
	filestring := string(filecontent)

	// Compare the AthinaFile object with the file in the directory
	return diffAthinaFileObjectAndString(athinafile, filestring)
}

func AthinaUpdate(options AthinaUpdateOptions) error {

	if options.symbol == "-all" {
		fmt.Println("Updating all files")
		for change := range AthinaLookForFileChanges() {
			switch change.action {
			case AthinaFileChangeActionAdd:
				AthinaAddFile(change.filename)
			case AthinaFileChangeActionDelete:
				AthinaDeleteFile(change.filename)
			case AthinaFileChangeActionModify:
				AthinaModifyFile(change.file, change.diffs)
			case AthinaFileChangeActionError:
				return change.err

			}
		}
	} else if options.symbol != "" {
		fmt.Println("Updating file: " + options.symbol)
		err := AthinaUpdateFile(options.symbol)
		if err != nil {
			fmt.Println(err)
			return err
		}

	} else {
		return errors.New("invalid options")
	}

	return nil

}

// This is the same as eg. 'git add <filename>'
func AthinaUpdateFile(filename string) error {

	// If the file does not exist but there exists a AthinaFile object for the file, we want to mark that the file has been deleted
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		if _, err := os.Stat(".athina/objects/" + filename); !os.IsNotExist(err) {
			err := AthinaDeleteFile(filename)
			if err != nil {
				fmt.Println(err)
				return err
			}
		}

		return nil
	}

	// If there is no AthinaFile object for this file, that means that this is a new file and we want to create one
	if _, err := os.Stat(".athina/objects/" + filename); os.IsNotExist(err) {
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
	if !isDeltaEmpty(diffs) {

		filediff := createFilediff(createFilediffOptions{deleted: false, change: AthinaFileChangeActionModify, delta: diffs})

		AthinaModifyFile(athinafile, []Filediff{filediff})

	} else {
		fmt.Println("No changes detected in file: " + filename)
	}

	return nil

}

func AthinaAddFile(filename string) error {

	fmt.Println("Adding file: " + filename)

	// Create a new AthinaFile object
	athinafile, err := createInitialAthinaFileObject(filename)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Save the AthinaFile object
	err = saveAthinaFileObject(athinafile)
	if err != nil {
		fmt.Println(err)
		return err

	}

	return nil
}

func AthinaDeleteFile(filename string) error {

	fmt.Println("Deleting file: " + filename)

	// Load the Athina object
	athinafile, err := loadAthinaFileObject(filename)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Make a new Filediff object
	filediff := createFilediff(createFilediffOptions{deleted: true, change: AthinaFileChangeActionDelete})
	athinafile.Diffs = append(athinafile.Diffs, filediff)

	err = saveAthinaFileObject(athinafile)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func AthinaModifyFile(athinafile AthinaFile, diffs []Filediff) error {

	fmt.Println("Modifying file: " + athinafile.Filename)

	// Append the new diffs to the Athina object
	athinafile.Diffs = append(athinafile.Diffs, diffs...)

	// Save the Athina object
	err := saveAthinaFileObject(athinafile)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func ModifyFile(filename string, new_content string) error {

	// Overwrite the file with the new content
	err := os.WriteFile(filename, []byte(new_content), 0644)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func AthinaDetectFileChange(filename string) (AthinaFileChange, error) {

	// Check if the file exists in the .athina/objects folder
	if _, err := os.Stat(".athina/objects/" + filename); os.IsNotExist(err) {
		return AthinaFileChange{action: AthinaFileChangeActionAdd, filename: filename}, nil
	}

	// Check if the file exists in the current directory
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return AthinaFileChange{action: AthinaFileChangeActionDelete, filename: filename}, nil
	}

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
	if !isDeltaEmpty(diff) {
		fmt.Println("Changes detected in file: " + filename)
		return AthinaFileChange{action: AthinaFileChangeActionModify, file: athinafile, filename: filename, diffs: athinafile.Diffs}, nil
	} else {
		return AthinaFileChange{action: AthinaFileChangeActionNone}, nil

	}

	// Test

}

func AthinaLookForFileChanges() <-chan AthinaFileChange {

	ch := make(chan AthinaFileChange)

	go func() {

		// Go through each file in the .athina/objects folder, and compare it to the current file in the directory
		// If there are changes, print out the changes
		// If there are no changes, do nothing

		// Get all the files in the .athina/objects folder
		files, err := os.ReadDir(".athina/objects")
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

				if !isDeltaEmpty(diff) {
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
			if currentfile.IsDir() {
				continue
			}

			if _, err := os.Stat(".athina/objects/" + currentfile.Name()); errors.Is(err, os.ErrNotExist) {
				// File exists in the current directory, but not in the .athina/objects folder
				// This means that the file has been added
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
	filediff := createFilediff(createFilediffOptions{added: true, change: AthinaFileChangeActionAdd})
	file.Diffs = append(file.Diffs, filediff)

	return file, nil

}

func AthinaPrintFileHistory(filename string, depth int) error {

	// Load the Athina object
	athinafile, err := loadAthinaFileObject(filename)
	if err != nil {
		fmt.Println(err)
		return err
	}

	for i := len(athinafile.Diffs) - 1; i >= 0 && depth > 0; i-- {
		diff := athinafile.Diffs[i]
		fmt.Println("Hash: " + diff.Hash)
		fmt.Println("Change: " + string(diff.Change))
		fmt.Println("Diff: " + diff.Delta)
		depth--

	}

	return nil
}

func AthinaRevertFileByHash(filename string, hash string) {

	// Load the Athina object
	athinafile, err := loadAthinaFileObject(filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	AthinaRevertFileObjectByHash(athinafile, hash)

}

func AthinaRevertFileObjectByHash(athinafile AthinaFile, hash string) {

	// First we check if there exists a filediff with this hash

	var found bool = false
	for _, filediff := range athinafile.Diffs {
		if filediff.Hash == hash {
			found = true
			break

		}
	}

	if !found {
		fmt.Println("No such hash found")
		return
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
	filediff := createFilediff(createFilediffOptions{deleted: false, change: AthinaFileChangeActionRevert, delta: diff})

	AthinaModifyFile(athinafile, []Filediff{filediff})
	ModifyFile(athinafile.Filename, origin)

	fmt.Println("File " + athinafile.Filename + " has been reverted to hash " + hash)

}

func main() {

	//@TODO : Currently we're also saving diffs of type 0 (no change), we should use the transform delta thing

	// Initialize the .athina folder
	initializeAthinaFolder()

	args := os.Args[1:]

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
			}
		}

		return

	}

	switch args[0] {

	case "help":
		fmt.Println("Usage: athina <command> <filename>")
		fmt.Println("Commands:")
		fmt.Println("  update <filename> - Check for changes and update the Athina history")
		return

	case "update":
		if len(args) < 2 {
			fmt.Println("Usage: athina update <filename>|-all")
			return
		}

		err := AthinaUpdate(AthinaUpdateOptions{symbol: args[1]})
		if err != nil {
			fmt.Println(err)
			return
		}

	case "history":
		if len(args) < 2 {
			fmt.Println("Usage: athina history <filename>")
			return
		}

		err := AthinaPrintFileHistory(args[1], 1)
		if err != nil {
			fmt.Println(err)
			return
		}

	case "revert":
		if len(args) != 3 {
			fmt.Println("Usage: athina revert <filename> <hash>")
			return
		}

		AthinaRevertFileByHash(args[1], args[2])

	case "reset":
		// if -f is passed, reset all files, otherwise prompt
		// if -f is not passed, prompt
		// if -f is passed, reset all files

		if len(args) < 2 || args[1] != "-f" {
			fmt.Println("Are you sure you want to reset all files? (y/n)")
			var response string
			fmt.Scanln(&response)
			if response != "y" {
				return
			}
		}

		// Remove .athina
		err := os.RemoveAll(GOGIT_FOLDER)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Initialize the .athina folder
		initializeAthinaFolder()

		fmt.Println("All files have been reset")

	}

}

// Test 1
// Test 2
// Test 3
// Test 4
