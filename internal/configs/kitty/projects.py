#!/usr/bin/env python3
"""
kitty project switcher script
Cmd+P in kitty to open a project from ~/Work, ~/code, ~/oss
"""

import os
import subprocess
import sys
from pathlib import Path


def find_projects():
    """Find all project directories."""
    projects = []

    base_dirs = [
        Path.home() / "Work",
        Path.home() / "code",
        Path.home() / "oss",
    ]

    for base_dir in base_dirs:
        if not base_dir.exists():
            continue

        for item in base_dir.iterdir():
            if item.is_dir() and not item.name.startswith("."):
                projects.append(item)

    return sorted(projects, key=lambda p: p.name)


def select_project(projects):
    """Use fzf to select a project."""
    project_lines = "\n".join(str(p) for p in projects)

    try:
        result = subprocess.run(
            ["fzf", "--preview", "ls -la {}"],
            input=project_lines,
            text=True,
            capture_output=True,
        )

        if result.returncode == 0:
            return result.stdout.strip()
    except FileNotFoundError:
        print("fzf not found", file=sys.stderr)

    return None


def main():
    projects = find_projects()

    if not projects:
        print("No projects found", file=sys.stderr)
        sys.exit(1)

    selected = select_project(projects)

    if selected:
        print(selected)


if __name__ == "__main__":
    main()
