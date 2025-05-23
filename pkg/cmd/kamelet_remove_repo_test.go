/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1"
)

const cmdKameletRemoveRepo = "remove-repo"

// nolint: unparam
func initializeKameletRemoveRepoCmdOptions(t *testing.T) (*kameletRemoveRepoCommandOptions, *cobra.Command, RootCmdOptions) {
	t.Helper()

	options, rootCmd := kamelTestPreAddCommandInit()
	kameletRemoveRepoCommandOptions := addTestKameletRemoveRepoCmd(*options, rootCmd)
	kamelTestPostAddCommandInit(t, rootCmd, options)

	return kameletRemoveRepoCommandOptions, rootCmd, *options
}

func addTestKameletRemoveRepoCmd(options RootCmdOptions, rootCmd *cobra.Command) *kameletRemoveRepoCommandOptions {
	// Add a testing version of kamelet remove-repo Command
	kameletRemoveRepoCmd, kameletRemoveRepoOptions := newKameletRemoveRepoCmd(&options)
	kameletRemoveRepoCmd.RunE = func(c *cobra.Command, args []string) error {
		return nil
	}
	kameletRemoveRepoCmd.PostRunE = func(c *cobra.Command, args []string) error {
		return nil
	}
	kameletRemoveRepoCmd.Args = ArbitraryArgs
	rootCmd.AddCommand(kameletRemoveRepoCmd)
	return kameletRemoveRepoOptions
}

func TestKameletRemoveRepoNoFlag(t *testing.T) {
	_, rootCmd, _ := initializeKameletRemoveRepoCmdOptions(t)
	_, err := ExecuteCommand(rootCmd, cmdKameletRemoveRepo, "foo")
	require.NoError(t, err)
}

func TestKameletRemoveRepoNonExistingFlag(t *testing.T) {
	_, rootCmd, _ := initializeKameletRemoveRepoCmdOptions(t)
	_, err := ExecuteCommand(rootCmd, cmdKameletRemoveRepo, "--nonExistingFlag", "foo")
	require.Error(t, err)
}

func TestKameletRemoveRepoURINotFoundEmpty(t *testing.T) {
	repositories := []v1.KameletRepositorySpec{}
	_, err := getURIIndex("foo", repositories)
	require.Error(t, err)
}

func TestKameletRemoveRepoURINotFoundNotEmpty(t *testing.T) {
	repositories := []v1.KameletRepositorySpec{{URI: "github:foo/bar"}}
	_, err := getURIIndex("foo", repositories)
	require.Error(t, err)
}

func TestKameletRemoveRepoURIFound(t *testing.T) {
	repositories := []v1.KameletRepositorySpec{{URI: "github:foo/bar1"}, {URI: "github:foo/bar2"}, {URI: "github:foo/bar3"}}
	i, err := getURIIndex("github:foo/bar2", repositories)
	require.NoError(t, err)
	assert.Equal(t, 1, i)
}
