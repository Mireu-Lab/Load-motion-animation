// Package c3d provides a binary parser for C3D (Coordinate 3D) motion capture files.
//
// C3D is a standardized binary file format commonly used in biomechanics, gait analysis,
// and optical motion capture systems (such as Vicon, Qualisys, and Motion Analysis).
// This package parses 3D point markers, frame rates, point labels, and trajectories
// into unified [motion.MotionData].
package c3d

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	"load-motion-animation/pkg/motion"
)

// Parser parses binary C3D motion capture files.
type Parser struct{}

var _ motion.StreamParser = (*Parser)(nil)

// NewParser creates a new C3D [Parser] instance.
func NewParser() *Parser {
	return &Parser{}
}

// C3DHeader represents the parsed header block (512 bytes) of a C3D binary file.
type C3DHeader struct {
	// ParamStartBlock is the 1-based block number where the parameter section starts.
	ParamStartBlock uint8
	// Magic is the signature byte (80 for standard C3D).
	Magic uint8
	// NumPoints is the number of 3D marker points per frame.
	NumPoints uint16
	// NumAnalog is the total number of analog measurements per 3D frame.
	NumAnalog uint16
	// StartFrame is the index of the first frame.
	StartFrame uint16
	// EndFrame is the index of the last frame.
	EndFrame uint16
	// MaxGap is the maximum interpolation gap length.
	MaxGap uint16
	// ScaleFactor converts integer coordinates to real units (negative indicates float storage).
	ScaleFactor float32
	// DataStartBlock is the 1-based block number where 3D/analog data begins.
	DataStartBlock uint16
	// AnalogRate is the number of analog samples per 3D frame.
	AnalogRate uint16
	// FrameRate is the 3D sampling rate in frames per second (Hz).
	FrameRate float32
}

// ParseFile reads and parses a C3D file from the filesystem into a complete [motion.MotionData] structure.
func (p *Parser) ParseFile(filePath string) (*motion.MotionData, error) {
	return p.StreamFile(filePath, nil)
}

// StreamFile opens the given C3D file and streams each decoded frame to callback without accumulating frames in memory.
func (p *Parser) StreamFile(filePath string, callback motion.FrameCallback) (*motion.MotionData, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read C3D file: %w", err)
	}
	return p.StreamBytes(data, filePath, callback)
}

// ParseReader reads all binary content from an [io.Reader] and decodes it.
func (p *Parser) ParseReader(r io.Reader, filename string) (*motion.MotionData, error) {
	return p.StreamReader(r, filename, nil)
}

// StreamReader decodes C3D frames from an [io.Reader] stream and calls callback for each frame if provided.
func (p *Parser) StreamReader(r io.Reader, filename string, callback motion.FrameCallback) (*motion.MotionData, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read C3D stream: %w", err)
	}
	return p.StreamBytes(data, filename, callback)
}

// ParseBytes parses C3D binary data from an in-memory byte slice.
func (p *Parser) ParseBytes(data []byte, filename string) (*motion.MotionData, error) {
	return p.StreamBytes(data, filename, nil)
}

