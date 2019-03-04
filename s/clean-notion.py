#!/usr/bin/env python3

# https://github.com/jamalex/notion-py
# https://medium.com/@jamiealexandre/introducing-notion-py-an-unofficial-python-api-wrapper-for-notion-so-603700f92369

import os
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
    fw = fmt["page_full_width"]
    sp = fmt["page_small_text"]
    should_update = not fw or not sp
    return should_update


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


def get_subpages(page):
    sub_pages = []
    for child in page.children:
        tp = child.__class__.__name__
        if tp == "PageBlock":
            sub_pages.append(child.id)
    return sub_pages


def clean_titles_and_format(start_id):
    api_token = os.environ.get("NOTION_TOKEN", "")
    if api_token == "":
        print("need NOTION_TOKEN env variable")
        exit(1)
    client = NotionClient(token_v2=api_token)

    to_visit = [normalize_id(start_id)]
    visited = {}
    while len(to_visit) > 0:
        page_id = to_visit[0]
        to_visit = to_visit[1:]
        normalized_page_id = normalize_id(page_id)
        print("Pages left: %d, getting page: %s" % (len(to_visit)+1, page_id))
        if page_id in visited:
            print("Skipping page %s because already visited" % page_id)
            continue
        page = client.get_block(page_id)
        page.refresh()
        visited[normalized_page_id] = True

        print("Got page with title '%s' and id '%s'" % (page.title, page.id))
        fix_title(page)
        fix_format(page)

        subpages = get_subpages(page)
        to_visit.extend(subpages)


def main():
    sql_id = "d1c8bb39bad4494e80abe28414c3d80e"
    clean_titles_and_format(sql_id)


if __name__ == "__main__":
    main()
