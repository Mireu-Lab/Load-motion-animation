import os
import subprocess
from typing import Optional
from .parser import MotionData

def export_gif(motion_data: MotionData, output_path: str, max_frames: int = 120, fps: int = 20) -> str:
    """Export motion data to animated GIF. Delegates to Go rendering engine for high performance."""
    # Check if Go binary exists
    go_bin = os.path.join(os.path.dirname(os.path.dirname(os.path.dirname(__file__))), "bin", "motion-parser")
    
    if os.path.exists(go_bin):
        cmd = [go_bin, "-input", motion_data.file_name, "-export", "gif", "-output", output_path, "-max-frames", str(max_frames)]
        result = subprocess.run(cmd, capture_output=True, text=True)
        if result.returncode == 0:
            return output_path
        else:
            raise RuntimeError(f"Go GIF renderer failed: {result.stderr}")
    else:
        # If binary not built yet, attempt go run or fallback notification
        cmd = ["go", "run", "./cmd/motion-parser", "-input", motion_data.file_name, "-export", "gif", "-output", output_path, "-max-frames", str(max_frames)]
        result = subprocess.run(cmd, capture_output=True, text=True)
        if result.returncode == 0:
            return output_path
        else:
            raise RuntimeError("GIF visualizer requires Go engine compiled at bin/motion-parser")
