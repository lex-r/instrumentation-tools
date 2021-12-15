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

package term

import (
	"github.com/gdamore/tcell"
	"github.com/lex-r/promq/promq/term/plot"
)

// GraphView is a widget that displays the given graph with on the screen.  You
// *must* supply a domain and range labeler.  Default spacings will be chosen
// for the tick spacing if not specified.
type GraphView struct {
	pos PositionBox

	Graph *plot.PlatonicGraph

	DomainLabeler plot.DomainLabeler
	RangeLabeler  plot.RangeLabeler

	DomainTickSpacing int
	RangeTickSpacing  int
}

func (g *GraphView) SetBox(box PositionBox) {
	g.pos = box
}

func (g *GraphView) FlushTo(screen tcell.Screen) {
	if g.Graph == nil {
		return
	}

	// default to ticks 10 apart
	rangeSpacing := g.RangeTickSpacing
	if rangeSpacing == 0 {
		rangeSpacing = 10
	}
	domainSpacing := g.DomainTickSpacing
	if domainSpacing == 0 {
		domainSpacing = 10
	}

	// calculate the axes, to figure out the "inner" size (columns and rows available for the plot
	// itself once we taking drawing the axis labels & ticks into account)...
	screenSize := plot.ScreenSize{Cols: plot.Column(g.pos.Cols), Rows: plot.Row(g.pos.Rows)}
	scale := func(p float64) float64 { return p }
	axes := plot.EvenlySpacedTicks(g.Graph, screenSize, plot.TickScaling{
		RangeScale:    scale,
		DomainDensity: domainSpacing,
		RangeDensity:  rangeSpacing,
	}, plot.Labeling{
		DomainLabeler: g.DomainLabeler,
		RangeLabeler:  g.RangeLabeler,
		LineSize:      1,
	})

	if axes.InnerGraphSize.Cols == 0 || axes.InnerGraphSize.Rows == 0 {
		// too small to render, just bail
		return
	}

	plot.DrawAxes(axes, func(row plot.Row, col plot.Column, contents rune, kind plot.AxisCellKind) {
		switch kind {
		case plot.DomainTickKind:
			var sty tcell.Style
			screen.SetContent(int(col)+g.pos.StartCol, int(row)+g.pos.StartRow, '┯', nil, sty)
		case plot.RangeTickKind:
			var sty tcell.Style
			screen.SetContent(int(col)+g.pos.StartCol, int(row)+g.pos.StartRow, '┨', nil, sty)
		case plot.YAxisKind:
			var sty tcell.Style
			screen.SetContent(int(col)+g.pos.StartCol, int(row)+g.pos.StartRow, '┃', nil, sty)
		case plot.XAxisKind:
			var sty tcell.Style
			screen.SetContent(int(col)+g.pos.StartCol, int(row)+g.pos.StartRow, '━', nil, sty)
		case plot.AxisCornerKind:
			var sty tcell.Style
			screen.SetContent(int(col)+g.pos.StartCol, int(row)+g.pos.StartRow, '┗', nil, sty)
		case plot.LabelKind:
			var sty tcell.Style
			screen.SetContent(int(col)+g.pos.StartCol, int(row)+g.pos.StartRow, contents, nil, sty)
		}

	})

	// ... and use that "inner" size as the space for the plot itself.
	screenGraph := g.Graph.ToScreen(scale, plot.BrailleCellScreenSize(axes.InnerGraphSize))
	renderedGraph := screenGraph.Render(plot.BrailleCellMapper)

	startCol := g.pos.StartCol + int(axes.MarginCols)
	startRow := g.pos.StartRow
	plot.DrawBraille(renderedGraph, func(row plot.Row, col plot.Column, contents rune, id plot.SeriesId) {
		var sty tcell.Style
		if id != plot.NoSeries {
			sty = sty.Foreground(tcell.Color(id % 256))
		}
		screen.SetContent(int(col)+startCol, int(row)+startRow, contents, nil, sty)
	})
}
