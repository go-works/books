"use strict";

/*
TODO:

Indent of:
1. item one
2. item two
   - sublist
   - sublist
*/

function Renderer() { }

function render(ast) {
  var walker = ast.walker()
    , event
    , type;

  this.buffer = '';
  this.lastOut = '\n';

  while ((event = walker.next())) {
    type = event.node.type;
    if (this[type]) {
      this[type](event.node, event.entering);
    }
  }
  return this.buffer;
}

function lit(str) {
  this.buffer += str;
  this.lastOut = str;
}

function cr() {
  if (this.lastOut !== '\n') {
    this.lit('\n');
  }
}

function out(str) {
  this.lit(str);
}

Renderer.prototype.render = render;
Renderer.prototype.out = out;
Renderer.prototype.lit = lit;
Renderer.prototype.cr = cr;

function MdRenderer(options) {
  options = options || {};

  this.lastOut = "\n";
  this.options = options;

  // TODO: probably needs to be for each nested level
  // of list
  this.listItemNo = 0;
}

/* Node methods */

function text(node) {
  this.out(node.literal);
}

function grandParentIsBlockQuote(node) {
  var gp = node.parent.parent;
  return ((gp !== null) && gp.type === 'block_quote');
}

function softbreak(node) {
  this.lit('\n');
  if (grandParentIsBlockQuote(node)) {
    this.lit("> ");
  }
}

function linebreak() {
  this.lit("\n");
  this.cr();
}

function link(node, entering) {
  if (entering) {
    this.lit('[');
  } else {
    this.lit('](' + node.destination + ')');
  }
}

function image(node, entering) {
  if (entering) {
    this.lit('![');
  } else {
    this.lit('](' + node.destination + ')');
  }
}

function emph(node, entering) {
  this.lit("*");
}

function strong(node, entering) {
  this.lit("**");
}

function skipParaNewline(node) {
  var p = node.parent;
  if (p === null) {
    return false;
  }
  if (p.type === 'block_quote') {
    return true;
  }
  var grandparent = node.parent.parent;
  if (grandparent !== null &&
    grandparent.type === 'list') {
    if (grandparent.listTight) {
      return true;
    }
  }
  return false;
}

function paragraph(node, entering) {
  if (skipParaNewline(node)) {
    return;
  }
  if (entering) {
    this.lit("\n");
    this.cr();
  } else {
    this.cr();
    this.lit("\n");
  }
}

function heading(node, entering) {
  if (entering) {
    this.cr();
    for (var i = 0; i < node.level; i++) {
      this.lit("#")
    }
    this.lit(" ");
  } else {
    this.cr();
    this.lit("\n");
  }
}

function code(node) {
  this.lit("`" + node.literal + "`");
}

function code_block(node) {
  var info_words = node.info ? node.info.split(/\s+/) : [];
  var lang = "";
  if (info_words.length > 0 && info_words[0].length > 0) {
    lang = info_words[0];
  }
  this.cr();
  this.lit("\n```" + lang);
  this.cr();
  this.out(node.literal);
  this.lit("```\n");
  this.cr();
}

function thematic_break(node) {
  this.cr();
  this.lit("\n---\n");
  this.cr();
}

function block_quote(node, entering) {
  if (entering) {
    this.cr();
    this.lit("> ");
  } else {
    this.cr();
  }
}

function list(node, entering) {
  if (entering) {
    var start = node.listStart || 1;
    this.listItemNo = start;
    this.cr();
  } else {
    this.cr();
  }
}

function getItemParent(node) {
  var p = node.parent;
  if (p.type != 'list') {
    console.log("item parent is not list but", p.type);
  }
  return p;
}

function item(node, entering) {
  if (entering) {
    var list = getItemParent(node);
    this.cr();
    var start = "* ";
    if (list.listType !== 'bullet') {
      start = this.listItemNo + ". ";
    }
    this.lit(start);
    this.listItemNo++;
  } else {
    this.cr();
  }
}

function html_inline(node) {
  this.lit(node.literal);
}

function html_block(node) {
  this.cr();
  this.lit(node.literal);
  this.cr();
}

function custom_inline(node, entering) {
  if (entering && node.onEnter) {
    this.lit(node.onEnter);
  } else if (!entering && node.onExit) {
    this.lit(node.onExit);
  }
}

function custom_block(node, entering) {
  this.cr();
  if (entering && node.onEnter) {
    this.lit(node.onEnter);
  } else if (!entering && node.onExit) {
    this.lit(node.onExit);
  }
  this.cr();
}

/* Helper methods */

function out(s) {
  this.lit(this.esc(s, false));
}

function escMd(s) {
  return s;
}

MdRenderer.prototype = Object.create(Renderer.prototype);

MdRenderer.prototype.text = text;
MdRenderer.prototype.html_inline = html_inline;
MdRenderer.prototype.html_block = html_block;
MdRenderer.prototype.softbreak = softbreak;
MdRenderer.prototype.linebreak = linebreak;
MdRenderer.prototype.link = link;
MdRenderer.prototype.image = image;
MdRenderer.prototype.emph = emph;
MdRenderer.prototype.strong = strong;
MdRenderer.prototype.paragraph = paragraph;
MdRenderer.prototype.heading = heading;
MdRenderer.prototype.code = code;
MdRenderer.prototype.code_block = code_block;
MdRenderer.prototype.thematic_break = thematic_break;
MdRenderer.prototype.block_quote = block_quote;
MdRenderer.prototype.list = list;
MdRenderer.prototype.item = item;
MdRenderer.prototype.custom_inline = custom_inline;
MdRenderer.prototype.custom_block = custom_block;

MdRenderer.prototype.esc = escMd;
MdRenderer.prototype.out = out;

module.exports = MdRenderer;