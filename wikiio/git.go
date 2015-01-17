package wikiio

import (
	. "github.com/OUCC/syaro/logger"

	"github.com/libgit2/git2go"

	"time"
)

type Change struct {
	Op     Op
	Commit *git.Commit
}

type Op int

const (
	OpAdd Op = iota
	OpUpdate
	OpRename
	OpRemove
)

func OpString(op Op) string {
	switch op {
	case OpAdd:
		return "Add"
	case OpUpdate:
		return "Edit"
	case OpRename:
		return "Rename"
	case OpRemove:
		return "Remove"
	}
	return ""
}

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

func getChanges(wpath string) []Change {
	// setup revision walker
	walk, _ := repo.Walk()
	walk.Sorting(git.SortTopological | git.SortTime)
	walk.PushHead()

	// find out current oid of the wpath
	head := getLastCommit()
	tree, _ := head.Tree()
	entry, err := tree.EntryByPath(wpath[1:])
	if err != nil {
		Log.Error("%s not found in HEAD", wpath)
		return nil
	}

	oid := entry.Id
	name := wpath[1:]
	previous := head
	found := true
	changes := make([]Change, 0)

	// walking func
	fun := func(c *git.Commit) bool {
		tree, _ := c.Tree()
		if entry := tree.EntryByName(name); entry != nil { // found (by name)
			if !entry.Id.Equal(oid) {
				oid = entry.Id.Copy()
				if found {
					Log.Debug("%s is updated in %s", name, previous.Id().String()[:7])

					changes = append(changes, Change{
						Op:     OpUpdate,
						Commit: previous,
					})
				} else {
					found = true
					Log.Debug("%s is removed in %s", name, previous.Id().String()[:7])

					changes = append(changes, Change{
						Op:     OpRemove,
						Commit: previous,
					})
				}
			}
		} else { // not found (by name)
			var i uint64
			ok := false
			for i = 0; i < tree.EntryCount(); i++ {
				entry := tree.EntryByIndex(i)
				if entry.Id.Equal(oid) && found { // found (by oid)
					name = entry.Name
					Log.Debug("%s is renamed in %s", name, previous.Id().String()[:7])

					changes = append(changes, Change{
						Op:     OpRename,
						Commit: previous,
					})
					ok = true
					break
				}
			}
			if !ok && found {
				found = false
				Log.Debug("%s is added in %s", name, previous.Id().String()[:7])

				changes = append(changes, Change{
					Op:     OpAdd,
					Commit: previous,
				})
			}
		}

		previous = c
		return true
	}

	if err := walk.Iterate(fun); err != nil {
		Log.Debug(err.Error())
	}

	if found { // file is added in last commit
		changes = append(changes, Change{
			Op:     OpAdd,
			Commit: previous,
		})
	}

	return changes
}
