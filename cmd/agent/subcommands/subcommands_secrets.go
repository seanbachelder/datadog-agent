// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build secrets
// +build secrets

package subcommands

import (
	"github.com/DataDog/datadog-agent/cmd/agent/command"
	cmdsecrethelper "github.com/DataDog/datadog-agent/cmd/agent/subcommands/secrethelper"
)

// secretsSubcommands returns SubcommandFactories for subcommands dependent on the `secrets` build tag.
func secretsSubcommands() []command.SubcommandFactory {
	return []command.SubcommandFactory{
		cmdsecrethelper.Commands,
	}
}
