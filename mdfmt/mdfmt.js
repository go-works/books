"use strict";

var fs = require('fs');
var commonmark = require('commonmark/lib/index.js');
var MdRenderer = require('./mdrender.js');
var AstRenderer = require('./astrender.js');

var parser = new commonmark.Parser({ smart: true });

var mdrenderer = new MdRenderer();
var astrenderer = new AstRenderer();

var file = process.argv[2];
var contents = fs.readFileSync(file, 'utf8');

var ast = parser.parse(contents);
var s = mdrenderer.render(ast);
console.log(s);
