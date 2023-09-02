#!/usr/bin/env python
import csv
import datetime
import json
import os
import subprocess
import sys
from typing import Any
from zoneinfo import ZoneInfo

UTC = ZoneInfo("UTC")
TIMEZONE = ZoneInfo("Europe/London")


def add_to_database(items: list[dict[str, Any]]) -> None:
    env = os.environ.get("ENVIRONMENT", "development")
    cmd = ["bookmarks", "add"]
    if env != "production":
        cmd = ["go", "run", ".", "add"]

    process = subprocess.Popen(cmd, stdin=subprocess.PIPE, text=True)
    process.communicate(json.dumps(items))
    if process.returncode:
        sys.exit(process.returncode)


def import_dropbox_csv(csvpath: str, read: bool) -> None:
    key_name = "ReadAt" if read else "SavedAt"

    with open(csvpath, newline="") as csvfile:
        reader = csv.DictReader(csvfile, fieldnames=("URL", "Timestamp"))
        items = [
            {
                "URL": item["URL"],
                key_name: (
                    datetime.datetime.strptime(
                        item["Timestamp"], "%B %d, %Y at %I:%M%p"
                    )
                    .astimezone(TIMEZONE)
                    .isoformat()
                ),
            }
            for item in reader
        ]

        add_to_database(items)


def import_instapaper_csv(csvpath: str) -> None:
    tags = {"instapaper"}

    with open(csvpath, newline="") as csvfile:
        reader = csv.DictReader(csvfile)
        items = [
            {
                "URL": item["URL"],
                "Title": item["Title"],
                "Description": item["Selection"],
                "SavedAt": datetime.datetime.fromtimestamp(int(item["Timestamp"]))
                .astimezone(UTC)
                .isoformat(),
                "ReadAt": 0 if item["Folder"] == "Archive" else None,
                "Tags": "{"
                + (",".join((tags | {item["Folder"].lower()}) - {"archive", "unread"}))
                + "}",
            }
            for item in reader
        ]

        add_to_database(items)


def import_json(jsonpath: str) -> None:
    tags = {"pinboard"}

    data = json.load(open(jsonpath))
    items = [
        (
            {
                "URL": item["href"],
                "Title": item["description"],
                "Description": item["extended"],
                "SavedAt": item["time"],
                "ReadAt": 0 if item["toread"] == "no" else None,
                "Tags": "{" + (",".join(tags | set(item["tags"].split(" ")))) + "}",
            }
        )
        for item in data
    ]

    add_to_database(items)


if __name__ == "__main__":
    args = zip(sys.argv[1::2], sys.argv[2::2])
    for arg, path in args:
        print(arg, path, file=sys.stderr)
        if arg == "--instapaper-csv":
            import_instapaper_csv(path)
        elif arg == "--pinboard-json":
            import_json(path)
        elif arg == "--saved-csv":
            import_dropbox_csv(path, False)
        elif arg == "--read-csv":
            import_dropbox_csv(path, True)
