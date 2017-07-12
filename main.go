package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		terror(errors.New("missing gopath"))
	}

	gopath, err := format(os.Args[1])
	if err != nil {
		terror(err)
	}

	shrc, err := shellrc()
	if err != nil {
		terror(err)
	}

	if err := backup(shrc, shrc+".pre.gos"); err != nil {
		terror(err)
	}

	if err := modify(shrc, gopath); err != nil {
		terror(err)
	}
}

func format(gopath string) (string, error) {
	if gopath[len(gopath)-1] == '/' {
		gopath = gopath[:len(gopath)-1]
	}

	fi, err := os.Stat(gopath)
	if err != nil {
		return gopath, err
	}

	if !fi.IsDir() {
		return gopath, fmt.Errorf("%s no such directory", fi.Name())
	}

	return gopath, nil
}

func shellrc() (string, error) {
	var rc string
	u, err := user.Current()
	if err != nil {
		return rc, err
	}

	rc = u.HomeDir
	if sh := os.Getenv("SHELL"); !strings.Contains(sh, "zsh") {
		return rc, fmt.Errorf("%s is not supported", sh)
	}

	return rc + "/.zshrc", nil
}

func backup(src, dest string) error {
	sf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sf.Close()

	df, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer df.Close()

	if _, err = io.Copy(df, sf); err != nil {
		return err
	}

	if err := df.Sync(); err != nil {
		return err
	}

	return nil
}

func modify(path, gopath string) error {
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(bs), "\n")
	pos, err := position(lines)
	if err != nil {
		return err
	}

	lines[pos] = fmt.Sprintf("export GOPATH=%s", gopath)
	output := strings.Join(lines, "\n")

	if err := ioutil.WriteFile(path, []byte(output), 0644); err != nil {
		return err
	}

	return nil
}

func position(lines []string) (int, error) {
	for i, l := range lines {
		if strings.HasPrefix(l, "export GOPATH=") {
			return i, nil
		}
	}
	return -1, errors.New("GOPATH not found")
}

func terror(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(2)
}
