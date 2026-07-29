from setuptools import setup, find_packages

setup(
    name="load-motion-animation",
    version="0.1.0",
    description="Unified Motion Capture Parser & Data Pipeline Engine (Python & Golang)",
    author="Mireu-Lab",
    packages=find_packages(),
    python_requires=">=3.8",
    entry_points={
        "console_scripts": [
            "load-motion=load_motion.cli:main",
        ],
    },
)
