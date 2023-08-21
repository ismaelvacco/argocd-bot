package main

import (
	"bytes"
	"fmt"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/spf13/viper"
	ssh2 "golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"
	"io"
	"net"
	"os"
	"time"
)

var (
	gitUsername  = "git"
	kustomizeEnv = "development"
)

func main() {
	viper.AutomaticEnv()
	privateKey := viper.GetString("git_private_key")
	if privateKey == "" {
		panic("git_private_key is empty")
	}
	repository := viper.GetString("git_repository")
	if repository == "" {
		panic("git_repository is empty")
	}

	imagePath := viper.GetString("image_path")
	if imagePath == "" {
		panic("image_path is empty")
	}

	imageVersion := viper.GetString("image_version")
	if imageVersion == "" {
		panic("image_version is empty")
	}

	authorName := viper.GetString("author_name")
	if authorName == "" {
		panic("author_name is empty")
	}

	authorEmail := viper.GetString("author_email")
	if authorEmail == "" {
		panic("author_email is empty")
	}

	branchName := viper.GetString("branch_name")
	if branchName == "" {
		panic("branch_name is empty")
	}

	publicKey, err := ssh.NewPublicKeys(gitUsername, []byte(privateKey), "")
	if err != nil {
		panic(err)
	}

	publicKey.HostKeyCallback = func(hostname string, remote net.Addr, key ssh2.PublicKey) error {
		return nil
	}

	fs := memfs.New()
	r, err := git.Clone(memory.NewStorage(), fs, &git.CloneOptions{
		RemoteName:    "origin",
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branchName)),
		Auth:          publicKey,
		URL:           viper.GetString("git_repository"),
		Progress:      os.Stdout,
	})

	if err != nil {
		panic(err)
	}

	manifestPath := fmt.Sprintf("environments/%s/kustomization.yaml", kustomizeEnv)
	file, err := fs.Open(manifestPath)
	if err != nil {
		panic(err)
	}

	buf := bytes.NewBuffer(nil)
	io.Copy(buf, file)
	file.Close()
	manifest := NewManifest(buf.Bytes())
	err = manifest.Parse()
	if err != nil {
		panic(err)
	}

	newImage := Image{
		Name:   imagePath,
		NewTag: imageVersion,
	}

	existImage := false
	for i, image := range manifest.Kustomize.Images {
		if image.Name == newImage.Name {
			existImage = true
			manifest.Kustomize.Images[i].NewTag = newImage.NewTag
			break
		}
	}

	if !existImage {
		manifest.Kustomize.Images = append(manifest.Kustomize.Images, newImage)
	}

	manifestContent, err := yaml.Marshal(manifest.Kustomize)
	if err != nil {
		panic(err)
	}

	w, err := r.Worktree()
	if err != nil {
		panic(err)
	}

	updateFile, err := fs.Create(manifestPath)
	if err != nil {
		panic(err)
	}

	_, err = updateFile.Write(manifestContent)
	if err != nil {
		panic(err)
	}

	_, err = w.Add(manifestPath)

	_, err = w.Status()
	if err != nil {
		panic(err)
	}

	message := fmt.Sprintf("Update image %s to version %s", imagePath, imageVersion)
	commit, err := w.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  authorName,
			Email: authorEmail,
			When:  time.Now(),
		},
		All:               true,
		AllowEmptyCommits: false,
	})

	if err != nil {
		panic(err)
	}

	_, err = w.Status()
	if err != nil {
		panic(err)
	}

	_, err = r.CommitObject(commit)
	err = r.Push(&git.PushOptions{
		Auth: publicKey,
	})
	if err != nil {
		panic(err)
	}

}
