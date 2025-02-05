/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package prom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/fatih/color"
	"github.com/golang/protobuf/proto"
	"github.com/hokaccha/go-prettyjson"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql"
	"gopkg.in/yaml.v2"
)

func ToPrettyJson(result *promql.Result) (*string, error) {
	if result.Err != nil {
		return nil, result.Err
	}
	if result.Warnings != nil {
		for _, e := range result.Warnings {
			fmt.Println(e)
		}
	}
	s, err := json.MarshalIndent(result.Value, "", "  ")
	if err != nil {
		return nil, err
	}
	return proto.String(string(s)), nil
}

func ToPrettyColoredJson(result *promql.Result) (*string, error) {
	if result.Err != nil {
		return nil, result.Err
	}
	if result.Warnings != nil {
		for _, e := range result.Warnings {
			fmt.Println(e)
		}
	}
	f := prettyjson.NewFormatter()
	f.Indent = 4
	f.KeyColor = color.New(color.FgGreen)
	f.NullColor = color.New(color.Underline)
	f.NumberColor = color.New(color.FgYellow)
	f.StringColor = color.New(color.FgHiCyan)
	f.BoolColor = nil

	s, err := f.Marshal(result.Value)
	if err != nil {
		return nil, err
	}
	return proto.String(string(s)), nil
}

func ToYaml(result *promql.Result) (*string, error) {
	if result.Err != nil {
		return nil, result.Err
	}
	if result.Warnings != nil {
		for _, e := range result.Warnings {
			fmt.Println(e)
		}
	}
	o, err := yaml.Marshal(result.Value)
	if err != nil {
		return nil, err
	}
	return proto.String(string(o)), nil
}

func ToPlain(result *promql.Result) (*string, error) {
	if result.Err != nil {
		return nil, result.Err
	}
	if result.Warnings != nil {
		for _, e := range result.Warnings {
			fmt.Println(e)
		}
	}
	o, err := PlainMarshal(result.Value)
	if err != nil {
		return nil, err
	}

	return proto.String(string(o)), nil
}

func ToPrettyFormat(res *promql.Result, outputType string, colorized bool) (*string, error) {
	switch outputType {
	case "plain":
		o, err := ToPlain(res)
		if err != nil {
			return nil, err
		}
		return o, nil

	case "json":
		var o *string
		var err error
		if colorized {
			o, err = ToPrettyColoredJson(res)
		} else {
			o, err = ToPrettyJson(res)
		}
		if err != nil {
			return nil, err
		}
		return o, nil

	case "yaml":
		o, err := ToYaml(res)
		if err != nil {
			return nil, err
		}
		return o, nil
	}
	return nil, fmt.Errorf("unsupported formatting option (%s)", outputType)
}

func PlainMarshal(value promql.Value) ([]byte, error) {
	switch val := value.(type) {
	case promql.Vector:
		var buf bytes.Buffer
		for i, v := range val {
			if v.Metric.Has(labels.MetricName) {
				buf.WriteString(v.Metric.Get(labels.MetricName))
				buf.WriteString("")
			}
			lbs := v.Metric.MatchLabels(false, labels.MetricName)
			if len(lbs) > 0 {
				buf.WriteString(lbs.String())
				buf.WriteString(" ")
			}
			vv := strconv.FormatFloat(v.V, 'f', -1, 64)
			buf.WriteString(fmt.Sprintf("%v", vv))
			if i < len(val)-1 {
				buf.WriteString("\n")
			}
		}
		return buf.Bytes(), nil
	default:
		return nil, fmt.Errorf("unsupported value type %T", val)
	}
}
