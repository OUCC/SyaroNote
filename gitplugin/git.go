package main

import (
	pb "github.com/OUCC/syaro/gitservice"

	"github.com/libgit2/git2go"

	"fmt"
	"time"
)

const (
	ENTRY_LIMIT = 100
)

func getRepo() *git.Repository {
	repo, err := git.OpenRepository(repoRoot)
	if err != nil {
		log.Panic(err)
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
		log.Debug("parent commit: %s", parent.Message())
	} else {
		log.Debug("parent not found (initial commit)")
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
		log.Error("failed to look up user.name: %s", err)
		return nil
	}
	email, err := config.LookupString("user.email")
	if err != nil {
		log.Error("failed to look up user.email: %s", err)
		return nil
	}
	return &git.Signature{
		Name:  name,
		Email: email,
		When:  time.Now(),
	}
}

func logCommit(c *git.Commit) {
	log.Notice("Git commit created")
	log.Info("Message: %s", c.Message())
	log.Info("Author: %s <%s>", c.Author().Name, c.Author().Email)
	log.Info("Committer: %s <%s>", c.Committer().Name, c.Committer().Email)
}

func getChanges(repo *git.Repository, wpath string) ([]*pb.Change, error) {
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
		return nil, fmt.Errorf("%s not found in HEAD", wpath)
	}

	oid := entry.Id
	name := wpath[1:]
	previous := head
	changes := make([]*pb.Change, 0)

	// revision walking func
	fun := func(c *git.Commit) bool {
		tree, _ := c.Tree()
		found := false

		// tree walking
		tree.Walk(func(dir string, entry *git.TreeEntry) int {
			if dir+entry.Name == name && entry.Id.Equal(oid) {
				// found (not changed)
				found = true
				return -1 // end tree walking

			} else if dir+entry.Name == name && !entry.Id.Equal(oid) { // found (by name)
				// entry is found but its contents is differ
				log.Debug("%s is updated in %s", name, previous.Id().String()[:7])

				changes = append(changes, &pb.Change{
					Op:    pb.Change_OpUpdate,
					Name:  previous.Author().Name,
					Email: previous.Author().Email,
					Msg:   previous.Message(),
				})
				oid = entry.Id.Copy()
				found = true
				return -1 // end tree walking

			} else if entry.Id.Equal(oid) { // found (by oid)
				// found a contents but its name is differ
				log.Debug("%s is renamed to %s in %s", entry.Name, name, previous.Id().String()[:7])

				changes = append(changes, &pb.Change{
					Op:    pb.Change_OpRename,
					Name:  previous.Author().Name,
					Email: previous.Author().Email,
					Msg:   previous.Message(),
				})
				name = entry.Name
				found = true
				return -1 // end tree walking
			}
			return 0 // continue tree walking
		})

		if !found {
			// contents not found
			log.Debug("%s is added in %s", name, previous.Id().String()[:7])

			changes = append(changes, &pb.Change{
				Op:    pb.Change_OpAdd,
				Name:  previous.Author().Name,
				Email: previous.Author().Email,
				Msg:   previous.Message(),
			})
			return false // end rev walking
		}

		previous = c
		if len(changes) == ENTRY_LIMIT { // limit number of entry
			return false // end rev walking
		}
		return true // continue rev walking
	}

	if err := walk.Iterate(fun); err != nil {
		return nil, fmt.Errorf("Error while rev walking: %s", err)
	}

	if l := len(changes); l < ENTRY_LIMIT && (l == 0 || changes[l-1].Op != pb.Change_OpAdd) {
		// file is added in last commit
		changes = append(changes, &pb.Change{
			Op:    pb.Change_OpAdd,
			Name:  previous.Author().Name,
			Email: previous.Author().Email,
			Msg:   previous.Message(),
		})
	}
	return changes, nil
}

func getAllChanges(repo *git.Repository) ([]*pb.Change, error) {
	// setup revision walker
	walk, _ := repo.Walk()
	walk.Sorting(git.SortTopological | git.SortTime)
	walk.PushHead()

	changes := make([]*pb.Change, 0)

	// walking func
	fun := func(c *git.Commit) bool {
		changes = append(changes, &pb.Change{
			Op:    pb.Change_OpNone,
			Name:  c.Author().Name,
			Email: c.Author().Email,
			Msg:   c.Message(),
		})

		if len(changes) == ENTRY_LIMIT {
			return false
		}
		return true
	}

	if err := walk.Iterate(fun); err != nil {
		return nil, err
	}

	return changes, nil
}
