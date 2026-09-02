import csv
import json
from typing import Optional
from .parser import MotionData
from .dataframe import to_dict

def export_csv(motion_data: MotionData, output_path: str, delimiter: str = ","):
    """Export motion data to CSV or TSV file"""
    flat = to_dict(motion_data)
    cols = list(flat.keys())
    num_rows = len(flat["frame"])

    with open(output_path, "w", newline="", encoding="utf-8") as f:
        writer = csv.writer(f, delimiter=delimiter)
        writer.writerow(cols)
        for i in range(num_rows):
            row = [flat[col][i] for col in cols]
            writer.writerow(row)

def export_sql(motion_data: MotionData, output_path: str, db_type: str = "sqlite", table_name: str = "motion_frames"):
    """Export batch SQL INSERT statements for PostgreSQL, MySQL, SQLite"""
    flat = to_dict(motion_data)
    num_rows = len(flat["frame"])

    with open(output_path, "w", encoding="utf-8") as f:
        if db_type == "postgresql":
            f.write(f"CREATE TABLE IF NOT EXISTS {table_name} (id BIGSERIAL PRIMARY KEY, motion_file VARCHAR(255), frame_index INT, time_sec DOUBLE PRECISION, joint_name VARCHAR(100), pos_x DOUBLE PRECISION, pos_y DOUBLE PRECISION, pos_z DOUBLE PRECISION, rot_x DOUBLE PRECISION, rot_y DOUBLE PRECISION, rot_z DOUBLE PRECISION);\n")
        elif db_type == "mysql":
            f.write(f"CREATE TABLE IF NOT EXISTS {table_name} (id BIGINT AUTO_INCREMENT PRIMARY KEY, motion_file VARCHAR(255), frame_index INT, time_sec DOUBLE, joint_name VARCHAR(100), pos_x DOUBLE, pos_y DOUBLE, pos_z DOUBLE, rot_x DOUBLE, rot_y DOUBLE, rot_z DOUBLE);\n")
        else:
            f.write(f"CREATE TABLE IF NOT EXISTS {table_name} (id INTEGER PRIMARY KEY AUTOINCREMENT, motion_file TEXT, frame_index INTEGER, time_sec REAL, joint_name TEXT, pos_x REAL, pos_y REAL, pos_z REAL, rot_x REAL, rot_y REAL, rot_z REAL);\n")

        f.write("BEGIN TRANSACTION;\n")
        batch = []
        for i in range(num_rows):
            file_esc = motion_data.file_name.replace("'", "''")
            j_esc = str(flat["joint_name"][i]).replace("'", "''")
            tup = f"('{file_esc}', {flat['frame'][i]}, {flat['time_sec'][i]:.6f}, '{j_esc}', {flat['pos_x'][i]:.6f}, {flat['pos_y'][i]:.6f}, {flat['pos_z'][i]:.6f}, {flat['rot_x'][i]:.6f}, {flat['rot_y'][i]:.6f}, {flat['rot_z'][i]:.6f})"
            batch.append(tup)

            if len(batch) >= 500:
                f.write(f"INSERT INTO {table_name} (motion_file, frame_index, time_sec, joint_name, pos_x, pos_y, pos_z, rot_x, rot_y, rot_z) VALUES\n" + ",\n".join(batch) + ";\n")
                batch = []

        if batch:
            f.write(f"INSERT INTO {table_name} (motion_file, frame_index, time_sec, joint_name, pos_x, pos_y, pos_z, rot_x, rot_y, rot_z) VALUES\n" + ",\n".join(batch) + ";\n")

        f.write("COMMIT;\n")

def export_json(motion_data: MotionData, output_path: str):
    """Export motion data to JSON format"""
    doc = {
        "format": motion_data.format,
        "file_name": motion_data.file_name,
        "total_frames": motion_data.total_frames,
        "frame_rate": motion_data.frame_rate,
        "joint_list": motion_data.joint_list,
        "frames": [
            {
                "frame_index": f.frame_index,
                "time_sec": f.time_sec,
                "joints": {k: v.to_dict() for k, v in f.joints.items()},
                "rotations": {k: v.to_dict() for k, v in f.rotations.items()},
            }
            for f in motion_data.frames
        ]
    }
    with open(output_path, "w", encoding="utf-8") as f:
        json.dump(doc, f, indent=2)

