"use strict";

var fs = require('fs');
var commonmark = require('commonmark/lib/index.js');
var MdRenderer = require('./mdrender.js');
var mdrender = require('./mdrender2.js');

var parser = new commonmark.Parser({ smart: true });

var mdrenderer = new MdRenderer();

var file = process.argv[2];
var contents = fs.readFileSync(file, 'utf8');

var ast = parser.parse(contents);
//var s = mdrenderer.render(ast);
var s = mdrender(ast);
console.log(s);
