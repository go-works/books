"use strict";

var fs = require('fs');
var commonmark = require('commonmark/lib/index.js');
var Renderer = require('commonmark/lib/render/renderer.js');

var reXMLTag = /\<[^>]*\>/;

function MdRenderer(options) {
    options = options || {};

    this.disableTags = 0;
    this.lastOut = "\n";

    this.indentLevel = 0;
    this.indent = '  ';

    this.options = options;
}

function render(ast) {

    this.buffer = '';

    var attrs;
    var tagname;
    var walker = ast.walker();
    var event, node, entering;
    var container;
    var selfClosing;
    var nodetype;

    var options = this.options;

    while ((event = walker.next())) {
        entering = event.entering;
        node = event.node;
        nodetype = node.type;

        container = node.isContainer;

        selfClosing = nodetype === 'thematic_break'
            || nodetype === 'linebreak'
            || nodetype === 'softbreak';

        tagname = nodetype;

        if (entering) {

            attrs = [];

            switch (nodetype) {
                case 'document':
                    break;
                case 'list':
                    if (node.listType !== null) {
                        attrs.push(['type', node.listType.toLowerCase()]);
                    }
                    if (node.listStart !== null) {
                        attrs.push(['start', String(node.listStart)]);
                    }
                    if (node.listTight !== null) {
                        attrs.push(['tight', (node.listTight ? 'true' : 'false')]);
                    }
                    var delim = node.listDelimiter;
                    if (delim !== null) {
                        var delimword = '';
                        if (delim === '.') {
                            delimword = 'period';
                        } else {
                            delimword = 'paren';
                        }
                        attrs.push(['delimiter', delimword]);
                    }
                    break;
                case 'code_block':
                    if (node.info) {
                        attrs.push(['info', node.info]);
                    }
                    break;
                case 'heading':
                    attrs.push(['level', String(node.level)]);
                    break;
                case 'link':
                case 'image':
                    attrs.push(['destination', node.destination]);
                    attrs.push(['title', node.title]);
                    break;
                case 'custom_inline':
                case 'custom_block':
                    attrs.push(['on_enter', node.onEnter]);
                    attrs.push(['on_exit', node.onExit]);
                    break;
                default:
                    break;
            }

            this.cr();
            this.out(this.tag(tagname, attrs, selfClosing));
            if (container) {
                this.indentLevel += 1;
            } else if (!container && !selfClosing) {
                var lit = node.literal;
                if (lit) {
                    this.out(" " + this.esc(lit));
                }
            }
        } else {
            this.indentLevel -= 1;
            this.cr();
        }
    }
    this.buffer += '\n';
    return this.buffer;
}

function out(s) {
    if (this.disableTags > 0) {
        this.buffer += s.replace(reXMLTag, '');
    } else {
        this.buffer += s;
    }
    this.lastOut = s;
}

function cr() {
    if (this.lastOut !== '\n') {
        this.buffer += '\n';
        this.lastOut = '\n';
        for (var i = this.indentLevel; i > 0; i--) {
            this.buffer += this.indent;
        }
    }
}

// Helper function to produce an XML tag.
function tag(name, attrs, selfclosing) {
    var result = name;
    if (attrs && attrs.length > 0) {
        var i = 0;
        var attrib;
        while ((attrib = attrs[i]) !== undefined) {
            result += ' ' + attrib[0] + '="' + this.esc(attrib[1]) + '"';
            i++;
        }
    }
    return result;
}

// quick browser-compatible inheritance
MdRenderer.prototype = Object.create(Renderer.prototype);
MdRenderer.prototype.render = render;
MdRenderer.prototype.out = out;
MdRenderer.prototype.cr = cr;
MdRenderer.prototype.tag = tag;
MdRenderer.prototype.esc = require('commonmark/lib/common').escapeXml;

var parser = new commonmark.Parser({ smart: true });
var renderer = new MdRenderer();

var file = process.argv[2];
var contents = fs.readFileSync(file, 'utf8');

var ast = parser.parse(contents);
var s = renderer.render(ast);
console.log(s);
