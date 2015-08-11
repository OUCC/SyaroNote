package main

import (
	pb "github.com/OUCC/syaro/gitservice"

	"github.com/libgit2/git2go"
	"golang.org/x/net/context"

	"fmt"
	"path/filepath"
)

type GitService struct{}

func (gs *GitService) Save(c context.Context, req *pb.SaveRequest) (*pb.CommitResponse, error) {
	repo := getRepo()
	sig := getDefaultSignature(repo)
	if req.Name != "" {
		sig.Name = req.Name
		sig.Email = req.Email
	}

	message := req.Msg
	if message == "" {
		message = "Updated " + filepath.Base(req.Path)
	}

	commit, err := commitChange(repo,
		func(idx *git.Index) error {
			if err := idx.AddByPath(req.Path[1:]); err != nil {
				return err
			}
			return nil
		},
		sig,
		message)
	if err != nil {
		log.Error("failed to commit modification: %s", err)
		return nil, err
	}
	defer commit.Free()
	logCommit(commit)

	sig = commit.Author()
	return &pb.CommitResponse{
		Name:  sig.Name,
		Email: sig.Email,
		Msg:   commit.Message(),
	}, nil
}

func (gs *GitService) Remove(c context.Context, req *pb.RemoveRequest) (*pb.CommitResponse, error) {
	repo := getRepo()
	sig := getDefaultSignature(repo)
	if req.Name != "" {
		sig.Name = req.Name
		sig.Email = req.Email
	}

	commit, err := commitChange(repo,
		func(idx *git.Index) error {
			return idx.RemoveAll(
				[]string{req.Path[1:]},
				func(path, spec string) int {
					log.Debug("git: removing %s", path)
					return 0
				})
		},
		sig,
		"Removed "+filepath.Base(req.Path))
	if err != nil {
		log.Error("failed to commit removal: %s", err)
		return nil, err
	}
	defer commit.Free()
	logCommit(commit)

	sig = commit.Author()
	return &pb.CommitResponse{
		Name:  sig.Name,
		Email: sig.Email,
		Msg:   commit.Message(),
	}, nil
}

func (gs *GitService) Rename(c context.Context, req *pb.RenameRequest) (*pb.CommitResponse, error) {
	repo := getRepo()
	sig := getDefaultSignature(repo)
	if req.Name != "" {
		sig.Name = req.Name
		sig.Email = req.Email
	}

	commit, err := commitChange(repo,
		func(idx *git.Index) error {
			err := idx.RemoveAll(
				[]string{req.Src[1:]},
				func(path, spec string) int {
					log.Debug("git: removing %s", path)
					return 0
				})
			if err != nil {
				return err
			}
			return idx.AddAll(
				[]string{req.Dst[1:]},
				git.IndexAddDefault,
				func(path, spec string) int {
					log.Debug("git: adding %s", path)
					return 0
				})
		},
		sig,
		fmt.Sprintf("Renamed %s -> %s", req.Src, req.Dst))

	if err != nil {
		log.Error("failed to commit renaming: %s", err)
		return nil, err
	}
	defer commit.Free()
	logCommit(commit)

	sig = commit.Author()
	return &pb.CommitResponse{
		Name:  sig.Name,
		Email: sig.Email,
		Msg:   commit.Message(),
	}, nil
}

func (gs *GitService) Changes(req *pb.ChangesRequest, stream pb.Git_ChangesServer) error {
	repo := getRepo()
	changes, err := getChanges(repo, req.Path)
	if err != nil {
		log.Error("failed to get changes:%s", err)
		return err
	}
	for _, c := range changes {
		if err = stream.Send(c); err != nil {
			return err
		}
	}
	return nil
}
