import os
import struct
import math
from typing import Dict, List, Any, Optional, Iterator, Tuple

class Vector3:
    __slots__ = ('x', 'y', 'z')
    def __init__(self, x: float = 0.0, y: float = 0.0, z: float = 0.0):
        self.x = float(x)
        self.y = float(y)
        self.z = float(z)

    def to_dict(self) -> Dict[str, float]:
        return {"x": self.x, "y": self.y, "z": self.z}

    def __repr__(self):
        return f"Vector3({self.x:.3f}, {self.y:.3f}, {self.z:.3f})"

class Joint:
    __slots__ = ('name', 'parent_name', 'offset', 'channels', 'children')
    def __init__(self, name: str, parent_name: Optional[str] = None):
        self.name = name
        self.parent_name = parent_name
        self.offset = Vector3()
        self.channels: List[str] = []
        self.children: List['Joint'] = []

class FrameData:
    __slots__ = ('frame_index', 'time_sec', 'joints', 'rotations', 'raw_values')
    def __init__(self, frame_index: int, time_sec: float):
        self.frame_index = frame_index
        self.time_sec = time_sec
        self.joints: Dict[str, Vector3] = {}
        self.rotations: Dict[str, Vector3] = {}
        self.raw_values: List[float] = []

class MotionData:
    def __init__(self, format_name: str, file_name: str):
        self.format = format_name
        self.file_name = file_name
        self.root_joint: Optional[Joint] = None
        self.joint_list: List[str] = []
        self.frame_rate: float = 0.0
        self.total_frames: int = 0
        self.frame_duration: float = 0.0
        self.frames: List[FrameData] = []
        self.meta_data: Dict[str, str] = {}

