package util

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/integralist/go-findroot/find"
	"github.com/pkg/errors"
)

func GetCurrentModule() (string, error) {
	rep, err := find.Repo()
	if err != nil {
		return "", errors.Wrap(err, "Please set up a git repository before using the generator")
	}
	root := rep.Path
	file, err := os.Open(root + "/go.mod")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	sp := strings.Split(scanner.Text(), "module ")
	return sp[1], nil
}
