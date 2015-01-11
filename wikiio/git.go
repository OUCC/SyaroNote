package wikiio

import (
	. "github.com/OUCC/syaro/logger"

	"github.com/libgit2/git2go"

	"time"
)

func commitChange(manip func(idx *git.Index) error, sig *git.Signature,
	message string) (*git.Commit, error) {
	// staging and get index tree
	idx, _ := repo.Index()
	defer idx.Free()
	if err := manip(idx); err != nil {
		return nil, err
	}
	treeid, _ := idx.WriteTree()
	tree, _ := repo.LookupTree(treeid)
	defer tree.Free()

	// get latest commit of current branch
	parent := getLastCommit()
	if parent != nil {
		defer parent.Free()
		Log.Debug("parent commit: %s", parent.Message())
	} else {
		Log.Debug("parent not found (initial commit)")
	}

	// commit
	var oid *git.Oid
	var err error
	if parent != nil {
		oid, err = repo.CreateCommit("HEAD", sig, sig, message, tree, parent)
	} else {
		oid, err = repo.CreateCommit("HEAD", sig, sig, message, tree)
	}
	if err != nil {
		return nil, err
	}

	// save index
	idx.Write()

	commit, _ := repo.LookupCommit(oid)
	return commit, nil
}

// getLastCommit returns latest commit of current branch
func getLastCommit() *git.Commit {
	ref, _ := repo.LookupReference("HEAD")
	defer ref.Free()
	ref, _ = ref.Resolve()
	var parent *git.Commit
	if ref != nil {
		parent, _ = repo.LookupCommit(ref.Target())
		return parent
	} else {
		return nil
	}
}

func getDefaultSignature() *git.Signature {
	config, _ := repo.Config()
	defer config.Free()
	name, err := config.LookupString("user.name")
	if err != nil {
		Log.Error("Git error: %s", err)
		return nil
	}
	email, err := config.LookupString("user.email")
	if err != nil {
		Log.Error("Git error: %s", err)
		return nil
	}
	return &git.Signature{
		Name:  name,
		Email: email,
		When:  time.Now(),
	}
}

func logCommit(c *git.Commit) {
	Log.Notice("Git commit created")
	Log.Info("Message: %s", c.Message())
	Log.Info("Author: %s <%s>", c.Author().Name, c.Author().Email)
	Log.Info("Committer: %s <%s>", c.Committer().Name, c.Committer().Email)
}