class BVHParser:
    def parse(self, file_path: str) -> MotionData:
        motion_data = MotionData("BVH", file_path)
        with open(file_path, "r", encoding="utf-8", errors="ignore") as f:
            lines = [line.strip() for line in f.readlines() if line.strip()]

        joint_stack: List[Joint] = []
        channel_map = []  # list of (joint_name, channel_type)
        in_hierarchy = False
        in_motion = False

        for line in lines:
            tokens = line.split()
            if not tokens:
                continue

            tag = tokens[0]
            if tag == "HIERARCHY":
                in_hierarchy = True
                continue
            elif tag == "MOTION":
                in_hierarchy = False
                in_motion = True
                continue

            if in_hierarchy:
                if tag in ("ROOT", "JOINT"):
                    name = tokens[1]
                    joint = Joint(name)
                    if joint_stack:
                        joint.parent_name = joint_stack[-1].name
                        joint_stack[-1].children.append(joint)
                    else:
                        motion_data.root_joint = joint
                    joint_stack.append(joint)
                    motion_data.joint_list.append(name)
                elif tag == "End":
                    parent = joint_stack[-1]
                    joint = Joint(f"{parent.name}_End", parent_name=parent.name)
                    parent.children.append(joint)
                    joint_stack.append(joint)
                elif tag == "}":
                    if joint_stack:
                        joint_stack.pop()
                elif tag == "OFFSET":
                    if len(tokens) >= 4 and joint_stack:
                        curr = joint_stack[-1]
                        curr.offset = Vector3(float(tokens[1]), float(tokens[2]), float(tokens[3]))
                elif tag == "CHANNELS":
                    if len(tokens) >= 2 and joint_stack:
                        num_ch = int(tokens[1])
                        curr = joint_stack[-1]
                        for i in range(num_ch):
                            ch = tokens[2 + i]
                            curr.channels.append(ch)
                            channel_map.append((curr.name, ch))

            elif in_motion:
                if line.startswith("Frames:"):
                    motion_data.total_frames = int(tokens[1])
                elif line.startswith("Frame Time:"):
                    dur = float(tokens[2])
                    motion_data.frame_duration = dur
                    if dur > 0:
                        motion_data.frame_rate = 1.0 / dur
                else:
                    vals = [float(x) for x in tokens]
                    frame_idx = len(motion_data.frames)
                    frame = FrameData(frame_idx, frame_idx * motion_data.frame_duration)
                    frame.raw_values = vals

                    joint_vals: Dict[str, Dict[str, float]] = {}
                    for idx, val in enumerate(vals):
                        if idx < len(channel_map):
                            jname, chtype = channel_map[idx]
                            if jname not in joint_vals:
                                joint_vals[jname] = {}
                            joint_vals[jname][chtype] = val

                    # Fast Forward Kinematics
                    identity_mat = (1.0,0.0,0.0,0.0, 0.0,1.0,0.0,0.0, 0.0,0.0,1.0,0.0, 0.0,0.0,0.0,1.0)
                    self._compute_fk_fast(motion_data.root_joint, identity_mat, joint_vals, frame.joints, frame.rotations)
                    motion_data.frames.append(frame)

        return motion_data

    def iter_frames(self, file_path: str) -> Iterator[FrameData]:
        """Stream frames one-by-one from BVH file using a generator."""
        with open(file_path, "r", encoding="utf-8", errors="ignore") as f:
            joint_stack: List[Joint] = []
            channel_map = []
            root_joint = None
            frame_duration = 0.0
            in_hierarchy = False
            in_motion = False
            frame_idx = 0

            for line in f:
                line = line.strip()
                if not line:
                    continue
                tokens = line.split()
                if not tokens:
                    continue

                tag = tokens[0]
                if tag == "HIERARCHY":
                    in_hierarchy = True
                    continue
                elif tag == "MOTION":
                    in_hierarchy = False
                    in_motion = True
                    continue

                if in_hierarchy:
                    if tag in ("ROOT", "JOINT"):
                        name = tokens[1]
                        joint = Joint(name)
                        if joint_stack:
                            joint.parent_name = joint_stack[-1].name
                            joint_stack[-1].children.append(joint)
                        else:
                            root_joint = joint
                        joint_stack.append(joint)
                    elif tag == "End":
                        parent = joint_stack[-1]
                        joint = Joint(f"{parent.name}_End", parent_name=parent.name)
                        parent.children.append(joint)
                        joint_stack.append(joint)
                    elif tag == "}":
                        if joint_stack:
                            joint_stack.pop()
                    elif tag == "OFFSET":
                        if len(tokens) >= 4 and joint_stack:
                            curr = joint_stack[-1]
                            curr.offset = Vector3(float(tokens[1]), float(tokens[2]), float(tokens[3]))
                    elif tag == "CHANNELS":
                        if len(tokens) >= 2 and joint_stack:
                            num_ch = int(tokens[1])
                            curr = joint_stack[-1]
                            for i in range(num_ch):
                                ch = tokens[2 + i]
                                curr.channels.append(ch)
                                channel_map.append((curr.name, ch))

                elif in_motion:
                    if line.startswith("Frames:"):
                        pass
                    elif line.startswith("Frame Time:"):
                        dur = float(tokens[2])
                        frame_duration = dur
                    else:
                        vals = [float(x) for x in tokens]
                        frame = FrameData(frame_idx, frame_idx * frame_duration)
                        frame.raw_values = vals

                        joint_vals: Dict[str, Dict[str, float]] = {}
                        for idx, val in enumerate(vals):
                            if idx < len(channel_map):
                                jname, chtype = channel_map[idx]
                                if jname not in joint_vals:
                                    joint_vals[jname] = {}
                                joint_vals[jname][chtype] = val

                        identity_mat = (1.0, 0.0, 0.0, 0.0, 0.0, 1.0, 0.0, 0.0, 0.0, 0.0, 1.0, 0.0, 0.0, 0.0, 0.0, 1.0)
                        self._compute_fk_fast(root_joint, identity_mat, joint_vals, frame.joints, frame.rotations)
                        yield frame
                        frame_idx += 1


    def _compute_fk_fast(self, joint: Optional[Joint], M: tuple, joint_vals: Dict[str, Dict[str, float]], positions: Dict[str, Vector3], rotations: Dict[str, Vector3]):
        if not joint:
            return

        j_vals = joint_vals.get(joint.name, {})
        tx = joint.offset.x + j_vals.get("Xposition", 0.0)
        ty = joint.offset.y + j_vals.get("Yposition", 0.0)
        tz = joint.offset.z + j_vals.get("Zposition", 0.0)

        rx = j_vals.get("Xrotation", 0.0)
        ry = j_vals.get("Yrotation", 0.0)
        rz = j_vals.get("Zrotation", 0.0)
        rotations[joint.name] = Vector3(rx, ry, rz)

        # Fast matrix multiplication for flat 16-tuple
        # Translation matrix T
        m0,m1,m2,m3, m4,m5,m6,m7, m8,m9,m10,m11, m12,m13,m14,m15 = M

        # Local T
        # T * R
        r0,r1,r2, r4,r5,r6, r8,r9,r10 = 1.0,0.0,0.0, 0.0,1.0,0.0, 0.0,0.0,1.0

        for ch in joint.channels:
            if ch == "Xrotation":
                rad = math.radians(rx)
                c, s = math.cos(rad), math.sin(rad)
                # R = R * Rx
                r1, r2 = r1*c + r2*s, -r1*s + r2*c
                r5, r6 = r5*c + r6*s, -r5*s + r6*c
                r9, r10 = r9*c + r10*s, -r9*s + r10*c
            elif ch == "Yrotation":
                rad = math.radians(ry)
                c, s = math.cos(rad), math.sin(rad)
                r0, r2 = r0*c - r2*s, r0*s + r2*c
                r4, r6 = r4*c - r6*s, r4*s + r6*c
                r8, r10 = r8*c - r10*s, r8*s + r10*c
            elif ch == "Zrotation":
                rad = math.radians(rz)
                c, s = math.cos(rad), math.sin(rad)
                r0, r1 = r0*c + r1*s, -r0*s + r1*c
                r4, r5 = r4*c + r5*s, -r4*s + r5*c
                r8, r9 = r8*c + r9*s, -r8*s + r9*c

        # Global matrix G = M * L
        # L has top-left (r0..r10), right column (tx, ty, tz)
        g0 = m0*r0 + m1*r4 + m2*r8
        g1 = m0*r1 + m1*r5 + m2*r9
        g2 = m0*r2 + m1*r6 + m2*r10
        g3 = m0*tx + m1*ty + m2*tz + m3

        g4 = m4*r0 + m5*r4 + m6*r8
        g5 = m4*r1 + m5*r5 + m6*r9
        g6 = m4*r2 + m5*r6 + m6*r10
        g7 = m4*tx + m5*ty + m6*tz + m7

        g8 = m8*r0 + m9*r4 + m10*r8
        g9 = m8*r1 + m9*r5 + m10*r9
        g10 = m8*r2 + m9*r6 + m10*r10
        g11 = m8*tx + m9*ty + m10*tz + m11

        G = (g0,g1,g2,g3, g4,g5,g6,g7, g8,g9,g10,g11, 0.0,0.0,0.0,1.0)
        positions[joint.name] = Vector3(g3, g7, g11)

        for child in joint.children:
            self._compute_fk_fast(child, G, joint_vals, positions, rotations)

