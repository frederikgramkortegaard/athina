package main

type AthinaFileChangeAction string

const (
	AthinaFileChangeActionAdd    AthinaFileChangeAction = "File add"
	AthinaFileChangeActionDelete AthinaFileChangeAction = "File delete"
	AthinaFileChangeActionModify AthinaFileChangeAction = "File modify"
	AthinaFileChangeActionRevert AthinaFileChangeAction = "File revert"
	AthinaFileChangeActionNone   AthinaFileChangeAction = "none"
	AthinaFileChangeActionError  AthinaFileChangeAction = "error"
)

type AthinaFileChange struct {
	action   AthinaFileChangeAction
	file     AthinaFile
	filename string
	diffs    []Filediff
	err      error
	delta    string
}
