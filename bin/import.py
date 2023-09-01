#!/usr/bin/env python
import csv
import datetime
import json
import os
import subprocess
import sys
from typing import Any
from urllib.parse import urlparse
from zoneinfo import ZoneInfo

UTC = ZoneInfo("UTC")
TIMEZONE = ZoneInfo("Europe/London")

url_normalisations: dict[str, Any] = {
    "add_www": {
        "theguardian.com",
    },
    "remove_www": {
        "www.jacobin.com",
        "www.jacobinmag.com",
        "www.tribunemag.co.uk",
    },
    "replace_domain": {
        "jacobinmag.com": "jacobin.com",
    },
    "force_https": {
        "www.theguardian.com",
        "jacobin.com",
        "tribunemag.co.uk",
        "newsocialist.org.uk",
    },
}


def add_to_database(items: list[dict[str, Any]]) -> None:
    env = os.environ.get("ENVIRONMENT", "development")
    cmd = ["bookmarks", "add"]
    if env != "production":
        cmd = ["go", "run", ".", "add"]

    process = subprocess.Popen(cmd, stdin=subprocess.PIPE, text=True)
    process.communicate(json.dumps(items))
    if process.returncode:
        sys.exit(process.returncode)


def normalise_url(url: str) -> str:
    parsed = urlparse(url)
    if parsed.netloc in url_normalisations["add_www"]:
        parsed = parsed._replace(netloc="www." + parsed.netloc)
    if parsed.netloc in url_normalisations["remove_www"]:
        parsed = parsed._replace(netloc=parsed.netloc.removeprefix("www."))
    if parsed.netloc in url_normalisations["replace_domain"]:
        parsed = parsed._replace(
            netloc=url_normalisations["replace_domain"][parsed.netloc]
        )

    if parsed.netloc == "medium.com" or parsed.netloc.endswith(".medium.com"):
        # special case
        parsed = parsed._replace(netloc="scribe.rip")

    if parsed.netloc in url_normalisations["force_https"]:
        parsed = parsed._replace(scheme="https")

    return parsed.geturl()


def import_dropbox_csv(csvpath: str, read: bool) -> None:
    key_name = "ReadAt" if read else "SavedAt"

    with open(csvpath, newline="") as csvfile:
        reader = csv.DictReader(csvfile, fieldnames=("URL", "Timestamp"))
        items = [
            {
                "URL": normalise_url(item["URL"]),
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
                "URL": normalise_url(item["URL"]),
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
                "URL": normalise_url(item["href"]),
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
    csvpath, *_ = sys.argv[1:]
    import_instapaper_csv(csvpath)
    import_dropbox_csv(
        "/Users/ben/Library/CloudStorage/Dropbox/IFTTT/Instapaper/Saved Items.txt",
        False,
    )
    import_dropbox_csv(
        "/Users/ben/Library/CloudStorage/Dropbox/IFTTT/Instapaper/Archived Items.txt",
        True,
    )
