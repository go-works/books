import { toDashID } from './notionapi.ts';

export interface NotionMeta {
  slug?: string;
  date?: string;
  tags?: string[];
  isDraft?: boolean;
  excerpt?: string;
}

export interface NotionPageAtt {
  att: string;
  value?: string;
}

export interface NotionPageText {
  text: string;
  atts: NotionPageAtt[];
}

export interface NotionPageProperty {
  propName: string;
  value: NotionPageText[];
}

export interface NotionPageBlock {
  type: string;
  blockId: string;
  properties: NotionPageProperty[];
  attributes: NotionPageAtt[];
  blockIds: string[];
}

export interface NotionPageImage {
  pageId: string;
  notionUrl: string;
  signedUrl: string;
  contentId: string;
}

export interface NotionImageNodes {
  imageUrl: string;
  localFile: {
    publicURL: string;
  };
}

export interface NotionPageLinkedPage {
  title: string;
  pageId: string;
}

export interface NotionPageDescription {
  pageId: string;
  title: string;
  indexPage: number;
  slug: string;
  excerpt: string;
  pageIcon: string;
  createdAt: string;
  tags: string[];
  isDraft: boolean;
  blocks: NotionPageBlock[];
  images: NotionPageImage[];
  linkedPages: NotionPageLinkedPage[];
}

// generic type to hold json data
export type JsonTypes = string | number | boolean | Date | Json | JsonArray;
export interface Json {
  [x: string]: JsonTypes;
}
export type JsonArray = Array<JsonTypes>;

/*
// plugin configuration data
export interface NotionsoPluginOptions extends PluginOptions {
  rootPageUrl: string;
  name: string;
  tokenv2?: string;
  downloadLocal: boolean;
  debug?: boolean;
}
*/

export interface NotionLoaderImageInformation {
  imageUrl: string;
  contentId: string;
}

export interface NotionLoaderImageResult {
  imageUrl: string;
  contentId: string;
  signedImageUrl: string;
}

export interface NotionLoader {
  loadPage(pageId: string): Promise<void>;
  downloadImages(
    images: NotionLoaderImageInformation[]
  ): Promise<NotionLoaderImageResult[]>;
  getBlockById(blockId: string): NotionPageBlock | undefined;
  getBlocks(copyTo: NotionPageBlock[], pageId: string): void;
  reset(): void;
}

export function extractPageIdFromPublicUrl(url: string): string | null {
  const len = url.length;
  if (len < 32) {
    return null;
  }
  // we first take the last 32 digits
  const id = url.substring(len - 32);

  // then we need to format as:
  // xxxxxxxx-yyyy-yyyy-yyyy-zzzzzzzzzzzz
  const sliceLengths = [8, 4, 4, 4, 12];
  const slices: string[] = [];
  let previous = 0;
  sliceLengths.forEach((slen) => {
    slices.push(id.substring(previous, previous + slen));
    previous += slen;
  });
  return slices.join("-");
}

export function notionPageTextToString(text: NotionPageText[]): string {
  const parts: string[] = [];
  text.forEach((t) => {
    parts.push(t.text);
  });
  return parts.join("");
}

// example:
// 'slug: first_page\ndate: 2019/12/31'
export function parseMetaBlock(
  data: NotionPageText[],
  meta: Record<string, string>
): boolean {
  const text = notionPageTextToString(data);
  const lines = text.split(/\r?\n/);
  let isMeta = false;
  lines.forEach((line) => {
    const pos = line.indexOf(":");
    if (pos > 0) {
      const key = line.substring(0, pos).trim();
      const value = line.substring(pos + 1).trim();
      if (key.length > 0 && value.length > 0) {
        meta[key] = value;
        isMeta = true;
      }
    } else {
      isMeta = false;
    }
  });
  return isMeta;
}

// line starting with that chacracter are meta information
// about the page
const META_MARKER_TAGS = "!";

/*
export interface NotionMeta {
  slug?: string;
  date?: Date;
  tags?: string[];
  isDraft: boolean;
}
*/
export function parseArrayString(line: string): string[] {
  const result: string[] = [];
  line
    .split(",")
    .map((t) => t.trim())
    .forEach((t) => result.push(t));

  return result;
}

export function parseDateValue(line: string): string {
  return new Date(line + " Z").toJSON();
}

export function parseSlug(line: string): string {
  return line.replace(/\W+/g, "-");
}

export function parseMetaLine(line: string): [boolean, string, string] {
  const l1 = line.trim();
  if (!l1.startsWith(META_MARKER_TAGS)) {
    return [false, "", ""];
  }
  const l2 = l1.substring(1).trim();
  const pos = l2.indexOf(":");
  if (pos > 0) {
    return [
      true,
      l2.substring(0, pos).trim().toLowerCase(),
      l2.substring(pos + 1).trim(),
    ];
  }

  const pos2 = l2.indexOf(" ");
  if (pos2 > 0) {
    return [
      true,
      l2.substring(0, pos2).trim().toLowerCase(),
      l2.substring(pos2 + 1).trim(),
    ];
  }
  return [true, l2.toLowerCase(), ""];
}

