import { NotionLoader } from './notion.ts';

const startPageID = ('2cab1ed2b7a44584b56b0d3ca9b80185');

const loader = new NotionLoader();

console.log("id:", startPageID);
await loader.loadPage(startPageID);
