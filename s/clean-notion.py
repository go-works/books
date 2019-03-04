#!/usr/bin/env python3

# https://github.com/jamalex/notion-py
# https://medium.com/@jamiealexandre/introducing-notion-py-an-unofficial-python-api-wrapper-for-notion-so-603700f92369

import os
from notion.client import NotionClient

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

def fix_title(s):
    parts = s.split(" ")
    if len(parts) == 1:
        return s
    if is_number(parts[0]):
        parts = parts[1:]
    return " ".join(parts)

def main():
    api_token = os.environ.get("NOTION_TOKEN", "")
    if api_token == "":
        print("need NOTION_TOKEN env variable")
        exit(1)
    client = NotionClient(token_v2=api_token)

    to_visit = [normalize_id(cppTop)]
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
        visited[normalized_page_id] = True
        title = page.title
        page_id = page.id
        print("Got page with title '%s' and id '%s'" % (title, page_id))
        #fw = page.get("full_width")
        #print("page.full_width:", fw)
        #print("page.small_text:", page.small_text)
        new_title = fix_title(title)
        if new_title != title:
            print("Changing title from '%s' to '%s'" % (title, new_title))
            page.title = new_title
        n_sub_pages = 0
        for child in page.children:
            tp = child.__class__.__name__
            if tp == "PageBlock":
                n_sub_pages += 1
                to_visit.append(child.id)
        if n_sub_pages > 0:
            print("Has %d sub-pages" % n_sub_pages)

if __name__ == "__main__":
    main()
