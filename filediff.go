package main

import (
	"strconv"

	"github.com/sergi/go-diff/diffmatchpatch"
)

type Filediff struct {
	Hash    string
	Diffs   []diffmatchpatch.Diff
	Delta   string
	Deleted bool
	Added   bool
	Change  AthinaFileChangeAction
}

func (f Filediff) getHash() string {
	dmp := diffmatchpatch.New()
	return sha1Hash(dmp.DiffPrettyText(f.Diffs) + f.Delta + strconv.FormatBool(f.Deleted) + strconv.FormatBool(f.Added) + string(f.Change))
}

type newFileDiffOptions struct {
	diffs   []diffmatchpatch.Diff
	deleted bool
	added   bool
	change  AthinaFileChangeAction
	delta   string
}

func newFilediff(options newFileDiffOptions) Filediff {

	var filediff Filediff = Filediff{
		Hash:    "",
		Diffs:   options.diffs,
		Delta:   options.delta,
		Deleted: options.deleted,
		Added:   options.added,
		Change:  options.change,
	}

	filediff.Hash = filediff.getHash()

	return filediff
}
