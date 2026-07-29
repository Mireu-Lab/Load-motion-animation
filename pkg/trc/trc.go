package trc

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"load-motion-animation/pkg/motion"
)

// Parser parses TRC (Track Row Column) motion files
type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) ParseFile(filePath string) (*motion.MotionData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open TRC file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	motionData := &motion.MotionData{
		Format:   "TRC",
		FileName: filePath,
		MetaData: make(map[string]string),
	}

	var markerNames []string
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 2 {
			fields = strings.Fields(line)
		}

		if lineNum == 3 && len(fields) >= 3 {
			dataRate, _ := strconv.ParseFloat(fields[0], 64)
			numFrames, _ := strconv.Atoi(fields[2])
			motionData.FrameRate = dataRate
			motionData.TotalFrames = numFrames
			if dataRate > 0 {
				motionData.FrameDuration = 1.0 / dataRate
			}
		} else if lineNum == 4 { // Marker names line
			for i := 2; i < len(fields); i++ {
				mName := strings.TrimSpace(fields[i])
				if mName != "" {
					markerNames = append(markerNames, mName)
				}
			}
			motionData.JointList = markerNames
		} else if lineNum >= 6 { // Data lines
			if len(fields) < 2 {
				continue
			}

			frameIdx, _ := strconv.Atoi(fields[0])
			timeSec, _ := strconv.ParseFloat(fields[1], 64)

			frameJoints := make(map[string]motion.Vector3)
			rawVals := make([]float64, 0, len(fields)-2)

			valIdx := 2
			for _, mName := range markerNames {
				if valIdx+2 < len(fields) {
					x, _ := strconv.ParseFloat(fields[valIdx], 64)
					y, _ := strconv.ParseFloat(fields[valIdx+1], 64)
					z, _ := strconv.ParseFloat(fields[valIdx+2], 64)
					frameJoints[mName] = motion.Vector3{X: x, Y: y, Z: z}
					rawVals = append(rawVals, x, y, z)
					valIdx += 3
				}
			}

			frame := motion.FrameData{
				FrameIndex: frameIdx,
				TimeSec:    timeSec,
				Joints:     frameJoints,
				Rotations:  make(map[string]motion.Vector3),
				RawValues:  rawVals,
			}

			motionData.Frames = append(motionData.Frames, frame)
		}
	}

	return motionData, nil
}
