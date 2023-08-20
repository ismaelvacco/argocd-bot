package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

var (
	gitRepository  = "git@github.com:zaxapp/gitops.git"
	privateKeyPath = "/Users/ismael.vacco/.ssh/id_rsa"
	gitUsername    = "git"
	gitDirectory   = "tmp"
	kustomizeEnv   = "development"
)

func main() {

	defer func() {
		err := os.RemoveAll(gitDirectory)
		if err != nil {
			panic(err)
		}
	}()
	publicKey, err := ssh.NewPublicKeysFromFile(gitUsername, privateKeyPath, "")
	if err != nil {
		panic(err)
	}

	_, err = os.Stat(privateKeyPath)
	if err != nil {
		panic(err)
	}

	_, err = git.PlainClone(gitDirectory, false, &git.CloneOptions{
		RemoteName:    "origin",
		ReferenceName: "refs/heads/develop",
		Auth:          publicKey,
		URL:           gitRepository,
		Progress:      os.Stdout,
	})

	if err != nil {
		panic(err)
	}

	manifestPath := fmt.Sprintf("%s/environments/%s/kustomization.yaml", gitDirectory, kustomizeEnv)
	fileContent, err := os.ReadFile(manifestPath)
	if err != nil {
		panic(err)
	}

	manifest := NewManifest(fileContent)
	err = manifest.Parse()
	if err != nil {
		panic(err)
	}

	newImage := Image{
		Name:   "ghcr.io/zaxapp/magento/php-fpm",
		NewTag: "dev-0f90a701148c901536390d58dc425ae31ea4cd8f",
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

	r, err := git.PlainOpen(gitDirectory)
	if err != nil {
		panic(err)
	}

	w, err := r.Worktree()
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(manifestPath, manifestContent, 0644)
	if err != nil {
		panic(err)
	}

	_, err = w.Add(manifestPath)

	status, err := w.Status()
	if err != nil {
		panic(err)
	}
	fmt.Println(status)

	commit, err := w.Commit("example go-git commit in branch", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Ismael Vacco",
			Email: "isma.vacco@gmail.com",
			When:  time.Now(),
		},
		All:               true,
		AllowEmptyCommits: false,
	})

	if err != nil {
		panic(err)
	}

	status, err = w.Status()
	if err != nil {
		panic(err)
	}
	fmt.Println(status)

	obj, err := r.CommitObject(commit)
	err = r.Push(&git.PushOptions{
		//RemoteName: "origin",
		//RefSpecs:   []config.RefSpec{"refs/heads/develop:refs/heads/develop"},
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(obj)
}
