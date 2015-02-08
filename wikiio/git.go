package wikiio

import (
	. "github.com/OUCC/syaro/logger"
	"github.com/OUCC/syaro/setting"

	"github.com/libgit2/git2go"

	"time"
)

type Change struct {
	Op     Op
	Commit *git.Commit
}

type Op int

const (
	OpNone Op = iota
	OpAdd
	OpUpdate
	OpRename

	ENTRY_LIMIT = 100
)

func OpString(op Op) string {
	switch op {
	case OpAdd:
		return "Add"
	case OpUpdate:
		return "Edit"
	case OpRename:
		return "Rename"
	}
	return ""
}

func getRepo() *git.Repository {
	if !setting.GitMode {
		return nil
	}

	repo, err := git.OpenRepository(setting.WikiRoot)
	if err != nil {
		Log.Panic(err)
	}
	return repo
}

func commitChange(repo *git.Repository, manip func(idx *git.Index) error, sig *git.Signature,
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
	parent := getLastCommit(repo)
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
func getLastCommit(repo *git.Repository) *git.Commit {
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

func getDefaultSignature(repo *git.Repository) *git.Signature {
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

func getChanges(repo *git.Repository, wpath string) []Change {
	if wpath == "/" {
		return getAllChanges(repo)
	}

	// setup revision walker
	walk, _ := repo.Walk()
	walk.Sorting(git.SortTopological | git.SortTime)
	walk.PushHead()

	// find out current oid of the wpath
	head := getLastCommit(repo)
	tree, _ := head.Tree()
	entry, err := tree.EntryByPath(wpath[1:])
	if err != nil {
		Log.Error("%s not found in HEAD", wpath)
		return nil
	}

	oid := entry.Id
	name := wpath[1:]
	previous := head
	changes := make([]Change, 0)

	// walking func
	fun := func(c *git.Commit) bool {
		tree, _ := c.Tree()
		if entry := tree.EntryByName(name); entry != nil { // found (by name)
			if !entry.Id.Equal(oid) { // entry is found but its contents is differ
				Log.Debug("%s is updated in %s", name, previous.Id().String()[:7])

				changes = append(changes, Change{
					Op:     OpUpdate,
					Commit: previous,
				})
				oid = entry.Id.Copy()
			}
		} else { // not found (by name)
			var i uint64
			found := false
			for i = 0; i < tree.EntryCount(); i++ {
				entry := tree.EntryByIndex(i)
				if entry.Id.Equal(oid) { // found (by oid)
					// found a contents but its name is differ
					Log.Debug("%s is renamed to %s in %s", entry.Name, name, previous.Id().String()[:7])

					changes = append(changes, Change{
						Op:     OpRename,
						Commit: previous,
					})
					name = entry.Name
					found = true
					break
				}
			}
			if !found {
				// contents not found
				Log.Debug("%s is added in %s", name, previous.Id().String()[:7])

				changes = append(changes, Change{
					Op:     OpAdd,
					Commit: previous,
				})
				return false // terminate walking
			}
		}

		previous = c
		if len(changes) == ENTRY_LIMIT { // limit number of entry
			return false
		}
		return true
	}

	if err := walk.Iterate(fun); err != nil {
		Log.Debug(err.Error())
	}

	if l := len(changes); l < ENTRY_LIMIT && changes[len(changes)-1].Op != OpAdd {
		// file is added in last commit
		changes = append(changes, Change{
			Op:     OpAdd,
			Commit: previous,
		})
	}
	return changes
}

func getAllChanges(repo *git.Repository) []Change {
	// setup revision walker
	walk, _ := repo.Walk()
	walk.Sorting(git.SortTopological | git.SortTime)
	walk.PushHead()

	changes := make([]Change, 0)

	// walking func
	fun := func(c *git.Commit) bool {
		changes = append(changes, Change{
			Op:     OpNone,
			Commit: c,
		})

		if len(changes) == ENTRY_LIMIT {
			return false
		}
		return true
	}

	if err := walk.Iterate(fun); err != nil {
		Log.Debug(err.Error())
	}

	return changes
}
