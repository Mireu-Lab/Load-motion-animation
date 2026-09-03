// Package trc provides a parser for Track Row Column (.trc) optical motion capture data files.
//
// TRC is an ASCII tab-delimited format widely used in biomechanics research and OpenSim
// musculoskeletal simulations to record 3D trajectories of reflective optical markers over time.
package trc

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"gitlab.com/mireu-lab/Load-motion-animation/pkg/motion"
)

// Parser parses Track Row Column (.trc) optical motion files.
type Parser struct{}

var _ motion.StreamParser = (*Parser)(nil)

// NewParser creates and returns a new TRC [Parser] instance.
func NewParser() *Parser {
	return &Parser{}
}

// ParseFile opens the specified TRC file and parses its header metadata and marker trajectories.
func (p *Parser) ParseFile(filePath string) (*motion.MotionData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open TRC file: %w", err)
	}
	defer file.Close()
	return p.ParseReader(file, filePath)
}

// StreamFile opens the specified TRC file and streams each frame to the provided callback without buffering all frames in memory.
func (p *Parser) StreamFile(filePath string, callback motion.FrameCallback) (*motion.MotionData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open TRC file: %w", err)
	}
	defer file.Close()
	return p.StreamReader(file, filePath, callback)
}

// ParseReader decodes TRC marker coordinates and timestamps from an [io.Reader] stream into a complete [motion.MotionData] instance.
func (p *Parser) ParseReader(r io.Reader, filename string) (*motion.MotionData, error) {
	return p.StreamReader(r, filename, nil)
}

// StreamReader decodes TRC marker coordinates and timestamps from an [io.Reader] stream.
// If callback is non-nil, each frame is passed to the callback without buffering all frames into memory.
func (p *Parser) StreamReader(r io.Reader, filename string, callback motion.FrameCallback) (*motion.MotionData, error) {
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	motionData := &motion.MotionData{
		Format:   "TRC",
		FileName: filename,
		MetaData: make(map[string]string),
	}

	var markerNames []string
	lineNum := 0
	frameCount := 0

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

			if callback != nil {
				if err := callback(frameIdx, timeSec, frameJoints, make(map[string]motion.Vector3), nil); err != nil {
					return nil, err
				}
			} else {
				frame := motion.FrameData{
					FrameIndex: frameIdx,
					TimeSec:    timeSec,
					Joints:     frameJoints,
					Rotations:  make(map[string]motion.Vector3),
					RawValues:  rawVals,
				}
				motionData.Frames = append(motionData.Frames, frame)
			}
			frameCount++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading TRC data: %w", err)
	}

	if motionData.TotalFrames == 0 {
		motionData.TotalFrames = frameCount
	}

	return motionData, nil
}
