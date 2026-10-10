#!/usr/bin/env python3
"""Verify a private development instance against a read-only source database."""
import argparse
import json
from pathlib import Path
import sqlite3
import urllib.request


def rows(db, table):
    columns = [row[1] for row in db.execute('PRAGMA table_info(' + table + ')')]
    return columns, db.execute('SELECT ' + ','.join('"' + c + '"' for c in columns) + ' FROM ' + table + ' ORDER BY id').fetchall()


def verify(source_conf, conf, base_url):
    with sqlite3.connect((source_conf.resolve() / 'database/database.db').as_uri() + '?mode=ro', uri=True) as source:
        with sqlite3.connect((conf.resolve() / 'database/database.db').as_uri() + '?mode=ro', uri=True) as copy:
            for table in ('item_icon', 'item_icon_group'):
                columns, expected = rows(source, table)
                actual = copy.execute('SELECT ' + ','.join('"' + c + '"' for c in columns) + ' FROM ' + table + ' ORDER BY id').fetchall()
                assert actual == expected, table + ' differs from source'
            owner = json.loads(source.execute("SELECT config_value FROM system_setting WHERE config_name='panel_public_user_id'").fetchone()[0])
            groups = source.execute('SELECT id,public_visit_hide FROM item_icon_group WHERE deleted_at IS NULL AND user_id=?', (owner,)).fetchall()
            cards = source.execute('SELECT id,title,url,icon_json,item_icon_group_id FROM item_icon WHERE deleted_at IS NULL AND user_id=?', (owner,)).fetchall()
            credentials = next(c for c in json.loads((conf / 'dev-credentials.json').read_text()) if c['userId'] == owner)

    def post(path, body=None, token=None):
        headers = {'Content-Type': 'application/json'}
        if token:
            headers['token'] = token
        req = urllib.request.Request(base_url.rstrip('/') + '/api/' + path, data=json.dumps(body).encode(), headers=headers)
        with urllib.request.urlopen(req, timeout=20) as response:
            return json.load(response)

    hidden = {g for g, hide in groups if hide == 1}
    public = post('panel/itemIconGroup/getList')
    assert public['code'] == 0, 'Public group request failed'
    assert {g['id'] for g in public['data']['list']} == {g for g, hide in groups if hide != 1}, 'Public visibility mismatch'
    visible_cards = []
    for group in public['data']['list']:
        response = post('panel/itemIcon/getListByGroupId', {'itemIconGroupId': group['id']})
        assert response['code'] == 0, 'Public cards request failed'
        visible_cards.extend(response['data']['list'])
    for group_id in hidden:
        assert post('panel/itemIcon/getListByGroupId', {'itemIconGroupId': group_id})['code'] != 0, 'Private group exposed'
    assert {c['id'] for c in visible_cards} == {c[0] for c in cards if c[4] not in hidden}, 'Public card mismatch'

    login = post('login', {'username': credentials['username'], 'password': credentials['password']})
    assert login['code'] == 0, 'Development login failed'
    token = login['data']['token']
    try:
        own = post('panel/itemIconGroup/getList', token=token)
        assert own['code'] == 0 and {g['id'] for g in own['data']['list']} == {g for g, _ in groups}, 'Owner group mismatch'
        own_cards = []
        for group in own['data']['list']:
            response = post('panel/itemIcon/getListByGroupId', {'itemIconGroupId': group['id']}, token)
            assert response['code'] == 0, 'Owner cards request failed'
            own_cards.extend(response['data']['list'])
        expected_cards = {c[0]: c for c in cards}
        assert len(own_cards) == len(cards), 'Owner card count mismatch'
        for card in own_cards:
            source_card = expected_cards[card['id']]
            assert (card['title'], card['url'], card['itemIconGroupId']) == (source_card[1], source_card[2], source_card[4]), 'Card data mismatch'
            assert card['icon'] == json.loads(source_card[3]), 'Icon data mismatch'
        return {'dataPreserved': True, 'groups': len(groups), 'cards': len(cards), 'privateGroups': len(hidden), 'publicGroups': len(public['data']['list']), 'publicCards': len(visible_cards), 'ownerCards': len(own_cards), 'privateAccessDenied': True, 'developmentLogin': True}
    finally:
        post('logout', token=token)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--source-conf', type=Path, required=True)
    parser.add_argument('--conf', type=Path, required=True)
    parser.add_argument('--base-url', default='http://127.0.0.1:3003')
    args = parser.parse_args()
    print(json.dumps(verify(args.source_conf, args.conf, args.base_url)))
