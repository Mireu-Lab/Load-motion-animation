package motion

import "fmt"

// Vector3 represents a 3D point or vector
type Vector3 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

func (v Vector3) String() string {
	return fmt.Sprintf("(%.3f, %.3f, %.3f)", v.X, v.Y, v.Z)
}

// Joint represents a joint node in a skeleton hierarchy
type Joint struct {
	Name       string    `json:"name"`
	Offset     Vector3   `json:"offset"`
	Channels   []string  `json:"channels"`
	Children   []*Joint  `json:"children,omitempty"`
	ParentName string    `json:"parent_name,omitempty"`
	EndSite    *Vector3  `json:"end_site,omitempty"`
}

// FrameData holds joint or marker positions/rotations for a single motion frame
type FrameData struct {
	FrameIndex int                 `json:"frame_index"`
	TimeSec    float64             `json:"time_sec"`
	Joints     map[string]Vector3  `json:"joints"`     // Joint positions (X, Y, Z)
	Rotations  map[string]Vector3  `json:"rotations"`  // Joint rotations (RX, RY, RZ) if applicable
	RawValues  []float64           `json:"raw_values"` // Raw frame channel data
}

// MotionData is the unified in-memory representation of loaded motion capture data
type MotionData struct {
	Format       string            `json:"format"`       // e.g. "BVH", "C3D", "TRC"
	FileName     string            `json:"file_name"`
	RootJoint    *Joint            `json:"root_joint,omitempty"`
	JointList    []string          `json:"joint_list"`
	FrameRate    float64           `json:"frame_rate"`
	TotalFrames  int               `json:"total_frames"`
	FrameDuration float64          `json:"frame_duration"`
	Frames       []FrameData       `json:"frames"`
	MetaData     map[string]string `json:"meta_data,omitempty"`
}

// Parser defines the common interface for all motion file parsers
type Parser interface {
	ParseFile(filePath string) (*MotionData, error)
}