export function parseBooleanValue(line: string): boolean {
  if (line.length === 0) {
    return true;
  }
  if (line === "0" || line === "false") {
    return false;
  }
  return true;
}

export function parseMetaText(meta: NotionMeta): (arg: string) => boolean {
  return (line: string): boolean => {
    const [isMetaLine, keyword, value] = parseMetaLine(line);
    if (!isMetaLine) {
      return false;
    }
    let isMeta = true;
    if (keyword === "draft") {
      meta.isDraft = parseBooleanValue(value);
    } else if (keyword === "date") {
      meta.date = parseDateValue(value);
    } else if (keyword === META_MARKER_TAGS) {
      meta.excerpt = value;
    } else if (keyword === "slug") {
      meta.slug = parseSlug(value);
    } else if (keyword === "tags") {
      meta.tags = parseArrayString(value);
    } else {
      isMeta = false;
    }
    return isMeta;
  };
}

export type NotionTextAttributes = string[][];

export type NotionText = [string, NotionTextAttributes?][];

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type MyObj = Record<string, any>;

// eslint-disable-next-line @typescript-eslint/no-unused-vars
function parseNotionText(text: NotionText): NotionPageText[] {
  const result: NotionPageText[] = [];
  text.forEach(([str, att]) => {
    const item: NotionPageText = {
      text: str,
      atts: [],
    };
    if (att) {
      att.forEach(([attName, ...rest]) => {
        item.atts.push({
          att: attName,
          value: rest && rest[0],
        });
      });
    }
    result.push(item);
  });
  return result;
}

function get(o: object, path: string, defaultValue: string): string {
  if (!o) {
    return defaultValue;
  }
  const parts = path.split(".");
  let co: any = o;
  for (let key of parts) {
    co = co[key];
    if (!co) {
      return defaultValue;
    }
  }
  return co.toString();
}

function recordToBlock(value: Json): NotionPageBlock | null {
  const block: NotionPageBlock = {
    type: value.type as string,
    blockId: value.id as string,
    properties: [],
    attributes: [],
    blockIds: [],
  };
  const properties: MyObj = (value.properties as object) || {};
  Object.keys(properties).forEach((propName) => {
    block.properties.push({
      propName,
      value: parseNotionText(properties[propName] as NotionText),
    });
  });
  ((value.content as []) || []).forEach((id) => block.blockIds.push(id));

  // extra attributes to grab for images
  if (block.type === "image") {
    block.attributes.push({
      att: "width",
      value: get(value as object, "format.block_width", "-1"),
    });
    block.attributes.push({
      att: "aspectRatio",
      value: get(value as object, "format.block_aspect_ratio", "-1"),
    });
  }
  if (block.type === "page") {
    block.attributes.push({
      att: "pageIcon",
      value: get(value as object, "format.page_icon", ""),
    });
  }
  return block;
}

export function recordMapToBlocks(
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  recordMap: any,
  blocks: NotionPageBlock[]
): NotionPageBlock[] {
  Object.keys(recordMap.block).forEach((key) => {
    const block = recordToBlock(recordMap.block[key].value as Json);
    if (block) {
      blocks.push(block);
    }
  });
  return blocks;
}

function getPropertyAsString(
  block: NotionPageBlock,
  propName: string,
  defaultValue: ""
): string {
  const property = block.properties.find((p) => p.propName === propName);
  if (!property) {
    return defaultValue;
  }
  return notionPageTextToString(property.value);
}

function getAttributeAsString(
  block: NotionPageBlock,
  attName: string,
  defaultValue: ""
): string {
  const att = block.attributes.find((p) => p.att === attName);
  if (!att || !att.value) {
    return defaultValue;
  }
  return att.value;
}