class C3DParser:
    def parse(self, file_path: str) -> MotionData:
        motion_data = MotionData("C3D", file_path)
        with open(file_path, "rb") as f:
            data = f.read()

        if len(data) < 512:
            raise ValueError("File too small for C3D header")

        param_block, magic = data[0], data[1]
        if magic != 80:
            raise ValueError(f"Invalid C3D magic byte: {magic}")

        num_points = struct.unpack("<H", data[2:4])[0]
        num_analog = struct.unpack("<H", data[4:6])[0]
        start_frame = struct.unpack("<H", data[6:8])[0]
        end_frame = struct.unpack("<H", data[8:10])[0]
        scale_factor = struct.unpack("<f", data[12:16])[0]
        data_block = struct.unpack("<H", data[16:18])[0]
        analog_rate = struct.unpack("<H", data[18:20])[0]
        frame_rate = struct.unpack("<f", data[20:24])[0]

        total_frames = max(0, end_frame - start_frame + 1)
        motion_data.total_frames = total_frames
        motion_data.frame_rate = float(frame_rate)
        if frame_rate > 0:
            motion_data.frame_duration = 1.0 / frame_rate

        point_names = [f"Point_{i+1}" for i in range(num_points)]
        motion_data.joint_list = point_names

        is_float = scale_factor < 0
        scale = abs(scale_factor) if not is_float else 1.0
        bytes_per_point = 16 if is_float else 8
        frame_size = num_points * bytes_per_point + num_analog * analog_rate * (4 if is_float else 2)

        data_offset = (data_block - 1) * 512
        if data_offset < 512 or data_offset >= len(data):
            data_offset = 512 * 2

        curr_pos = data_offset
        for f in range(total_frames):
            if curr_pos + frame_size > len(data):
                break

            frame = FrameData(f, f * motion_data.frame_duration)
            for i in range(num_points):
                pt_offset = curr_pos + i * bytes_per_point
                if is_float:
                    x, y, z = struct.unpack("<fff", data[pt_offset:pt_offset+12])
                else:
                    ix, iy, iz = struct.unpack("<hhh", data[pt_offset:pt_offset+6])
                    x, y, z = ix * scale, iy * scale, iz * scale
                
                name = point_names[i]
                frame.joints[name] = Vector3(x, y, z)
                frame.raw_values.extend([x, y, z])

            motion_data.frames.append(frame)
            curr_pos += frame_size

        return motion_data

    def iter_frames(self, file_path: str) -> Iterator[FrameData]:
        """Stream frames one-by-one from C3D file."""
        motion_data = self.parse(file_path)
        for frame in motion_data.frames:
            yield frame