def export_stream_csv(frame_iterator, output_path: str, delimiter: str = ","):
    """Stream motion frames directly to a CSV or TSV file without buffering in memory."""
    cols = ["frame", "time_sec", "joint_name", "pos_x", "pos_y", "pos_z", "rot_x", "rot_y", "rot_z"]
    with open(output_path, "w", newline="", encoding="utf-8") as f:
        writer = csv.writer(f, delimiter=delimiter)
        writer.writerow(cols)
        for frame in frame_iterator:
            for jname, pos in frame.joints.items():
                rot = frame.rotations.get(jname)
                writer.writerow([
                    frame.frame_index,
                    f"{frame.time_sec:.6f}",
                    jname,
                    f"{pos.x:.6f}",
                    f"{pos.y:.6f}",
                    f"{pos.z:.6f}",
                    f"{rot.x:.6f}" if rot else "0.000000",
                    f"{rot.y:.6f}" if rot else "0.000000",
                    f"{rot.z:.6f}" if rot else "0.000000",
                ])

def export_stream_sql(frame_iterator, output_path: str, db_type: str = "sqlite", table_name: str = "motion_frames", file_name: str = "stream"):
    """Stream batch SQL INSERT statements directly to a file."""
    with open(output_path, "w", encoding="utf-8") as f:
        if db_type == "postgresql":
            f.write(f"CREATE TABLE IF NOT EXISTS {table_name} (id BIGSERIAL PRIMARY KEY, motion_file VARCHAR(255), frame_index INT, time_sec DOUBLE PRECISION, joint_name VARCHAR(100), pos_x DOUBLE PRECISION, pos_y DOUBLE PRECISION, pos_z DOUBLE PRECISION, rot_x DOUBLE PRECISION, rot_y DOUBLE PRECISION, rot_z DOUBLE PRECISION);\n")
        elif db_type == "mysql":
            f.write(f"CREATE TABLE IF NOT EXISTS {table_name} (id BIGINT AUTO_INCREMENT PRIMARY KEY, motion_file VARCHAR(255), frame_index INT, time_sec DOUBLE, joint_name VARCHAR(100), pos_x DOUBLE, pos_y DOUBLE, pos_z DOUBLE, rot_x DOUBLE, rot_y DOUBLE, rot_z DOUBLE);\n")
        else:
            f.write(f"CREATE TABLE IF NOT EXISTS {table_name} (id INTEGER PRIMARY KEY AUTOINCREMENT, motion_file TEXT, frame_index INTEGER, time_sec REAL, joint_name TEXT, pos_x REAL, pos_y REAL, pos_z REAL, rot_x REAL, rot_y REAL, rot_z REAL);\n")

        f.write("BEGIN TRANSACTION;\n")
        batch = []
        file_esc = file_name.replace("'", "''")

        for frame in frame_iterator:
            for jname, pos in frame.joints.items():
                rot = frame.rotations.get(jname)
                rx = rot.x if rot else 0.0
                ry = rot.y if rot else 0.0
                rz = rot.z if rot else 0.0
                j_esc = str(jname).replace("'", "''")
                tup = f"('{file_esc}', {frame.frame_index}, {frame.time_sec:.6f}, '{j_esc}', {pos.x:.6f}, {pos.y:.6f}, {pos.z:.6f}, {rx:.6f}, {ry:.6f}, {rz:.6f})"
                batch.append(tup)

                if len(batch) >= 500:
                    f.write(f"INSERT INTO {table_name} (motion_file, frame_index, time_sec, joint_name, pos_x, pos_y, pos_z, rot_x, rot_y, rot_z) VALUES\n" + ",\n".join(batch) + ";\n")
                    batch = []

        if batch:
            f.write(f"INSERT INTO {table_name} (motion_file, frame_index, time_sec, joint_name, pos_x, pos_y, pos_z, rot_x, rot_y, rot_z) VALUES\n" + ",\n".join(batch) + ";\n")

        f.write("COMMIT;\n")

