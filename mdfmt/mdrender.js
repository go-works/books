// add up to 2 \n at the end of the buffer
function nl(buf) {
    const n = buf.length;
    if (n < 2) {
        return buf;
    }
    if (buf[n - 1] != '\n') {
        return buf + "\n\n";
    }
    if (buf[n - 2] != '\n') {
        return buf + "\n";
    }
    return buf;
}

const needEscapingList = ["\\", "'", "*", "_", "{", "}", "[", "]", "(", ")", "#", "+", "-", ">", "<", "!"];
function needsEscaping(s) {
    if (needEscapingList.includes(s)) {
        return true;
    }
    return false;
}

function esc(s) {
    if (needsEscaping(s)) {
        return "\\" + s;
    }
    // TODO: if s is "." we should escape if it follows a number
    return s;
}

// returns true if should skip newline on entering a container
function skipEnterNl(node) {
    const p = node.parent;
    if (p.type === 'item' || p.type === "block_quote") {
        return true;
    }
    return false;
}

function grandParentIsBlockQuote(node) {
    var gp = node.parent.parent;
    return ((gp !== null) && gp.type === 'block_quote');
}

function mdrender(ast) {
    let buf = '';
    let listItemNoStack = [];

    const walker = ast.walker();
    while ((event = walker.next())) {
        const entering = event.entering;
        const node = event.node;
        const t = node.type;
        if (t === 'document') {
            continue;
        }

        if (t === 'emph') {
            buf += "*";
            continue;
        } else if (t === 'strong') {
            buf += "**";
            continue;
        } else if (t === 'text') {
            buf += esc(node.literal);
            continue;
        } else if (t === 'code') {
            buf += "`" + esc(node.literal) + "`";
            continue;
        } else if (t === 'html_inline') {
            buf + node.literal;
            continue;
        } else if (t == "link" || t === "image") {
            if (entering) {
                if (t === "image") {
                    buf += "!";
                }
                buf += '['
            } else {
                buf += '](' + (node.destination || "");
                if (node.title) {
                    buf += ' "' + node.title + '"';
                }
                buf += ')';
            }
            continue;
        } else if (t === "linebreak") {
            buf = nl(buf);
            continue;
        } else if (t === "softbreak") {
            buf += "\n";
            if (grandParentIsBlockQuote(node)) {
                buf += "> ";
            }
            continue;
        } else if (t === "hardbreak") {
            // https://spec.commonmark.org/0.28/#hard-line-breaks
            buf += "  \n";
        }

        if (entering) {
            if (node.isContainer) {
                if (!skipEnterNl(node)) {
                    buf = nl(buf);
                }
            }
            if (t === "thematic_break") {
                buf += "---";
            } else if (t === "code_block") {
                buf += "```" + (node.info || "") + "\n";
                buf += node.literal;
                buf += "```\n"
            } else if (t === "html_block") {
                buf += node.literal;
                buf += "\n";
            } else if (t === 'list') {
                const start = node.listStart || 1;
                listItemNoStack.push(start);
            } else if (t === "item") {
                const list = node.parent;
                let start = "* ";
                const idx = listItemNoStack.length - 1;
                if (list.listType !== 'bullet') {
                    start = listItemNoStack[idx] + ". ";
                }
                listItemNoStack[idx] = listItemNoStack[idx] + 1;
                buf += start;
            } else if (t === 'heading') {
                for (var i = 0; i < node.level; i++) {
                    buf += "#"
                }
                buf += " ";
            } else if (t === 'block_quote') {
                buf += "> ";
            }
        } else {
            if (t === 'list') {
                listItemNoStack.pop();
            }
            // only happens for containers
            buf = nl(buf);
        }
    }

    return buf;
}

module.exports = mdrender;