class TRCParser:
    def parse(self, file_path: str) -> MotionData:
        motion_data = MotionData("TRC", file_path)
        with open(file_path, "r", encoding="utf-8", errors="ignore") as f:
            lines = [line.strip() for line in f.readlines() if line.strip()]

        marker_names = []
        for line_num, line in enumerate(lines, 1):
            fields = line.split("\t") if "\t" in line else line.split()
            if line_num == 3 and len(fields) >= 3:
                data_rate = float(fields[0])
                num_frames = int(fields[2])
                motion_data.frame_rate = data_rate
                motion_data.total_frames = num_frames
                if data_rate > 0:
                    motion_data.frame_duration = 1.0 / data_rate
            elif line_num == 4:
                marker_names = [m.strip() for m in fields[2:] if m.strip()]
                motion_data.joint_list = marker_names
            elif line_num >= 6:
                if len(fields) < 2:
                    continue
                frame_idx = int(fields[0])
                time_sec = float(fields[1])
                frame = FrameData(frame_idx, time_sec)
                val_idx = 2
                for m_name in marker_names:
                    if val_idx + 2 < len(fields):
                        x, y, z = float(fields[val_idx]), float(fields[val_idx+1]), float(fields[val_idx+2])
                        frame.joints[m_name] = Vector3(x, y, z)
                        frame.raw_values.extend([x, y, z])
                        val_idx += 3
                motion_data.frames.append(frame)

        return motion_data

    def iter_frames(self, file_path: str) -> Iterator[FrameData]:
        """Stream frames one-by-one from TRC file."""
        with open(file_path, "r", encoding="utf-8", errors="ignore") as f:
            marker_names = []
            frame_duration = 0.0
            for line_num, line in enumerate(f, 1):
                line = line.strip()
                if not line:
                    continue
                fields = line.split("\t") if "\t" in line else line.split()
                if line_num == 3 and len(fields) >= 3:
                    data_rate = float(fields[0])
                    if data_rate > 0:
                        frame_duration = 1.0 / data_rate
                elif line_num == 4:
                    marker_names = [m.strip() for m in fields[2:] if m.strip()]
                elif line_num >= 6:
                    if len(fields) < 2:
                        continue
                    frame_idx = int(fields[0])
                    time_sec = float(fields[1])
                    frame = FrameData(frame_idx, time_sec)
                    val_idx = 2
                    for m_name in marker_names:
                        if val_idx + 2 < len(fields):
                            x, y, z = float(fields[val_idx]), float(fields[val_idx+1]), float(fields[val_idx+2])
                            frame.joints[m_name] = Vector3(x, y, z)
                            frame.raw_values.extend([x, y, z])
                            val_idx += 3
                    yield frame

def load_motion(file_path: str) -> MotionData:
    """Load complete motion dataset into memory from a file (.bvh, .c3d, .trc)."""
    ext = os.path.splitext(file_path)[1].lower()
    if ext == ".bvh":
        return BVHParser().parse(file_path)
    elif ext == ".c3d":
        return C3DParser().parse(file_path)
    elif ext == ".trc":
        return TRCParser().parse(file_path)
    else:
        raise ValueError(f"Unsupported file format: {ext}")

def iter_motion_frames(file_path: str) -> Iterator[FrameData]:
    """Stream motion frames one-by-one as a generator for low-memory processing."""
    ext = os.path.splitext(file_path)[1].lower()
    if ext == ".bvh":
        return BVHParser().iter_frames(file_path)
    elif ext == ".c3d":
        return C3DParser().iter_frames(file_path)
    elif ext == ".trc":
        return TRCParser().iter_frames(file_path)
    else:
        raise ValueError(f"Unsupported file format: {ext}")

def load_motion_stream(file_path: str) -> Tuple[MotionData, Iterator[FrameData]]:
    """Load metadata container and return an iterator yielding each frame."""
    ext = os.path.splitext(file_path)[1].lower()
    if ext == ".bvh":
        p = BVHParser()
        data = p.parse(file_path)
        data.frames = []  # Clear buffered frames for low memory
        return data, p.iter_frames(file_path)
    elif ext == ".c3d":
        p = C3DParser()
        data = p.parse(file_path)
        frames = data.frames
        data.frames = []
        return data, iter(frames)
    elif ext == ".trc":
        p = TRCParser()
        data = p.parse(file_path)
        data.frames = []
        return data, p.iter_frames(file_path)
    else:
        raise ValueError(f"Unsupported file format: {ext}")

