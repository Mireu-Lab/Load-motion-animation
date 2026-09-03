package trc_test

import (
	"fmt"
	"strings"

	"gitlab.com/mireu-lab/Load-motion-animation/pkg/trc"
)

func ExampleParser_ParseReader() {
	trcContent := `PathFileType	4	(X/Y/Z)	test.trc
DataRate	CameraRate	NumFrames	NumMarkers	Units	OrigDataRate	OrigDataStartFrame	OrigNumFrames
60.0	60.0	1	1	mm	60.0	1	1
Frame#	Time	Marker1		
			X1	Y1	Z1
1	0.000	10.0	20.0	30.0
`
	parser := trc.NewParser()
	data, err := parser.ParseReader(strings.NewReader(trcContent), "test.trc")
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	fmt.Printf("Format: %s, Frames: %d, Rate: %.1f\n", data.Format, data.TotalFrames, data.FrameRate)
	// Output: Format: TRC, Frames: 1, Rate: 60.0
}
