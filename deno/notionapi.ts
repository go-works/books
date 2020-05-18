const notionHost = "https://www.notion.so";
// modern Chrome
const userAgent  = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3483.0 Safari/537.36";
const acceptLang = "en-US,en;q=0.9";
const apiGetRecordValues = "/api/v3/getRecordValues";
const apiGetSignedFileUrls = "/api/v3/getSignedFileUrls";
const apiLoadPageChunk = "/api/v3/loadPageChunk";
const apiLoadUserContent = "/api/v3/loadUserContent";
const apiQueryCollection = "/api/v3/queryCollection";

class Client {
  AuthToken: string = "";
}

function len(s: string): number {
  return s.length;
}

const dashIDLen   = len("2131b10c-ebf6-4938-a127-7089ff02dbe4");
const noDashIDLen = len("2131b10cebf64938a1277089ff02dbe4");

function toDashID(id: string): string {
  const s: string = id.replace("-", "");
  if (len(s) != noDashIDLen) {
		return id
	}
	const res = id.substr(0, 8) + "-" + id.substr(8, 4) + "-" + id.substr(12, 4) + "-" + id.substr(16, 4) + "-" + id.substr(20);
	return res
}

function isValidDashID(id: string): boolean {
	if (len(id) != dashIDLen) {
		return false;
	}
	if (id[8] != '-' ||
		id[13] != '-' ||
		id[18] != '-' ||
		id[23] != '-') {
		return false;
  }
	// for i := range id {
	// 	if !isValidDashIDChar(id[i]) {
	// 		return false
	// 	}
	// }
	return true;
}

function doApi(c: Client, apiURL: string, requestData: any): any {
  return null;
}

// TODO: return Page object
function downloadPage(c: Client, pageID: string): any {
	const id = toDashID(pageID)
	if (!isValidDashID(id)) {
    throw new Error(`{id} is not a valid Notion page id"`);
  }

  return null;
}

console.log("hello");
