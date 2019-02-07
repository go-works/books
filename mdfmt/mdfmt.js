"use strict";

/*
To test:
node ./mdfmt/mdfmt.js "books/csharp-language_fmt/0010-getting-started-with-csharp-language/010 Creating a new console application Visual Studio.md"
*/

var fs = require('fs');
var commonmark = require('commonmark/lib/index.js');
var mdrender = require('./mdrender.js');

var parser = new commonmark.Parser({ smart: true });

var file = process.argv[2];
var contents = fs.readFileSync(file, 'utf8');

var ast = parser.parse(contents);
var s = mdrender(ast);
console.log(s);
