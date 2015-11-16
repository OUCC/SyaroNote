package main

import (
	pb "github.com/OUCC/syaro/gitservice"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"io"
	"strconv"
)

func gitCommit(commitFunc func(pb.GitClient) (*pb.CommitResponse, error)) error {
	if !setting.gitMode {
		return nil
	}

	conn, err := grpc.Dial("127.0.0.1:" + strconv.Itoa(setting.port+1))
	if err != nil {
		log.Debug("Dial error: %s", err)
		return err
	}
	defer conn.Close()

	client := pb.NewGitClient(conn)
	res, err := commitFunc(client)
	if err != nil {
		log.Debug("Git error: %s", err)
		return err
	}
	log.Debug("commit id: %s, message: %s", res.Msg, res.Msg)
	return nil
}

func getChanges(wpath string) []*pb.Change {
	conn, err := grpc.Dial("127.0.0.1:" + strconv.Itoa(setting.port+1))
	if err != nil {
		log.Debug("Dial error: %s", err)
		return nil
	}
	defer conn.Close()

	client := pb.NewGitClient(conn)
	stream, err := client.Changes(context.Background(), &pb.ChangesRequest{
		Path: wpath,
	})
	if err != nil {
		log.Debug("Git error: %s", err)
		return nil
	}

	changes := make([]*pb.Change, 0)
	for {
		c, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Debug("Stream error: %s", err)
			return nil
		}
		changes = append(changes, c)
	}
	return changes
}
