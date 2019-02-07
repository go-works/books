"use strict";

var fs = require('fs');
var commonmark = require('commonmark/lib/index.js');
var AstRenderer = require('./astrender.js');

var parser = new commonmark.Parser();

var astrenderer = new AstRenderer();

var file = process.argv[2];
var contents = fs.readFileSync(file, 'utf8');

var ast = parser.parse(contents);
var s = astrenderer.render(ast);
console.log(s);
