package main_test

import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/containernetworking/cni/pkg/version"
	"github.com/onsi/gomega/gexec"
)

func TestVersion(t *testing.T) {
	pathToBin, err := gexec.Build("github.com/futurewei-cloud/neubernetes")
	if err != nil {
		t.Fatalf("failed to build binary: %v", err)
	}
	defer gexec.CleanupBuildArtifacts()

	cmd := exec.Command(pathToBin, "")
	cmd.Env = append(cmd.Env, "CNI_COMMAND=VERSION")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err = cmd.Run()
	if err != nil {
		t.Fatalf("failed to run binary: %v", err)
	}

	dec := version.PluginDecoder{}
	pluginInfo, err := dec.Decode(stdout.Bytes())
	if err != nil {
		t.Fatalf("failed to parse version output: %v", err)
	}

	SupportedVersions := pluginInfo.SupportedVersions()
	for _, v := range SupportedVersions {
		if v == "0.3.1" {
			return
		}
	}

	t.Fatalf("expecting '0.3.1', got %q", SupportedVersions)
}
