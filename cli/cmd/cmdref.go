// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var cmdRefDir string

var cmdRef = &cobra.Command{
	Use:   "cmdref",
	Short: "Generate kjournal command reference",
	RunE: func(cmd *cobra.Command, args []string) error {
		return genCmdRef()
	},
	Hidden: true,
}

func genCmdRef() error {
	rootCmd.DisableAutoGenTag = true
	return doc.GenMarkdownTreeCustom(rootCmd, cmdRefDir, filePrepend, linkHandler)
}

func linkHandler(s string) string {
	return s
}

func filePrepend(s string) string {
	// Prepend a HTML comment that this file is autogenerated. So that
	// users are warned before fixing issues in the Markdown files.  Should
	// never show up on the web.
	return fmt.Sprintf("%s\n\n", "<!-- This file was autogenerated via cilium cmdref, do not edit manually-->")
}

func init() {
	cmdRef.Flags().StringVarP(&cmdRefDir, "directory", "d", "./", "Path to the output directory")
	rootCmd.AddCommand(cmdRef)
}
