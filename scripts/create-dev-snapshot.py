#!/usr/bin/env python3
"""Create an isolated, credential-reset development snapshot without writing to production."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import secrets
import shutil
import sqlite3


def create_snapshot(source, target):
    source, target = source.resolve(), target.resolve()
    if source == target or source in target.parents or target in source.parents:
        raise ValueError("Snapshot must be separate from the source")
    source_db = source / "database/database.db"
    if not source_db.is_file():
        raise ValueError("Source database does not exist")
    target.mkdir(parents=True, exist_ok=True, mode=0o700)
    os.chmod(target, 0o700)
    database = target / "database"
    database.mkdir(mode=0o700, exist_ok=True)
    destination_db = database / "database.db"
    if destination_db.exists():
        raise ValueError("Refusing to overwrite an existing development database")
    with sqlite3.connect(source_db.as_uri() + "?mode=ro", uri=True) as src:
        with sqlite3.connect(destination_db) as dest:
            src.backup(dest)
    os.chmod(destination_db, 0o600)
    for name in ("uploads", "custom"):
        if (source / name).is_dir():
            shutil.copytree(source / name, target / name, dirs_exist_ok=True)
        else:
            (target / name).mkdir(exist_ok=True)
    for name in ("favicon.ico",):
        if (source / name).is_file():
            shutil.copy2(source / name, target / name)
    # Generate separate logins and discard every production password and token.
    with sqlite3.connect(destination_db) as db:
        credentials = []
        for user_id, in db.execute("SELECT id FROM user WHERE deleted_at IS NULL"):
            password = secrets.token_urlsafe(24)
            hashed = password
            for _ in range(3):
                hashed = hashlib.md5(hashed.encode()).hexdigest()
            username = "dev-" + str(user_id) + "@usagi.local"
            db.execute("UPDATE user SET username=?,mail=?,password=?,token='',referral_code='' WHERE id=?", (username, username, hashed, user_id))
            credentials.append({"userId": user_id, "username": username, "password": password})
        # Reset deleted accounts too, preventing recovery of production credentials.
        db.execute("UPDATE user SET password='',token='',referral_code='' WHERE deleted_at IS NOT NULL")
        tables = {row[0] for row in db.execute("SELECT name FROM sqlite_master WHERE type='table'")}
        if "open_api_token" in tables:
            db.execute("DELETE FROM open_api_token")
        # Configuration from the current installation may contain SMTP or registration secrets.
        db.execute("DELETE FROM system_setting WHERE config_name NOT IN ('panel_public_user_id')")
        report = {
            "groups": db.execute("SELECT COUNT(*) FROM item_icon_group WHERE deleted_at IS NULL").fetchone()[0],
            "privateGroups": db.execute("SELECT COUNT(*) FROM item_icon_group WHERE deleted_at IS NULL AND public_visit_hide=1").fetchone()[0],
            "cards": db.execute("SELECT COUNT(*) FROM item_icon WHERE deleted_at IS NULL").fetchone()[0],
        }
    credentials_file = target / "dev-credentials.json"
    fd = os.open(credentials_file, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
    with os.fdopen(fd, "w") as file:
        json.dump(credentials, file, ensure_ascii=False)
    (target / "conf.ini").write_text("[base]\nhttp_port=3002\ndatabase_drive=sqlite\ncache_drive=memory\nqueue_drive=memory\nsource_path=./uploads\nsource_temp_path=./runtime/temp\n\n[sqlite]\nfile_path=./database/database.db\n")
    return report


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--source-conf", type=Path, required=True)
    parser.add_argument("--target-conf", type=Path, required=True)
    args = parser.parse_args()
    print(json.dumps(create_snapshot(args.source_conf, args.target_conf)))
