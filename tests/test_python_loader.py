import unittest
import os
import sys
import tempfile

sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "python")))

from load_motion import (
    load_motion,
    iter_motion_frames,
    load_motion_stream,
    TRCParser,
    to_dataframe,
    export_csv,
    export_sql,
    export_json,
    export_stream_csv,
    export_stream_sql,
)

class TestPythonLoader(unittest.TestCase):
    def setUp(self):
        self.sample_bvh = os.path.join(os.path.dirname(__file__), "..", "_sample", "Sample_motion.bvh")
        self.sample_c3d = os.path.join(os.path.dirname(__file__), "..", "_sample", "arthuman-sample.c3d")

    def test_bvh_loading(self):
        if not os.path.exists(self.sample_bvh):
            self.skipTest("Sample BVH file missing")
        data = load_motion(self.sample_bvh)
        self.assertEqual(data.format, "BVH")
        self.assertGreater(data.total_frames, 0)
        self.assertGreater(len(data.joint_list), 0)

    def test_bvh_streaming(self):
        if not os.path.exists(self.sample_bvh):
            self.skipTest("Sample BVH file missing")
        count = 0
        for frame in iter_motion_frames(self.sample_bvh):
            count += 1
            if count >= 10:
                break
        self.assertEqual(count, 10)

    def test_c3d_loading(self):
        if not os.path.exists(self.sample_c3d):
            self.skipTest("Sample C3D file missing")
        data = load_motion(self.sample_c3d)
        self.assertEqual(data.format, "C3D")
        self.assertGreater(data.total_frames, 0)

    def test_trc_parsing(self):
        sample_trc = (
            "PathFileType\t4\t(X/Y/Z)\ttest.trc\n"
            "DataRate\tCameraRate\tNumFrames\tNumMarkers\tUnits\tOrigDataRate\tOrigDataStartFrame\tOrigNumFrames\n"
            "60.0\t60.0\t2\t2\tmm\t60.0\t1\t2\n"
            "Frame#\tTime\tMarker1\t\t\tMarker2\t\t\n"
            "\t\tX1\tY1\tZ1\tX2\tY2\tZ2\n"
            "1\t0.016667\t1.0\t2.0\t3.0\t4.0\t5.0\t6.0\n"
            "2\t0.033333\t1.1\t2.1\t3.1\t4.1\t5.1\t6.1\n"
        )
        with tempfile.NamedTemporaryFile(mode="w", suffix=".trc", delete=False) as f:
            f.write(sample_trc)
            trc_path = f.name

        try:
            data = load_motion(trc_path)
            self.assertEqual(data.format, "TRC")
            self.assertEqual(data.total_frames, 2)
            self.assertEqual(len(data.joint_list), 2)
            self.assertEqual(len(data.frames), 2)
        finally:
            os.unlink(trc_path)

    def test_exporters(self):
        if not os.path.exists(self.sample_bvh):
            self.skipTest("Sample BVH file missing")
        data = load_motion(self.sample_bvh)
        with tempfile.TemporaryDirectory() as tmp_dir:
            csv_file = os.path.join(tmp_dir, "out.csv")
            sql_file = os.path.join(tmp_dir, "out.sql")
            json_file = os.path.join(tmp_dir, "out.json")

            export_csv(data, csv_file)
            self.assertTrue(os.path.exists(csv_file))

            export_sql(data, sql_file, db_type="postgresql")
            self.assertTrue(os.path.exists(sql_file))

            export_json(data, json_file)
            self.assertTrue(os.path.exists(json_file))

    def test_streaming_exporters(self):
        if not os.path.exists(self.sample_bvh):
            self.skipTest("Sample BVH file missing")
        with tempfile.TemporaryDirectory() as tmp_dir:
            csv_file = os.path.join(tmp_dir, "stream_out.csv")
            sql_file = os.path.join(tmp_dir, "stream_out.sql")

            # Stream first 20 frames
            def first_20():
                for idx, frame in enumerate(iter_motion_frames(self.sample_bvh)):
                    if idx >= 20:
                        break
                    yield frame

            export_stream_csv(first_20(), csv_file)
            self.assertTrue(os.path.exists(csv_file))
            self.assertGreater(os.path.getsize(csv_file), 0)

            export_stream_sql(first_20(), sql_file, db_type="sqlite")
            self.assertTrue(os.path.exists(sql_file))
            self.assertGreater(os.path.getsize(sql_file), 0)

if __name__ == "__main__":
    unittest.main()
