package main

import (
	pb "github.com/OUCC/syaro/gitservice"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"html/template"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const BACKUP_SUFFIX = "~"

type WikiFile struct {
	fileInfo  os.FileInfo
	files     []*WikiFile
	parentDir *WikiFile
	wikiPath  string
}

// Base name of file (with ext)
func (f *WikiFile) Name() string {
	return filepath.Base(f.wikiPath)
}

func (f *WikiFile) NameWithoutExt() string {
	return removeExt(f.Name())
}

func (f *WikiFile) WikiPath() string { return f.wikiPath }

// WikiPathList returns slice of each WikiFile in wikipath
// (slice doesn't include urlPrefix)
func (f *WikiFile) WikiPathList() []*WikiFile {
	log.Debug("building...")
	s := strings.Split(removeExt(f.WikiPath()), "/")
	if s[0] == "" {
		s = s[1:]
	}

	ret := make([]*WikiFile, len(s))
	for i := 0; i < len(ret); i++ {
		path := "/" + strings.Join(s[:i+1], "/")
		log.Debug("load %s", path)
		//		p, err := LoadPage(path)
		wfile, err := loadFile(path)
		if err != nil {
			log.Debug("error in wikiio.Load(path): %s", err)
		}
		ret[i] = wfile
	}
	log.Debug("finish")
	return ret
}

// WIKIROOT/a/b/c.md
func (f *WikiFile) FilePath() string {
	return filepath.Join(setting.wikiRoot, f.wikiPath)
}

// URLPREFIX/a/b/c.md
func (f *WikiFile) URLPath() template.URL {
	path := filepath.Join(setting.urlPrefix, f.wikiPath)

	// url escape and revert %2F -> /
	return template.URL(strings.Replace(url.QueryEscape(path), "%2F", "/", -1))
}

func (f *WikiFile) IsDir() bool { return f.fileInfo.IsDir() }

func (f *WikiFile) IsDirMainPage() bool {
	return !f.IsDir() &&
		(strings.HasPrefix(f.WikiPath(), "/Home.") ||
			f.NameWithoutExt() == f.ParentDir().Name())
}

func (f *WikiFile) DirMainPage() *WikiFile {
	if !f.IsDir() {
		return nil
	}

	var name string
	if f.WikiPath() == "/" {
		name = "Home"
	} else {
		name = f.Name()
	}

	for _, file := range f.files {
		if file.NameWithoutExt() == name {
			return file
		}
	}

	// not found
	return nil
}

func (f *WikiFile) IsMarkdown() bool { return isMarkdown(f.wikiPath) }

func (f *WikiFile) Files() []*WikiFile { return f.files }

func (f *WikiFile) Folders() []*WikiFile {
	ret := make([]*WikiFile, 0, len(f.files))
	for _, file := range f.files {
		if file.IsDir() {
			ret = append(ret, file)
		}
	}
	return ret
}

func (f *WikiFile) MdFiles() []*WikiFile {
	ret := make([]*WikiFile, 0, len(f.files))
	for _, file := range f.files {
		if file.IsMarkdown() {
			ret = append(ret, file)
		}
	}
	return ret
}

func (f *WikiFile) OtherFiles() []*WikiFile {
	ret := make([]*WikiFile, 0, len(f.files))
	for _, file := range f.files {
		if !file.IsDir() && !file.IsMarkdown() {
			ret = append(ret, file)
		}
	}
	return ret
}

func (f *WikiFile) ParentDir() *WikiFile { return f.parentDir }

func (f *WikiFile) Raw() []byte {
	if f.IsDir() {
		return nil
	}

	b, err := ioutil.ReadFile(f.FilePath())
	if err != nil {
		log.Fatal(err)
	}

	return b
}

func (f *WikiFile) RawBackup() []byte {
	if f.IsDir() {
		return nil
	}

	b, _ := ioutil.ReadFile(f.FilePath() + BACKUP_SUFFIX)
	return b
}

func (f *WikiFile) Save(b []byte, message, name, email string) error {
	if err := ioutil.WriteFile(f.FilePath(), b, 0644); err != nil {
		return err
	}

	f.RemoveBackup()

	if strings.TrimSpace(message) == "" {
		message = "Updated " + f.wikiPath
	}

	return gitCommit(func(client pb.GitClient) (*pb.CommitResponse, error) {
		return client.Save(context.Background(), &pb.SaveRequest{
			Path:  f.wikiPath,
			Name:  name,
			Email: email,
			Msg:   message,
		})
	})
}

func (f *WikiFile) SaveBackup(b []byte) error {
	return ioutil.WriteFile(f.FilePath()+BACKUP_SUFFIX, b, 0644)
}

func (f *WikiFile) Remove() error {
	if err := os.RemoveAll(f.FilePath()); err != nil {
		return err
	}

	refreshRequired = true

	return gitCommit(func(client pb.GitClient) (*pb.CommitResponse, error) {
		return client.Remove(context.Background(), &pb.RemoveRequest{
			Path: f.wikiPath,
			Msg:  "Removed " + f.wikiPath,
		})
	})
}

func (f *WikiFile) RemoveBackup() error {
	return os.Remove(f.FilePath() + BACKUP_SUFFIX)
}

func (f *WikiFile) History() []*pb.Change {
	if setting.gitMode {
		conn, err := grpc.Dial("127.0.0.1:" + strconv.Itoa(setting.port+1))
		if err != nil {
			log.Debug("Dial error: %s", err)
			return nil
		}
		defer conn.Close()

		client := pb.NewGitClient(conn)
		stream, err := client.Changes(context.Background(), &pb.ChangesRequest{
			Path: f.wikiPath,
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
	} else {
		return nil
	}
}
