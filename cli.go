package main

import (
	"fmt"
	"os"
	"strconv"
)

func resetAthinaRepository() {

	err := os.RemoveAll(ATHINA_FOLDER)
	if err != nil {
		fmt.Println(err)
		return
	}

	initializeAthinaFolder()

}

func addFileToIgnoreList(filename string) {

	// Check if the filename is already ignored
	found := false
	for _, ignored := range config.Ignored {
		if ignored == filename {
			found = true
			return
		}
	}

	if !found {
		config.Ignored = append(config.Ignored, filename)
		config.Save()
	}

}

func initializeAthinaFolder() error {

	// Ensure that there exists a .athina folder in the current directory. If not, create one
	if _, err := os.Stat(ATHINA_FOLDER); os.IsNotExist(err) {
		err := os.Mkdir(ATHINA_FOLDER, 0755)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	// Ensure that there exists a .athina/objects folder in the current directory. If not, create one
	if _, err := os.Stat(ATHINA_PATH_TO_OBJECTS); os.IsNotExist(err) {
		err := os.Mkdir(ATHINA_PATH_TO_OBJECTS, 0755)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	// Make the config file if it does not exist
	if _, err := os.Stat(ATHINA_CONFIG); os.IsNotExist(err) {
		file, err := os.Create(ATHINA_CONFIG)
		if err != nil {
			fmt.Println(err)
			return err
		}
		defer file.Close()

		_, err = file.WriteString(`{"ignored":[]}`)
		if err != nil {
			fmt.Println(err)
			return err
		}

	}

	return nil
}

const DEFAULT_HISTORY_DEPTH int = 5

func printFileHistory(filename string, depth int) error {

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
		fmt.Println("Diff (Delta): " + diff.Delta)
		depth--

	}

	return nil
}

func AthinaListFiles() []string {

	files, err := os.ReadDir(ATHINA_PATH_TO_OBJECTS)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	var filenames []string
	for _, file := range files {
		filenames = append(filenames, file.Name())
	}

	return filenames
}

// Removes and re-initializes the .athina/object entry for said file
func AthinaResetFile(filename string) error {

	// Remove the file from the .athina/objects directory
	err := os.Remove(ATHINA_PATH_TO_OBJECTS + filename)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Re-initialize the file
	err = AthinaAddFile(filename)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil

}

func handleCLI(args []string) {

	// @NOTE: Args has already had the first element removed, meaning that args[0] is the first argument
	if len(args) == 0 {
		fmt.Println("Usage: athina [command] [args]")
		return
	}

	switch args[0] {

	case "update": //@NOTE : This is basically a combination of the add and commit commands

		// Update a specific file
		if len(args) == 2 {
			err := AthinaUpdateFile(args[1])
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("File \"" + args[1] + "\" has been updated")
		} else {
			// Update all files
			err := AthinaUpdateAllFiles(true) //@NOTE : 'true' enables logging to stdout
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("All files have been updated")
		}

	case "help":
		fmt.Println("Usage: athina [command] [args]")
		fmt.Println("Commands:")
		fmt.Println("  init: Initialize the Athina repository")
		fmt.Println("  reset: Reset the Athina repository")
		fmt.Println("  history [filename] [depth]: Print the history of a file")
		fmt.Println("  ignore [filename]: Add a file to the ignore list")
		fmt.Println("  revert [filename] [hash]: Revert a file to a previous state")

	case "revert":
		if len(args) < 3 {
			fmt.Println("Usage: athina revert [filename] [hash]")
			return
		}

		err := AthinaRevertFileByHash(args[1], args[2])
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("File \"" + args[1] + "\" has been reverted to hash \"" + args[2] + "\"")

	case "init":
		initializeAthinaFolder()
		fmt.Println("Athina has been initialized")
	case "reset":

		// Reset a file
		if len(args) == 2 {
			err := AthinaResetFile(args[1])
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("File \"" + args[1] + "\" has been reset")
			return
		}

		resetAthinaRepository()
		fmt.Println("Athina has been reset")
	case "history":
		if len(args) < 2 {
			fmt.Println("Usage: athina history [filename] [depth]")
			return
		}

		if len(args) == 3 {
			depth, err := strconv.Atoi(args[2])
			if err != nil {
				fmt.Println("Depth must be an integer, but got: " + args[2])
				return
			}
			printFileHistory(args[1], depth)
		} else {
			printFileHistory(args[1], DEFAULT_HISTORY_DEPTH)
		}

	case "ignore":
		if len(args) < 2 {
			fmt.Println("Usage: athina ignore [filename]")
			return
		}

		addFileToIgnoreList(args[1])
		fmt.Println("File \"" + args[1] + "\" has been added to the ignore list")

	case "list":
		if len(args) == 1 {
			fmt.Println("Usage: athina list [files|ignored]")
			return
		}

		if args[1] == "files" {
			for _, file := range AthinaListFiles() {
				fmt.Println(file)
			}
		} else if args[1] == "ignored" {
			for _, ignored := range config.Ignored {
				fmt.Println(ignored)
			}
		} else {

			fmt.Println("Usage: athina list [files|ignored]")
			return
		}

	default:
		fmt.Println("Usage: athina [command] [args]")
	}
}
