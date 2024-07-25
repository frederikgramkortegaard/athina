package main

import (
	"encoding/json"
	"os"
)

type CommitItem struct {
	Hash      string     `json:"hash"`
	Filename  string     `json:"filename"`
	Filediffs []Filediff `json:"filediffs"`
}

func (c CommitItem) getHash() string {

	var hash string = c.Filename
	for _, filediff := range c.Filediffs {
		hash += filediff.getHash()
	}

	return sha1Hash(hash)

}

func newCommitItem(filename string, filediffs []Filediff) CommitItem {

	var commitItem CommitItem = CommitItem{
		Hash:      "",
		Filename:  filename,
		Filediffs: filediffs,
	}

	commitItem.Hash = commitItem.getHash()

	return commitItem

}

type Commit struct {
	Hash  string       `json:"hash"`
	Items []CommitItem `json:"items"`
}

func (c Commit) getHash() string {

	var hash string = ""
	for _, item := range c.Items {
		hash += item.getHash()
	}

	return sha1Hash(hash)

}

func newCommit(items []CommitItem) Commit {

	var commit Commit = Commit{
		Hash:  "",
		Items: items,
	}

	commit.Hash = commit.getHash()

	return commit

}

func (c Commit) addCommitItem(item CommitItem) Commit {
	c.Items = append(c.Items, item)
	c.Hash = c.getHash()
	return c
}

func (c Commit) removeCommitItem(item CommitItem) Commit {
	var newItems []CommitItem
	for _, i := range c.Items {
		if i.Hash != item.Hash {
			newItems = append(newItems, i)
		}
	}
	c.Items = newItems
	c.Hash = c.getHash()
	return c
}

func (c Commit) getCommitItemByFilename(filename string) CommitItem {
	for _, item := range c.Items {
		if item.Filename == filename {
			return item
		}
	}
	return CommitItem{}
}

func (c Commit) getCommitItemByHash(hash string) CommitItem {
	for _, item := range c.Items {
		if item.Hash == hash {
			return item
		}
	}
	return CommitItem{}
}

type Stash struct {
	Stashes []Commit `json:"stashes"`
}

func (s Stash) getCommitByHash(hash string) Commit {
	for _, commit := range s.Stashes {
		if commit.Hash == hash {
			return commit
		}
	}
	return Commit{}
}

func loadStash() (Stash, error) {

	file, err := os.ReadFile(ATHINA_STASH)
	if err != nil {
		return stash, err
	}

	err = json.Unmarshal(file, &stash)
	if err != nil {
		return stash, err
	}

	return stash, nil

}

func (s Stash) Save() error {

	stashJSON, err := json.Marshal(s)
	if err != nil {
		return err
	}

	err = os.WriteFile(ATHINA_STASH, stashJSON, 0644)
	if err != nil {
		return err
	}

	return nil

}
