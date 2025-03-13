package gui

import (
	"fmt"
	qt "github.com/mappu/miqt/qt6"
	"strconv"
)

var (
	latencyFont = qt.NewQFont()
)

func init() {
	latencyFont.SetBold(true)
	latencyFont.SetPointSizeF(7.5)
}

func LatencyPainter(pixmap *qt.QPixmap, latency uint16) {
	var text string
	if latency == 0 {
		text = "-1"
	} else if latency < 1000 {
		text = strconv.FormatUint(uint64(latency), 10)
	} else {
		text = strconv.FormatFloat(float64(latency)/1000, 'f', 0, 64) + "k"
	}

	pixmap.Fill1(qt.NewQColor2(qt.Transparent))
	painter := qt.NewQPainter()
	color := qt.NewQColor3(0, 200, 0) // green
	if latency > 500 {
		color = qt.NewQColor3(255, 200, 0) // yellow
	}
	if latency > 1500 {
		color = qt.NewQColor3(255, 0, 0) // red
	}
	painter.Begin(pixmap.QPaintDevice)
	painter.SetPenWithPen(qt.NewQPen2(qt.NoPen))
	painter.SetBrush(qt.NewQBrush3(color))
	painter.SetRenderHint(qt.QPainter__Antialiasing | qt.QPainter__TextAntialiasing)
	painter.SetFont(latencyFont)
	painter.DrawRect2(0, 0, 22, 22)
	painter.SetPen(qt.NewQColor2(qt.Black))
	painter.DrawText6(pixmap.Rect(), int(qt.AlignCenter), text)
	painter.End()
}

func LatencyText(name string, latency uint16) string {
	return fmt.Sprintf("%s\t%dms", name, latency)
}
