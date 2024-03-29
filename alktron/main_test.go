/*
Copyright 2019 The Arktos Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main_test

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"

	"github.com/containernetworking/cni/pkg/version"
	"github.com/onsi/gomega/gexec"
)

func TestVersion(t *testing.T) {
	pathToBin, err := gexec.Build("github.com/futurewei-cloud/cniplugins/alktron/")
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

func TestAddWithEmptyVPC(t *testing.T) {
	pathToBin, err := gexec.Build("github.com/futurewei-cloud/cniplugins/alktron/")
	if err != nil {
		t.Fatalf("failed to build binary: %v", err)
	}
	defer gexec.CleanupBuildArtifacts()

	cmd := exec.Command(pathToBin, "")
	cmd.Env = append(cmd.Env, "CNI_COMMAND=ADD")
	cmd.Env = append(cmd.Env, "CNI_ARGS=foo=bar") //invalid cni args for neutron integration which requires VPC, NICs
	cmd.Env = append(cmd.Env, "CNI_CONTAINERID=c")
	cmd.Env = append(cmd.Env, "CNI_NETNS=n")
	cmd.Env = append(cmd.Env, "CNI_IFNAME=ens01")
	cmd.Env = append(cmd.Env, "CNI_PATH=.")
	cmd.Stdin = strings.NewReader(`{"cniVersion": "0.3.1","name": "dbnet", "type": "alktron"}`)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	cmd.Run()

	exitCode := cmd.ProcessState.ExitCode()
	if exitCode != 1 {
		t.Errorf("expecting exit code 1; got %d", exitCode)
	}

	out := stdout.String()
	t.Logf("stdout: %s", out)
	if !strings.Contains(out, "empty VPC") {
		t.Errorf("stdout expecting 'empty VPC', got %q", out)
	}
}
