# examples/sw_template_test.py
"""
pyjinja.py - A Jinja2-inspired templating engine for Pylearn.
Supports {{ expressions }}, {% control blocks %}, and {# comments #}.
"""

# Standard library imports
import re
import sys

# ==============================================================================
# Configuration & Constants
# ==============================================================================

# Default delimiters
DELIMITERS = {
    "variable": ("{{", "}}"),
    "block": ("{%", "%}"),
    "comment": ("{#", "#}")
}

# Compile regex for tokenizing
def _make_lexer_regex():
    var_start, var_end = DELIMITERS["variable"]
    block_start, block_end = DELIMITERS["block"]
    comment_start, comment_end = DELIMITERS["comment"]

    # Escape delimiters for regex
    def esc(s):
        return re.escape(s)

    # --- FIX 1: Converted the complex f-string to a regular string ---
    # This avoids multi-line and verbose regex issues with the custom parser.
    pattern = '(?P<COMMENT>' + esc(comment_start) + '.*?' + esc(comment_end) + ')|(?P<VAR>' + esc(var_start) + '(.*?)' + esc(var_end) + ')|(?P<BLOCK>' + esc(block_start) + '(.*?)' + esc(block_end) + ')|(?P<TEXT>.+?)(?=' + esc(var_start) + '|' + esc(block_start) + '|' + esc(comment_start) + '|$)'
    
    # re.DOTALL | re.VERBOSE in Python are integer flags. 4 | 64 = 68
    # We will assume your `re` module ignores flags for now, as the pattern works without VERBOSE.
    return re.compile(pattern)

_TOKEN_REGEX = _make_lexer_regex()

# ==============================================================================
# Node Base Classes (AST)
# ==============================================================================

class Node:
    def render(self, context):
        raise NotImplementedError()

class TextNode(Node):
    def __init__(self, text):
        self.text = text
    def render(self, context):
        return self.text

class ExpressionNode(Node):
    def __init__(self, expr):
        self.expr = expr.strip()
    def render(self, context):
        try:
            result = eval_with_context(self.expr, context)
            return str(result)
        except Exception as e:
            return f"[[ERROR: {e}]]"

class IfNode(Node):
    def __init__(self, condition, body, else_body=None):
        self.condition = condition.strip()
        self.body = body
        self.else_body = else_body or []
    def render(self, context):
        try:
            if eval_with_context(self.condition, context):
                # --- FIX 2: Replaced generator expression with a loop ---
                parts = []
                for node in self.body:
                    parts.append(node.render(context))
                return "".join(parts)
            else:
                # --- FIX 3: Replaced generator expression with a loop ---
                parts = []
                for node in self.else_body:
                    parts.append(node.render(context))
                return "".join(parts)
        except Exception as e:
            return ''

class ForNode(Node):
    def __init__(self, var_name, iterable, body, else_body=None):
        self.var_name = var_name.strip()
        self.iterable = iterable.strip()
        self.body = body
        self.else_body = else_body or []
    def render(self, context):
        try:
            items = eval_with_context(self.iterable, context)
            if not items:
                # --- FIX 4: Replaced generator expression with a loop ---
                parts = []
                for node in self.else_body:
                    parts.append(node.render(context))
                return "".join(parts)

            rendered_parts = []
            for item in items:
                local_ctx = dict(context)
                local_ctx[self.var_name] = item
                # --- FIX 5: Replaced generator expression with a loop ---
                inner_parts = []
                for node in self.body:
                    inner_parts.append(node.render(local_ctx))
                rendered_parts.append("".join(inner_parts))
            return "".join(rendered_parts)
        except Exception as e:
            return ''

# ==============================================================================
# Helper: Safe Evaluation with Context
# ==============================================================================

def eval_with_context(expr, context):
    allowed_names = {
        "__builtins__": {},
        "True": True,
        "False": False,
        "None": None,
    }
    allowed_names.update(context)
    return eval(expr, allowed_names, {})

# ==============================================================================
# Parser
# ==============================================================================

