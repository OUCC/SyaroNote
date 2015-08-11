package main

import (
	pb "github.com/OUCC/syaro/gitservice"

	"google.golang.org/grpc"

	"net"
	"os"
)

var repoRoot string

// arg1: :<port number>
// arg2: repository root path
func main() {
	if len(os.Args) != 3 {
		panic("invalid argument")
	}

	repoRoot = os.Args[2]
	setupLogger()

	lis, err := net.Listen("tcp", os.Args[1])
	if err != nil {
		log.Fatal("failed to listen: ", err)
	}
	server := grpc.NewServer()

	pb.RegisterGitServer(server, new(GitService))
	server.Serve(lis)
}
