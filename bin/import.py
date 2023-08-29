#!/usr/bin/env python
import csv
import datetime
import json
import sqlite3
import sys
from typing import Any
from urllib.parse import urlparse

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
    },
}


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


def create_or_update(
    db: sqlite3.Connection,
    url: str,
    tags: set[str],
    title: str,
    description: str,
    read_at: int | datetime.datetime | None,
    saved_at: datetime.datetime,
    now: datetime.datetime,
) -> None:
    cursor = db.cursor()
    entry = dict(
        cursor.execute("SELECT * FROM links WHERE url=?", (url,)).fetchone() or {}
    )

    if entry:
        changed = False

        for tag in entry["tags"].lower().strip("{}").split(","):
            if tag == "instapaper":
                if "pinboard" in tags:
                    tags.remove("pinboard")
                continue
            tags.add(tag)
        for tag in tags:
            if tag not in entry["tags"].lower().strip("{}").split(","):
                changed = True
                entry["tags"] = "{" + (",".join(tags)) + "}"
                break

        if entry["title"] != title:
            entry["title"] = title
            changed = True
        if entry["description"] != description:
            entry["description"] = description
            changed = True
        if str(entry["read_at"]) != str(read_at):
            entry["read_at"] = read_at
            changed = True
        if str(entry["saved_at"]) != str(saved_at):
            entry["saved_at"] = saved_at
            changed = True

        if changed:
            entry["updated_at"] = now

            cursor.execute(
                """UPDATE links SET
                        url=:url,
                        created_at=:created_at,
                        updated_at=:updated_at,
                        saved_at=:saved_at,
                        read_at=:read_at,
                        title=:title,
                        description=:description,
                        tags=:tags
                        WHERE url=:url;""",
                entry,
            )
    else:
        entry = {
            "url": url,
            "title": title,
            "description": description,
            "saved_at": saved_at,
            "read_at": read_at,
            "tags": "{" + (",".join(tags)) + "}",
            "created_at": now,
            "updated_at": now,
        }
        cursor.execute(
            """INSERT INTO links
                    (url, created_at, updated_at, saved_at, read_at, title, description, tags)
                    VALUES(:url, :created_at, :updated_at, :saved_at, :read_at, :title, :description, :tags);""",
            entry,
        )

    db.commit()


def import_dropbox_csv(csvpath: str, dbpath: str, read: bool) -> None:
    db = sqlite3.connect(dbpath)
    cursor = db.cursor()
    now = datetime.datetime.now()

    with open(csvpath, newline="") as csvpath:
        reader = csv.DictReader(csvpath, fieldnames=("URL", "Timestamp"))
        for item in reader:
            link = {
                "url": normalise_url(item["URL"]),
                "date": datetime.datetime.strptime(
                    item["Timestamp"].removesuffix("AM").removesuffix("PM"),
                    "%B %d, %Y at %H:%M",
                ),
                "now": now,
            }

            if read:
                cursor.execute(
                    """UPDATE links SET updated_at=:now, read_at=:date
                    WHERE url=:url AND read_at <> :date;""",
                    link,
                )
                cursor.execute(
                    """INSERT OR IGNORE INTO links
                    (url, created_at, updated_at, saved_at, read_at)
                    VALUES(:url, :now, :now, :now, :date);""",
                    link,
                )
            else:
                cursor.execute(
                    """UPDATE links SET updated_at=:now, saved_at=:date
                    WHERE url=:url AND saved_at <> :date;""",
                    link,
                )
                cursor.execute(
                    """INSERT OR IGNORE INTO links
                        (url, created_at, updated_at, saved_at)
                        VALUES(:url, :now, :now, :date);""",
                    link,
                )
            db.commit()


def import_instapaper_csv(csvpath: str, dbpath: str) -> None:
    db = sqlite3.connect(dbpath)
    db.row_factory = sqlite3.Row

    now = datetime.datetime.now()

    with open(csvpath, newline="") as csvpath:
        reader = csv.DictReader(csvpath)
        for item in reader:
            url = normalise_url(item["URL"])
            title = item["Title"]
            description = item["Selection"]
            saved_at = datetime.datetime.fromtimestamp(int(item["Timestamp"]))
            read_at = 0 if item["Folder"] == "Archive" else None
            tags = {"instapaper"}

            if item["Folder"] not in ["Archive", "Unread"]:
                tags.add(item["Folder"].lower())

            create_or_update(db, url, tags, title, description, read_at, saved_at, now)


def import_json(jsonpath: str, dbpath: str) -> None:
    db = sqlite3.connect(dbpath)
    db.row_factory = sqlite3.Row
    cursor = db.cursor()

    now = datetime.datetime.now()

    data = json.load(open(jsonpath))
    for row in data:
        url = normalise_url(row["href"])
        title = row["description"]
        description = row["extended"]
        tags = set(row["tags"].split(" "))
        saved_at = row["time"]
        read_at = 0 if row["toread"] == "no" else None

        tags.add("pinboard")

        create_or_update(db, url, tags, title, description, read_at, saved_at, now)


if __name__ == "__main__":
    csvpath, dbpath, *_ = sys.argv[1:]
    import_instapaper_csv(csvpath, dbpath)
    import_dropbox_csv(
        "/Users/ben/Library/CloudStorage/Dropbox/IFTTT/Instapaper/Saved Items.txt",
        dbpath,
        False,
    )
    import_dropbox_csv(
        "/Users/ben/Library/CloudStorage/Dropbox/IFTTT/Instapaper/Archived Items.txt",
        dbpath,
        True,
    )