export async function loadPage(
  pageId: string,
  rootPageId: string,
  indexPage: number,
  notionLoader: NotionLoader,
  debug: boolean
): Promise<NotionPageDescription> {
  // we load the given page
  await notionLoader.loadPage(pageId);

  // and parse its description block
  const page = notionLoader.getBlockById(pageId);
  if (!page) {
    //reporter.error(`could not retreieve page with id: ${pageId}`);
    throw Error("error retrieving page");
  }

  if (page.type !== "page") {
    throw new Error("invalid page");
  }

  const imageDescriptions: NotionPageImage[] = [];
  const linkedPages: NotionPageLinkedPage[] = [];
  const meta: NotionMeta = {};
  const metaParser = parseMetaText(meta);

  // parse all the blocks retrieved from notion
  for (const blockId of page.blockIds) {
    const block = notionLoader.getBlockById(blockId);
    if (!block) {
      //reporter.error(`could not retrieve block with id: ${blockId}`);
      throw Error("error retrieving block in page");
    }
    switch (block.type) {
      case "page":
        linkedPages.push({
          pageId: block.blockId,
          title: getPropertyAsString(block, "title", ""),
        });
        break;
      case "text":
        {
          // for the text blocks, we parse them to see if they contain
          // meta attributes, if not, we addf them as regular blocks
          const text = getPropertyAsString(block, "title", "").trim();
          if (metaParser(text)) {
            // we change the type to meta to avoid the rendering of this text block
            block.type = "meta";
          }
        }
        break;
      case "image":
        imageDescriptions.push({
          pageId,
          notionUrl: getPropertyAsString(block, "source", ""),
          signedUrl: "",
          contentId: block.blockId,
        });
        break;
      case "ignore":
        // guess what... we ignore that one
        break;
      default:
        // we keep the block by defaut
        break;
    }
  }

  const item: NotionPageDescription = {
    pageId,
    title: getPropertyAsString(page, "title", ""),
    indexPage,
    slug: meta.slug || `${indexPage}`,
    createdAt: meta.date || new Date().toISOString(),
    tags: meta.tags || [],
    isDraft: !!meta.isDraft,
    excerpt: meta.excerpt || "",
    pageIcon: getAttributeAsString(page, "pageIcon", ""),
    blocks: [],
    images: imageDescriptions,
    linkedPages,
  };
  // we return all the blocks
  // TODO: as we already got those blocks above, we may want to build the list as we go
  notionLoader.getBlocks(item.blocks, rootPageId);
  return item;
}

interface NotionApiDownloadInfo {
  url: string;
  permissionRecord: {
    table: string;
    id: string;
  };
}

export function notionLoader(debug = true): NotionLoader {
  let _blocks: NotionPageBlock[] = [];

  return {
    loadPage: async (pageId: string): Promise<void> => {
      const urlLoadPageChunk = "https://www.notion.so/api/v3/loadPageChunk";

      pageId = toDashID(pageId);
      const postData = {
        pageId: pageId,
        limit: 100000,
        cursor: { stack: [] },
        chunkNumber: 0,
        verticalColumns: false,
      };

      const options: RequestInit = {
        method: "POST",
        headers: {
          "content-type": "application/json",
          credentials: "include",
          accept: "*/*",
          "accept-language": "en-US,en;q=0.9,fr;q=0.8",
        },
        body: JSON.stringify(postData, null, 0),
      };
      const response = await fetch(urlLoadPageChunk, options);
      if (response.status !== 200) {
        // reporter.error(
        //   `error retrieving data from notion. status=${response.status}`
        // );
        throw new Error(
          `Error retrieving data - status: ${response.status}`
        );
      }
      const data: any = await response.json()
      console.log("data:", data);
      recordMapToBlocks((data && data.recordMap) || {}, _blocks);
    },

    downloadImages(
      images: NotionLoaderImageInformation[]
    ): Promise<NotionLoaderImageResult[]> {
      const urlGetSignedFileUrls =
        "https://www.notion.so/api/v3/getSignedFileUrls";
      const urls: NotionApiDownloadInfo[] = [];
      images.forEach((image) => {
        urls.push({
          url: image.imageUrl,
          permissionRecord: {
            table: "block",
            id: image.contentId,
          },
        });
      });

      const dataForUrls = {
        urls,
      };

      const options: RequestInit = {
        method: "POST",
        headers: {
          "content-type": "application/json",
          credentials: "include",
          accept: "*/*",
          "accept-language": "en-US,en;q=0.9,fr;q=0.8",
        },
        body: JSON.stringify(dataForUrls, null, 0),
      };

      const result: NotionLoaderImageResult[] = [];

      return fetch(urlGetSignedFileUrls, options)
        .then(function (response) {
          if (response.status !== 200) {
            // reporter.error(
            //   `Error retrieving images ${images} , status is: ${response.status}`
            // );
          } else {
            // if (debug) {
            //   console.log(
            //     util.inspect(response.data, {
            //       colors: true,
            //       depth: null,
            //     })
            //   );
            // }
            response.json().then(function (data: any) {
              const arr: string[] =
                (data && (data.signedUrls as string[])) || ([] as string[]);
              arr.forEach((signedUrl, index) => {
                result.push({
                  imageUrl: images[index].imageUrl,
                  contentId: images[index].contentId,
                  signedImageUrl: signedUrl,
                });
              });
            });
          }
          return result;
        })
        .catch(function (error) {
          console.log("Error:");
          console.log(error);
          return result;
        });
    },
    getBlockById(blockId: string): NotionPageBlock | undefined {
      return _blocks.find((b) => b.blockId === blockId);
    },
    getBlocks(copyTo: NotionPageBlock[], pageId: string): void {
      _blocks
        .filter((b) => b.blockId !== pageId)
        .forEach((b) => copyTo.push(b));
    },
    reset(): void {
      _blocks = [];
    },
  };
}
