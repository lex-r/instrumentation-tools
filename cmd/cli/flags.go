/*
Copyright 2019 The Kubernetes Authors.

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

package cli

import (
	"fmt"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type PromQCommand struct {
	Streams genericclioptions.IOStreams
}

type PromQFlags struct {
	List       bool
	PromQuery  string
	Output     string
	HostNames  []string
	Continuous bool
}

type PQableCommand interface {
	Run(bool) error
	Fprintf(string, ...interface{})
}

func (c *PromQCommand) Fprintf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(c.Streams.Out, format, args...)
}
