/*
Copyright 2020 The Kubernetes Authors.

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

package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/lex-r/promq/cmd/cli"
	"github.com/lex-r/promq/cmd/metrics"
)

// PromQOptions provides information required to updat
type PromQOptions struct {
	args  []string
	flags cli.PromQFlags
	genericclioptions.IOStreams
}

// NewPromQOptions provides an instance of PQOptions
func NewPromQOptions(streams genericclioptions.IOStreams) *PromQOptions {
	return &PromQOptions{
		IOStreams: streams,
	}
}

type RootPromQCmd struct {
	*cobra.Command
	options *PromQOptions
}

func addFlags(cmd *cobra.Command, options *PromQOptions) {
	cmd.Flags().BoolVarP(&options.flags.Continuous, "continuous", "c", options.flags.Continuous, "if true, runs continuously (i.e. gathers samples in mem)")
	cmd.Flags().BoolVarP(&options.flags.List, "list", "l", options.flags.List, "if true, lists out observed metric names.")
	cmd.Flags().StringVarP(&options.flags.PromQuery, "query", "q", "", "if specified, uses this query for analyzing a prometheus endpoint.")
	cmd.Flags().StringVarP(&options.flags.Output, "output", "o", "json", "output format for data, defaults to json")
	cmd.Flags().StringArrayVarP(&options.flags.HostNames, "targets", "t", options.flags.HostNames, "prometheus metrics endpoint")
}

// NewCmdPromQ provides a cobra command wrapping AnalyzeOptions
func NewCmdPromQ(streams genericclioptions.IOStreams) *RootPromQCmd {
	o := NewPromQOptions(streams)
	cmd := &cobra.Command{
		Use: "promq [options]",
		Example: `
promq                                               # for interactive mode
promq -l                                            # to list metrics  
promq -q "apiserver_request_total" -ojson           # to query for all metrics matching the promql query in json
promq -q "apiserver_request_total" -oyaml           # to query for all metrics matching the promql query in yaml
`,
		SilenceUsage: true,

		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			ac, err := o.toPromQCmd()
			if err != nil {
				return err
			}
			metricCmd := metrics.MetricsCommand{
				PromQCommand: ac,
				// todo(han): let's hardcode this for now.
				//  I'm not terribly crazy about specifying this during each invocation
				Period: 1 * time.Second,
				Window: 1 * time.Minute,
			}
			if err := metricCmd.Run(o.flags); err != nil {
				return err
			}
			return nil
		},
	}
	promq := &RootPromQCmd{Command: cmd, options: o}

	addFlags(cmd, o)

	return promq
}

// Complete sets all information required for updating the current context
func (o *PromQOptions) Complete(cmd *cobra.Command, args []string) error {
	o.args = args
	return nil
}

// Validate ensures that all required arguments and flag values are provided
func (o *PromQOptions) Validate() error {
	if len(o.flags.HostNames) == 0 {
		return fmt.Errorf("specify at least one target")
	}

	return nil
}

func (o *PromQOptions) Run() error {
	return nil
}

func (o *PromQOptions) toPromQCmd() (cli.PromQCommand, error) {
	ac := cli.PromQCommand{Streams: o.IOStreams}
	return ac, nil
}
