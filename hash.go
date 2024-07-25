package main

import (
	"crypto/sha1"

	"github.com/sergi/go-diff/diffmatchpatch"
)

func sha1Hash(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	return string(bs)
}

func hashFilediff(fd Filediff) string {

	dmp := diffmatchpatch.New()

	return sha1Hash(dmp.DiffPrettyText(fd.Diffs) + fd.Delta)
}

// Test
