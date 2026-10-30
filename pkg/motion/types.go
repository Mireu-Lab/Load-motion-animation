// Package motion provides unified in-memory data structures and interfaces for
// 3D and 2D motion capture and animation data.
//
// It defines the standard core types including [Vector3] for 3D spatial points,
// [Joint] for skeletal hierarchies, [FrameData] for time-sampled frame states,
// and [MotionData] as the canonical container produced by all format parsers
// (such as BVH, C3D, and TRC).
package motion

import (
	"fmt"
	"io"
)

// Vector3 represents a 3-dimensional Cartesian coordinate point or translation/rotation vector.
type Vector3 struct {
	// X coordinate / roll value
	X float64 `json:"x"`
	// Y coordinate / pitch value
	Y float64 `json:"y"`
	// Z coordinate / yaw value
	Z float64 `json:"z"`
}

// String returns a human-readable string representation of the 3D vector.
func (v Vector3) String() string {
	return fmt.Sprintf("(%.3f, %.3f, %.3f)", v.X, v.Y, v.Z)
}

// Joint represents a single joint node within an articulated skeletal hierarchy.
type Joint struct {
	// Name is the unique identifier of the joint (e.g. "Hips", "LeftArm").
	Name string `json:"name"`
	// Offset is the static relative 3D translation offset from its parent joint.
	Offset Vector3 `json:"offset"`
	// Channels defines the degree of freedom names for motion playback (e.g. "Xposition", "Zrotation").
	Channels []string `json:"channels"`
	// Children contains child joint nodes nested directly beneath this joint.
	Children []*Joint `json:"children,omitempty"`
	// ParentName is the name of the parent joint, or empty string if this is the root.
	ParentName string `json:"parent_name,omitempty"`
	// EndSite specifies the terminal offset if this joint has an end effector site.
	EndSite *Vector3 `json:"end_site,omitempty"`
}

// FrameData holds joint or marker positions, rotations, and raw channel values for a single motion frame.
type FrameData struct {
	// FrameIndex is the zero-based sequence index of the frame.
	FrameIndex int `json:"frame_index"`
	// TimeSec is the elapsed timestamp in seconds corresponding to this frame.
	TimeSec float64 `json:"time_sec"`
	// Joints maps each joint/marker name to its computed 3D world-space position.
	Joints map[string]Vector3 `json:"joints"`
	// Rotations maps each joint name to its Euler rotation angles (in degrees), if applicable.
	Rotations map[string]Vector3 `json:"rotations"`
	// RawValues holds the raw float channel values parsed directly from the source stream.
	RawValues []float64 `json:"raw_values"`
}

// MotionData is the unified in-memory container holding loaded motion capture data.
type MotionData struct {
	// Format is the source file format identifier (e.g., "BVH", "C3D", "TRC").
	Format string `json:"format"`
	// FileName is the original path or name of the motion file.
	FileName string `json:"file_name"`
	// RootJoint is the root joint of the skeleton hierarchy (for hierarchical formats like BVH).
	RootJoint *Joint `json:"root_joint,omitempty"`
	// JointList is the ordered list of all joint or marker names.
	JointList []string `json:"joint_list"`
	// FrameRate is the recording frequency in frames per second (FPS).
	FrameRate float64 `json:"frame_rate"`
	// TotalFrames is the total number of frames recorded in this motion clip.
	TotalFrames int `json:"total_frames"`
	// FrameDuration is the time interval between successive frames in seconds (1.0 / FrameRate).
	FrameDuration float64 `json:"frame_duration"`
	// Frames contains the sequential array of frame data samples.
	Frames []FrameData `json:"frames"`
	// MetaData contains format-specific metadata key-value attributes.
	MetaData map[string]string `json:"meta_data,omitempty"`
}

// Duration returns the total duration of the motion clip in seconds.
func (m *MotionData) Duration() float64 {
	if m.FrameRate > 0 {
		return float64(m.TotalFrames) / m.FrameRate
	}
	return float64(m.TotalFrames) * m.FrameDuration
}

// HasRotations reports whether the motion data contains joint rotation angles.
func (m *MotionData) HasRotations() bool {
	if len(m.Frames) == 0 {
		return false
	}
	return len(m.Frames[0].Rotations) > 0
}

// Parser defines the common interface implemented by all motion file format parsers.
type Parser interface {
	// ParseFile loads and decodes motion data from the given filesystem path.
	ParseFile(filePath string) (*MotionData, error)
}

// FrameCallback is invoked for each decoded frame in streaming mode.
// Returning a non-nil error will immediately abort the stream parsing.
type FrameCallback func(frameIndex int, timeSec float64, positions map[string]Vector3, rotations map[string]Vector3, root *Joint) error

// ReaderParser defines a parser capable of reading and parsing from an io.Reader.
type ReaderParser interface {
	Parser
	ParseReader(r io.Reader, filename string) (*MotionData, error)
}

// StreamParser defines a parser capable of streaming motion frames chunk-by-chunk or frame-by-frame.
type StreamParser interface {
	ReaderParser
	StreamReader(r io.Reader, filename string, callback FrameCallback) (*MotionData, error)
	StreamFile(filePath string, callback FrameCallback) (*MotionData, error)
}
