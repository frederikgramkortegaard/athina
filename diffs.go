package main

import (
	"fmt"
	"os"

	"github.com/sergi/go-diff/diffmatchpatch"
)

func emulateDeltaDiffsFromAthinaFileObject(athinafile AthinaFile) (string, error) {

	dmp := diffmatchpatch.New()
	origin := athinafile.Origin
	for _, filediff := range athinafile.Diffs {
		if filediff.Deleted && filediff.Added {
			continue
		}

		if isDeltaDiffEmpty(filediff.Delta) {
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