// StreamBytes decodes C3D binary data and passes each frame to callback if non-nil.
func (p *Parser) StreamBytes(data []byte, filename string, callback motion.FrameCallback) (*motion.MotionData, error) {

	if len(data) < 512 {
		return nil, fmt.Errorf("file too small for C3D header")
	}

	header := C3DHeader{
		ParamStartBlock: data[0],
		Magic:           data[1],
		NumPoints:       binary.LittleEndian.Uint16(data[2:4]),
		NumAnalog:       binary.LittleEndian.Uint16(data[4:6]),
		StartFrame:      binary.LittleEndian.Uint16(data[6:8]),
		EndFrame:        binary.LittleEndian.Uint16(data[8:10]),
		MaxGap:          binary.LittleEndian.Uint16(data[10:12]),
		ScaleFactor:     math.Float32frombits(binary.LittleEndian.Uint32(data[12:16])),
		DataStartBlock:  binary.LittleEndian.Uint16(data[16:18]),
		AnalogRate:      binary.LittleEndian.Uint16(data[18:20]),
		FrameRate:       math.Float32frombits(binary.LittleEndian.Uint32(data[20:24])),
	}

	if header.Magic != 80 {
		return nil, fmt.Errorf("invalid C3D magic byte: expected 80, got %d", header.Magic)
	}

	totalFrames := int(header.EndFrame - header.StartFrame + 1)
	if totalFrames <= 0 {
		totalFrames = 0
	}

	// Read Point Labels from parameter section
	pointNames := parseC3DLabels(data, header.ParamStartBlock, header.NumPoints)

	motionData := &motion.MotionData{
		Format:      "C3D",
		FileName:    filename,
		JointList:   pointNames,
		FrameRate:   float64(header.FrameRate),
		TotalFrames: totalFrames,
		Frames:      make([]motion.FrameData, 0, totalFrames),
		MetaData:    make(map[string]string),
	}

	if header.FrameRate > 0 {
		motionData.FrameDuration = 1.0 / float64(header.FrameRate)
	}

	// Parse 3D data frames
	dataOffset := int(header.DataStartBlock-1) * 512
	if dataOffset < 512 || dataOffset >= len(data) {
		dataOffset = 512 * 2 // Fallback
	}

	scale := header.ScaleFactor
	isFloat := scale < 0

	bytesPerPoint := 8 // 4 * int16
	if isFloat {
		bytesPerPoint = 16 // 4 * float32
	}

	frameSize := int(header.NumPoints)*bytesPerPoint + int(header.NumAnalog)*int(header.AnalogRate)*2
	if isFloat {
		frameSize = int(header.NumPoints)*bytesPerPoint + int(header.NumAnalog)*int(header.AnalogRate)*4
	}

	if frameSize <= 0 {
		return motionData, nil
	}

	currPos := dataOffset
	for f := 0; f < totalFrames && (currPos+frameSize) <= len(data); f++ {
		frameJoints := make(map[string]motion.Vector3)
		rawVals := make([]float64, 0, header.NumPoints*3)

		for i := 0; i < int(header.NumPoints); i++ {
			ptOffset := currPos + i*bytesPerPoint
			var x, y, z float64

			if isFloat {
				x = float64(math.Float32frombits(binary.LittleEndian.Uint32(data[ptOffset : ptOffset+4])))
				y = float64(math.Float32frombits(binary.LittleEndian.Uint32(data[ptOffset+4 : ptOffset+8])))
				z = float64(math.Float32frombits(binary.LittleEndian.Uint32(data[ptOffset+8 : ptOffset+12])))
			} else {
				ix := int16(binary.LittleEndian.Uint16(data[ptOffset : ptOffset+2]))
				iy := int16(binary.LittleEndian.Uint16(data[ptOffset+2 : ptOffset+4]))
				iz := int16(binary.LittleEndian.Uint16(data[ptOffset+4 : ptOffset+6]))
				sc := math.Abs(float64(scale))
				x = float64(ix) * sc
				y = float64(iy) * sc
				z = float64(iz) * sc
			}

			ptName := pointNames[i]
			frameJoints[ptName] = motion.Vector3{X: x, Y: y, Z: z}
			rawVals = append(rawVals, x, y, z)
		}

		timeSec := float64(f) * motionData.FrameDuration
		if callback != nil {
			if err := callback(f, timeSec, frameJoints, make(map[string]motion.Vector3), nil); err != nil {
				return nil, err
			}
		} else {
			frame := motion.FrameData{
				FrameIndex: f,
				TimeSec:    timeSec,
				Joints:     frameJoints,
				Rotations:  make(map[string]motion.Vector3),
				RawValues:  rawVals,
			}
			motionData.Frames = append(motionData.Frames, frame)
		}
		currPos += frameSize
	}

	return motionData, nil
}

func parseC3DLabels(data []byte, paramStart uint8, numPoints uint16) []string {
	labels := make([]string, numPoints)
	for i := range labels {
		labels[i] = fmt.Sprintf("Point_%d", i+1)
	}

	paramOffset := int(paramStart-1) * 512
	if paramOffset >= len(data) {
		return labels
	}

	// Scan parameter section for LABELS string
	idx := bytes.Index(data[paramOffset:], []byte("LABELS"))
	if idx == -1 {
		return labels
	}

	labelsPos := paramOffset + idx + 6
	if labelsPos+6 >= len(data) {
		return labels
	}

	// Skip param description offset (2 bytes)
	// Read dimension info: num_dims, dim1 (str len), dim2 (num items)
	// In C3D specs: [offset_to_next (2), num_dims (1), dim1 (1), dim2 (2 or 1)]
	pos := labelsPos
	// Find first byte with string length (e.g. 8)
	for pos < labelsPos+20 && pos < len(data) {
		if data[pos] > 0 && data[pos] <= 64 { // plausible string length
			strLen := int(data[pos])
			if pos+2 < len(data) {
				count := int(binary.LittleEndian.Uint16(data[pos+1 : pos+3]))
				if count == int(numPoints) || (count > 0 && count <= 500) {
					startStr := pos + 3
					for i := 0; i < count && i < int(numPoints); i++ {
						sOffset := startStr + i*strLen
						if sOffset+strLen <= len(data) {
							lbl := strings.TrimSpace(string(data[sOffset : sOffset+strLen]))
							lbl = strings.Map(func(r rune) rune {
								if r >= 32 && r <= 126 {
									return r
								}
								return -1
							}, lbl)
							if lbl != "" {
								labels[i] = lbl
							}
						}
					}
					return labels
				}
			}
		}
		pos++
	}

	return labels
}
