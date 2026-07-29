from typing import Dict, List, Any
from .parser import MotionData

def to_dict(motion_data: MotionData) -> Dict[str, List[Any]]:
    """Convert MotionData to flat columnar dictionary format"""
    data = {
        "frame": [],
        "time_sec": [],
        "joint_name": [],
        "pos_x": [],
        "pos_y": [],
        "pos_z": [],
        "rot_x": [],
        "rot_y": [],
        "rot_z": [],
    }

    for frame in motion_data.frames:
        for joint_name, pos in frame.joints.items():
            rot = frame.rotations.get(joint_name)
            data["frame"].append(frame.frame_index)
            data["time_sec"].append(frame.time_sec)
            data["joint_name"].append(joint_name)
            data["pos_x"].append(pos.x)
            data["pos_y"].append(pos.y)
            data["pos_z"].append(pos.z)
            data["rot_x"].append(rot.x if rot else 0.0)
            data["rot_y"].append(rot.y if rot else 0.0)
            data["rot_z"].append(rot.z if rot else 0.0)

    return data

def to_dataframe(motion_data: MotionData, library: str = "auto") -> Any:
    """Convert MotionData to Pandas or Polars DataFrame if library available, else dict"""
    flat_dict = to_dict(motion_data)

    if library in ("pandas", "auto"):
        try:
            import pandas as pd
            return pd.DataFrame(flat_dict)
        except ImportError:
            if library == "pandas":
                raise ImportError("Pandas library is not installed.")

    if library in ("polars", "auto"):
        try:
            import polars as pl
            return pl.DataFrame(flat_dict)
        except ImportError:
            if library == "polars":
                raise ImportError("Polars library is not installed.")

    return flat_dict
