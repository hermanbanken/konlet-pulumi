package internal

import (
	"gopkg.in/yaml.v2"
)

type KonletDeclContainer struct {
	Image           string                `yaml:"image"`
	Name            string                `yaml:"name"`
	Command         []string              `yaml:"command"`
	Args            []string              `yaml:"args"`
	SecurityContext KonletSecurityContext `yaml:"securityContext"`
	Stdin           bool                  `yaml:"stdin"`
	Tty             bool                  `yaml:"tty"`
	VolumeMounts    []struct{}            `yaml:"volumeMounts"`
	Env             []KonletDeclEnv `yaml:"env"`
}

type KonletDeclEnv struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type KonletSecurityContext struct {
	Privileged bool `yaml:"privileged"`
}

type KonletDeclSpec struct {
	Containers    []KonletDeclContainer `yaml:"containers"`
	RestartPolicy string                `yaml:"restartPolicy"`
	Volumes       []struct{}            `yaml:"volumes"`
}

type KonletDeclaration struct {
	Spec KonletDeclSpec `yaml:"spec"`
}

// https://github.com/GoogleCloudPlatform/konlet/blob/master/gce-containers-startup/types/api.go#L77
func ContainerDeclaration(decl KonletDeclaration) string {
	const comment = `# DISCLAIMER:
# This container declaration format is not a public API and may change without
# notice. Please use gcloud command-line tool or Google Cloud Console to run
# Containers on Google Compute Engine.`

	data, err := yaml.Marshal(decl)
	if err != nil {
		panic(err)
	}

	return comment + "\n\n" + string(data)
}