class Parser:
    def __init__(self, source):
        self.source = source
        self.tokens = list(_TOKEN_REGEX.finditer(source))
        self.pos = 0

    def parse(self):
        nodes = []
        while self.pos < len(self.tokens):
            token = self.tokens[self.pos]
            if token.lastgroup == "TEXT":
                nodes.append(TextNode(token.group()))
            elif token.lastgroup == "VAR":
                nodes.append(ExpressionNode(token.group(2)))
            elif token.lastgroup == "BLOCK":
                block_content = token.group(2)
                nodes.append(self.parse_block(block_content))
            self.pos += 1
        return nodes

    def parse_block(self, content):
        parts = content.split(None, 1)
        if not parts:
            return TextNode('')
        cmd = parts[0]
        args = parts[1].strip() if len(parts) > 1 else ''

        if cmd == "if":
            body, else_body, end_pos = self.parse_conditional()
            self.pos = end_pos
            return IfNode(args, body, else_body)
        elif cmd == "for":
            var_name = args.split()[0] if args else "item"
            iterable = " ".join(args.split()[2:]) if " in " in args else "[]"
            body, else_body, end_pos = self.parse_loop()
            self.pos = end_pos
            return ForNode(var_name, iterable, body, else_body)
        elif cmd == "endif" or cmd == "endfor":
            return TextNode('')
        else:
            return TextNode(f"[[UNKNOWN BLOCK: {cmd}]]")

    def parse_conditional(self):
        body, else_body, nesting = [], [], 1
        self.pos += 1
        current = body
        while self.pos < len(self.tokens):
            token = self.tokens[self.pos]
            if token.lastgroup == "BLOCK":
                txt = token.group(2).strip()
                if txt == "endif":
                    nesting -= 1
                    if nesting == 0:
                        return body, else_body, self.pos
                elif txt == "else" and nesting == 1:
                    current = else_body
                    self.pos += 1
                    continue
                elif txt.startswith("if "):
                    nesting += 1
            node = self.parse_block_node(token)
            current.append(node)
            self.pos += 1
        return body, else_body, self.pos

    def parse_loop(self):
        body, else_body, nesting = [], [], 1
        self.pos += 1
        current = body
        while self.pos < len(self.tokens):
            token = self.tokens[self.pos]
            if token.lastgroup == "BLOCK":
                txt = token.group(2).strip()
                if txt == "endfor":
                    nesting -= 1
                    if nesting == 0:
                        return body, else_body, self.pos
                elif txt == "else" and nesting == 1:
                    current = else_body
                    self.pos += 1
                    continue
                elif txt.startswith("for "):
                    nesting += 1
            node = self.parse_block_node(token)
            current.append(node)
            self.pos += 1
        return body, else_body, self.pos

    def parse_block_node(self, token):
        if token.lastgroup == "TEXT":
            return TextNode(token.group())
        elif token.lastgroup == "VAR":
            return ExpressionNode(token.group(2))
        elif token.lastgroup == "BLOCK":
            return self.parse_block(token.group(2))
        return TextNode('')

# ==============================================================================
# Template Class
# ==============================================================================

class Template:
    def __init__(self, source):
        self.source = source
        self.nodes = Parser(source).parse()

    def render(self, **context):
        # --- FIX 6: Replaced generator expression with a loop ---
        parts = []
        for node in self.nodes:
            parts.append(node.render(context))
        return "".join(parts)

# ==============================================================================
# High-Level Functions
# ==============================================================================

def template_from_string(source):
    return Template(source)

def template_from_file(filepath):
    try:
        # Pylearn doesn't support 'with' yet, so we do it manually.
        f = open(filepath, 'r')
        source = f.read()
        f.close()
        return Template(source)
    except Exception as e:
        raise RuntimeError(f"Could not read template file '{filepath}': {e}")

# ==============================================================================
# Example Usage
# ==============================================================================

if __name__ == "__main__":
    tmpl_src = """
    <h1>Hello, {{ name }}!</h1>
    <ul>
    {% for item in items %}
        <li>{{ item }}</li>
    {% else %}
        <li>No items found.</li>
    {% endfor %}
    </ul>
    {% if user_logged_in %}
        <p>Welcome back!</p>
    {% endif %}
    """
    
    tmpl = template_from_string(tmpl_src)

    result = tmpl.render(
        name="Pylearn",
        items=["Apple", "Banana", "Cherry"],
        user_logged_in=True
    )
    print(result)