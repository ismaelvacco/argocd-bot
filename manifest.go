package main

import (
	"gopkg.in/yaml.v3"
)

type Image struct {
	Name   string `yaml:"name"`
	NewTag string `yaml:"newTag"`
}

type Kustomize struct {
	ApiVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Resources  []string `yaml:"resources"`
	Images     []Image  `yaml:"images"`
}

type Manifest struct {
	FileContent []byte
	Kustomize   Kustomize
}

func NewManifest(fileContent []byte) *Manifest {
	return &Manifest{
		FileContent: fileContent,
	}
}

func (m *Manifest) Parse() error {
	var k Kustomize
	err := yaml.Unmarshal(m.FileContent, &k)
	if err != nil {
		return err
	}
	m.Kustomize = k
	return nil
}
