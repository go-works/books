#!/usr/bin/env python3

# https://github.com/jamalex/notion-py
# https://medium.com/@jamiealexandre/introducing-notion-py-an-unofficial-python-api-wrapper-for-notion-so-603700f92369

import os
import random
import time
from notion.client import NotionClient
from notion.operations import build_operation

javascriptTop = "0b121710a160402fa9fd4646b87bed99"
cppTop = "ad527dc6d4a7420b923494d0b9bfb560"
expectedIdLen = len("0b121710a160402fa9fd4646b87bed99")

toFixTest = "https://www.notion.so/kjkpublic/000-How-to-make-iterator-usable-inside-async-callback-function-ee2c02d47ef44fbf883238558e314394"


def normalize_id(s):
    s = s.replace("-", "")
    if len(s) != expectedIdLen:
        raise "unexpected id '" + s + "'"
    return s


def is_number(s):
    isn = False
    try:
        n = int(s)
        isn = True
    except:
        pass
    # print("is number:", s, ", is:", isn)
    return isn


def clean_title(s):
    parts = s.split(" ")
    if len(parts) == 1:
        return s
    if is_number(parts[0]):
        parts = parts[1:]
    return " ".join(parts)


def should_update_format(fmt):
    if fmt is None:
        return True
    fw = fmt.get("page_full_width", False)
    sp = fmt.get("page_small_text", False)
    return (not fw) or (not sp)


def fix_format(page):
    fmt = page.get(["format"])
    #print("format: %s of class %s" % (fmt, fmt.__class__.__name__))
    if not should_update_format(fmt):
        return
    args = {'page_full_width': True, 'page_small_text': True}
    path = ["format"]
    table = page._table
    op = build_operation(id=page.id, path=path,
                         args=args, table=table, command="update")
    res = page._client.submit_transaction(op)
    print("  updated page format")


def fix_title(page):
    new_title = clean_title(page.title)
    if new_title != page.title:
        print("  '%s' to '%s'" % (page.title, new_title))
        page.title = new_title


def get_subpages_may_throw(page):
    sub_pages = []
    for child in page.children:
        tp = child.__class__.__name__
        if tp == "PageBlock":
            sub_pages.append(child.id)
    return sub_pages


def get_subpages(page):
    try:
        res = get_subpages_may_throw(page)
        return res
    except:
        sleep_secs = 3
        print("get_subpages_may_throw() failed, sleeping for %d seconds" % sleep_secs)
        time.sleep(sleep_secs)
        return []


def clean_titles_and_format(start_id):
    api_token = os.environ.get("NOTION_TOKEN", "")
    if api_token == "":
        print("need NOTION_TOKEN env variable")
        exit(1)
    client = NotionClient(token_v2=api_token)

    n_pages = 0
    to_visit = [normalize_id(start_id)]
    visited = {}
    while len(to_visit) > 0:
        n_pages += 1
        # randomize order just for fun
        random.shuffle(to_visit)
        page_id = to_visit[0]
        to_visit = to_visit[1:]
        normalized_page_id = normalize_id(page_id)
        if page_id in visited:
            print("Skipping '%s' because already visited" % page_id)
            continue
        print("Pages left: %d, getting page %d: %s..." %
              (len(to_visit)+1, n_pages, page_id), end='')
        page = client.get_block(page_id)
        page.refresh()
        visited[normalized_page_id] = True

        print(" got '%s'" % page.title)
        try:
            fix_title(page)
            fix_format(page)
        except:
            print("fix_title or fix_format threw an exception")
            time.sleep(3)

        subpages = get_subpages(page)
        to_visit.extend(subpages)


def main():
    sql_id = "d1c8bb39bad4494e80abe28414c3d80e"
    python_id = "12e6f78e68a5497290c96e1365ae6259"  # finished partially, rerun?
    javascript_id = "0b121710a160402fa9fd4646b87bed99"  # finished with 50 left

    # TODO:
    android_id = "f90b0a6b648343e28dc5ed6e8f5c0780"  # need to re-run
    java_id = "d37cda98a07046f6b2cc375731ea3bdb"

    kotlin_id = "2bdd47318f3a4e8681dda289a8b3472b"  # only format
    postgresql_id = "799304340f2c4081b6c4b7eb28df368e"  # only format
    dart_id = "0e2d248bf94b4aebaefbcf51ae435df0"  # only format

    cpp_id = "ad527dc6d4a7420b923494d0b9bfb560"  # only format
    mysql_id = "4489ab73989f4ae9912486561e165deb"  # seems done
    ios_id = "3626edc1bd044431afddc89648a7050f"  # mostly done

    clean_titles_and_format(java_id)


if __name__ == "__main__":
    main()
