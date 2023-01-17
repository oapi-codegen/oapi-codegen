package util

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/client"
	"github.com/go-git/go-git/v5/storage/memory"
)

var (
	errNotSshUrl   = errors.New("not an ssh url")
	errLocalPath   = errors.New("uri is a local path")
	errNotGitUrl   = errors.New("not a git url")
	errInvalidPath = errors.New("invalid path")

	sshPattern = regexp.MustCompile(`^(?:([a-zA-Z+-.]+)*://)?(?:([^@\s]+)@)([^:]+(?::\d+)?):(/?.+)$`)
)

func LoadSwagger(filePath string) (swagger *openapi3.T, err error) {

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	// extend default loader with an additional git loader
	loader.ReadFromURIFunc = openapi3.URIMapCache(readFromGit)

	u, err := parseUrl(filePath)
	if err == nil && u.Scheme != "" && u.Host != "" {
		return loader.LoadFromURI(u)
	} else {
		absolutePath, err := cleanLocalPath(filePath)
		if err != nil {
			return nil, err
		}
		return loader.LoadFromFile(absolutePath)
	}
}

func readFromGit(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
	repoUrl, filePath, branch, err := splitRepoUrl(url)
	if err != nil {
		switch url.Scheme {
		case "", "file":
			return readLocalFile(url.Path)
		default:
			return openapi3.ReadFromHTTP(http.DefaultClient)(loader, url)
		}
	}

	fs := memfs.New()
	storer := memory.NewStorage()

	o := &git.CloneOptions{
		URL:           repoUrl,
		SingleBranch:  true,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		Depth:         1,
	}

	_, err = git.Clone(storer, fs, o)
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w: %v", err, filePath)
	}

	f, err := fs.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error in cloned git repository: %w: %v", err, filePath)
	}
	defer f.Close()
	return io.ReadAll(f)
}

func splitRepoUrl(u *url.URL) (repoUrl, filePath, branch string, err error) {
	defer func() {
		// fetch default branch
		if err == nil && branch == "" {
			branch, err = fetchDefaultBranch(repoUrl)
		}
	}()

	if strings.Contains(u.Path, "@") {
		tokens := strings.Split(u.Path, "@")
		branch = tokens[len(tokens)-1]
		u.Path = strings.Join(tokens[:len(tokens)-1], "@")
	}

	pathParts := strings.Split(u.Path, "/")
	for idx, p := range pathParts {
		if strings.Contains(p, ".git") {
			if idx+1 > len(pathParts) {
				return "", "", "", fmt.Errorf("%w: invalid git url path: %q", errNotGitUrl, u.Path)
			}

			u.Path = path.Join(pathParts[:idx+1]...)
			filePath := path.Join(pathParts[idx+1:]...)
			if filePath == "" {
				return "", "", "", fmt.Errorf("%w: invalid git file path: %s", errNotGitUrl, filePath)
			}
			return u.String(), filePath, branch, nil
		}
	}

	return "", "", "", fmt.Errorf("%w: %s", errNotGitUrl, u.String())
}

func parseUrl(urlStr string) (*url.URL, error) {
	u, err := detectSSH(urlStr)
	if err == nil {
		return u, nil
	}

	u, err = url.ParseRequestURI(urlStr)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "file" {
		return nil, errLocalPath
	}

	return u, nil
}

func detectSSH(src string) (*url.URL, error) {
	matched := sshPattern.FindStringSubmatch(src)
	if len(matched) == 0 {
		return nil, errNotSshUrl
	}
	if matched[1] != "" && !strings.Contains(matched[1], "ssh") {
		return nil, fmt.Errorf("%w: %s", errNotSshUrl, src)
	}

	user := matched[2]
	host := matched[3]
	path := matched[4]
	qidx := strings.Index(path, "?")
	if qidx == -1 {
		qidx = len(path)
	}

	var u url.URL
	u.Scheme = "ssh"
	u.User = url.User(user)
	u.Host = host
	u.Path = path[0:qidx]
	if qidx < len(path) {
		q, err := url.ParseQuery(path[qidx+1:])
		if err != nil {
			return nil, fmt.Errorf("%w: %s", errNotSshUrl, err)
		}
		u.RawQuery = q.Encode()
	}

	return &u, nil
}

func cleanLocalPath(localPath string) (absolutePath string, err error) {
	localPath = strings.TrimPrefix(localPath, "file://")

	// not an url, try to get fro mlocal path
	absolutePath, err = filepath.Abs(localPath)
	if err != nil {
		return "", fmt.Errorf("%w: %v", errInvalidPath, err)
	}
	return absolutePath, nil
}

func readLocalFile(localPath string) ([]byte, error) {
	absolutePath, err := cleanLocalPath(localPath)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(absolutePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errInvalidPath, err)
	}
	return data, nil
}

func fetchDefaultBranch(repoUrl string) (branch string, err error) {
	e, err := transport.NewEndpoint(repoUrl)
	if err != nil {
		return "", err
	}
	cli, err := client.NewClient(e)
	if err != nil {
		return "", err
	}
	s, err := cli.NewUploadPackSession(e, nil)
	if err != nil {
		return "", err
	}
	info, err := s.AdvertisedReferences()
	if err != nil {
		return "", err
	}
	refs, err := info.AllReferences()
	if err != nil {
		return "", err
	}
	headReference := refs["HEAD"].Target()
	headBranch := headReference.String()

	return strings.TrimPrefix(headBranch, "refs/heads/"), nil
}
