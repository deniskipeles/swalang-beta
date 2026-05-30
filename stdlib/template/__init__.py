"""
template/__init__.py — Production-ready templating engine for Swalang.

Syntax is a strict subset of Jinja2 so templates are portable:

  {{ expression }}           — output an expression (HTML-escaped by default)
  {{ expression | raw }}     — output without escaping
  {# comment #}              — ignored at render time
  {% if expr %}...{% endif %}
  {% elif expr %}
  {% else %}
  {% for item in iterable %}...{% endfor %}
  {% for item in iterable %}...{% else %}...{% endfor %}
  {% set var = expression %}
  {% block name %}...{% endblock %}  (for inheritance)
  {% extends "base.html" %}          (template inheritance)
  {% include "partial.html" %}       (include another template)
  {% raw %}...{% endraw %}           (literal, no processing)

Filters:
  {{ value | upper }}
  {{ value | lower }}
  {{ value | title }}
  {{ value | strip }}
  {{ value | length }}
  {{ value | default("fallback") }}
  {{ value | replace("a","b") }}
  {{ value | truncate(80) }}
  {{ value | escape }}               (always HTML-escape)
  {{ value | raw }}                  (never HTML-escape)
  {{ value | join(", ") }}           (join a list)
  {{ value | first }}
  {{ value | last }}
  {{ value | int }}
  {{ value | float }}
  {{ value | list }}
  {{ value | reverse }}
  {{ value | sort }}
  {{ value | unique }}
  {{ value | keys }}                 (dict keys as list)
  {{ value | values }}               (dict values as list)
  {{ value | items }}                (dict items as list of [k,v])
  {{ value | tojson }}               (serialize to JSON string)

Tests  (used in {% if %} expressions):
  defined(x)     — x is not None
  undefined(x)   — x is None
  none(x)        — x is None
  string(x)      — isinstance str
  number(x)      — isinstance int or float
  iterable(x)    — has __iter__ or is list/dict/str
  mapping(x)     — isinstance dict
  sequence(x)    — isinstance list or str

Global functions available in templates:
  range(start, stop, step=1)
  len(x)
  str(x)
  int(x)
  float(x)
  bool(x)
  list(x)
  dict(**kw)
  max(a, b)
  min(a, b)
  abs(x)
  round(x, n=0)
  zip(a, b)
  enumerate(seq, start=0)
  tojson(x)      — JSON-serialize (uses cjson if available)
"""

import pcre2
import os
import cjson

# ==============================================================================
#  HTML Escaping
# ==============================================================================

_HTML_ESCAPE_TABLE = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;',
}


def escape_html(value):
    """Escape a string for safe insertion into HTML."""
    if not isinstance(value, str):
        value = _to_str(value)
    result = ""
    for ch in value:
        result = result + _HTML_ESCAPE_TABLE.get(ch, ch)
    return result


class _SafeString:
    """A string that will not be auto-escaped by the template engine."""
    def __init__(self, s):
        self._s = s if isinstance(s, str) else str(s)

    def __str__(self):
        return self._s

    def __add__(self, other):
        return _SafeString(self._s + str(other))


def Markup(s):
    """Mark a string as safe HTML (will not be escaped)."""
    return _SafeString(s)

# ==============================================================================
#  Type helpers
# ==============================================================================

def _to_str(val):
    if val is None:
        return ""
    if isinstance(val, bool):
        return "True" if val else "False"
    if isinstance(val, _SafeString):
        return val._s
    return str(val)

# ==============================================================================
#  Filters
# ==============================================================================

def _filter_upper(v, *a):      return _to_str(v).upper()
def _filter_lower(v, *a):      return _to_str(v).lower()
def _filter_title(v, *a):
    s = _to_str(v)
    words = s.split(' ')
    result = []
    for w in words:
        if w:
            result.append(w[0].upper() + w[1:].lower())
        else:
            result.append(w)
    return ' '.join(result)
def _filter_strip(v, *a):      return _to_str(v).strip()
def _filter_length(v, *a):     return len(v) if v is not None else 0
def _filter_escape(v, *a):     return escape_html(_to_str(v))
def _filter_raw(v, *a):        return _SafeString(_to_str(v))
def _filter_int(v, *a):
    try: 
        return int(v)
    except Exception: 
        return 0
def _filter_float(v, *a):
    try: 
        return float(v)
    except Exception: 
        return 0.0
def _filter_list(v, *a):
    if isinstance(v, list): return v
    if isinstance(v, str): return list(v)
    if isinstance(v, dict): return list(v)
    return []
def _filter_reverse(v, *a):
    if isinstance(v, list): return list(reversed(v))
    if isinstance(v, str):
        result = ""
        for ch in v:
            result = ch + result
        return result
    return v
def _filter_sort(v, *a):
    if isinstance(v, list): return sorted(v)
    return v
def _filter_unique(v, *a):
    if isinstance(v, list):
        seen = []
        for x in v:
            if x not in seen:
                seen.append(x)
        return seen
    return v
def _filter_first(v, *a):
    if isinstance(v, (list, str)) and len(v) > 0: return v[0]
    return None
def _filter_last(v, *a):
    if isinstance(v, (list, str)) and len(v) > 0: return v[-1]
    return None
def _filter_keys(v, *a):       return list(v.keys()) if isinstance(v, dict) else []
def _filter_values(v, *a):     return list(v.values()) if isinstance(v, dict) else []
def _filter_items(v, *a):      return list(v.items()) if isinstance(v, dict) else []
def _filter_tojson(v, *a):     return cjson.dumps(v)

def _filter_default(v, fallback="", *a):
    return v if v is not None else fallback

def _filter_replace(v, old="", new="", *a):
    return _to_str(v).replace(old, new)

def _filter_truncate(v, length=255, *a):
    s = _to_str(v)
    if len(s) <= length:
        return s
    return s[:length - 3] + "..."

def _filter_join(v, sep="", *a):
    if isinstance(v, list):
        return sep.join([_to_str(x) for x in v])
    return _to_str(v)


_FILTERS = {
    'upper':    _filter_upper,
    'lower':    _filter_lower,
    'title':    _filter_title,
    'strip':    _filter_strip,
    'length':   _filter_length,
    'count':    _filter_length,
    'escape':   _filter_escape,
    'e':        _filter_escape,
    'raw':      _filter_raw,
    'safe':     _filter_raw,
    'int':      _filter_int,
    'float':    _filter_float,
    'list':     _filter_list,
    'reverse':  _filter_reverse,
    'sort':     _filter_sort,
    'unique':   _filter_unique,
    'first':    _filter_first,
    'last':     _filter_last,
    'keys':     _filter_keys,
    'values':   _filter_values,
    'items':    _filter_items,
    'tojson':   _filter_tojson,
    'default':  _filter_default,
    'd':        _filter_default,
    'replace':  _filter_replace,
    'truncate': _filter_truncate,
    'join':     _filter_join,
}

# ==============================================================================
#  Template tests (used in {% if %} conditions: is defined, is string, ...)
# ==============================================================================

_TESTS = {
    'defined':   (lambda v: v is not None),
    'undefined': (lambda v: v is None),
    'none':      (lambda v: v is None),
    'string':    (lambda v: isinstance(v, str)),
    'number':    (lambda v: isinstance(v, (int, float)) and not isinstance(v, bool)),
    'integer':   (lambda v: isinstance(v, int) and not isinstance(v, bool)),
    'float':     (lambda v: isinstance(v, float)),
    'boolean':   (lambda v: isinstance(v, bool)),
    'iterable':  (lambda v: isinstance(v, (list, dict, str, tuple))),
    'mapping':   (lambda v: isinstance(v, dict)),
    'sequence':  (lambda v: isinstance(v, (list, str, tuple))),
    'odd':       (lambda v: isinstance(v, int) and v % 2 != 0),
    'even':      (lambda v: isinstance(v, int) and v % 2 == 0),
    'divisibleby': (lambda v, n: isinstance(v, int) and v % n == 0),
}

# ==============================================================================
#  Template global functions
# ==============================================================================

def _tojson(x):
    return cjson.dumps(x)

_GLOBALS = {
    'range':     range,
    'len':       len,
    'str':       str,
    'int':       int,
    'float':     float,
    'bool':      bool,
    'list':      list,
    'dict':      dict,
    'max':       max,
    'min':       min,
    'abs':       abs,
    'round':     round,
    'enumerate': enumerate,
    'tojson':    _tojson,
    'escape':    escape_html,
    'Markup':    Markup,
    'True':      True,
    'False':     False,
    'None':      None,
}

# ==============================================================================
#  Expression Evaluator
# ==============================================================================

class _EvalError(Exception):
    pass


def _rfind(s, sub):
    """Find the last index of sub in s (mimics rfind)."""
    i = len(s) - len(sub)
    while i >= 0:
        if s[i:i+len(sub)] == sub:
            return i
        i = i - 1
    return -1


def _eval_expr(expr_str, ctx):
    """
    Evaluate a template expression string against context dict ctx.
    Supports: variable lookup, attribute/index access, filter pipes,
    string/int/float/bool literals, comparison operators, boolean ops,
    function calls (for global functions), and the 'is' test operator.

    Returns the Python value.
    """
    expr_str = expr_str.strip()
    if not expr_str:
        return ""

    pipe_parts = _split_pipes(expr_str)
    if len(pipe_parts) > 1:
        value = _eval_single(pipe_parts[0].strip(), ctx)
        for filter_call in pipe_parts[1:]:
            value = _apply_filter(value, filter_call.strip(), ctx)
        return value

    return _eval_single(expr_str, ctx)


def _split_pipes(expr):
    """Split expression on | that are not inside () [] '' \"\"."""
    parts  = []
    depth  = 0
    in_sq  = False   # single-quote string
    in_dq  = False   # double-quote string
    cur    = ""

    for ch in expr:
        if ch == "'" and not in_dq:
            in_sq = not in_sq
        elif ch == '"' and not in_sq:
            in_dq = not in_dq
        elif not in_sq and not in_dq:
            if ch in '([':
                depth = depth + 1
            elif ch in ')]':
                depth = depth - 1
            elif ch == '|' and depth == 0:
                parts.append(cur)
                cur = ""
                continue
        cur = cur + ch

    parts.append(cur)
    return parts


def _apply_filter(value, filter_expr, ctx):
    """Apply a single filter expression (possibly with args) to value."""
    filter_expr = filter_expr.strip()

    if '(' in filter_expr:
        idx = filter_expr.find('(')
        fname = filter_expr[:idx].strip()
        args_str = filter_expr[idx + 1:].rstrip(')')
        args = _parse_filter_args(args_str, ctx)
    else:
        fname = filter_expr
        args  = []

    fn = _FILTERS.get(fname)
    if fn is None:
        raise _EvalError(format_str("Unknown filter: '{fname}'"))
    return fn(value, *args)


def _parse_filter_args(args_str, ctx):
    """Parse comma-separated filter arguments."""
    args_str = args_str.strip()
    if not args_str:
        return []
    parts = []
    current = ""
    depth = 0
    in_sq = False
    in_dq = False
    for ch in args_str:
        if ch == "'" and not in_dq:
            in_sq = not in_sq
        elif ch == '"' and not in_sq:
            in_dq = not in_dq
        elif not in_sq and not in_dq:
            if ch in '([':
                depth = depth + 1
            elif ch in ')]':
                depth = depth - 1
            elif ch == ',' and depth == 0:
                parts.append(_eval_single(current.strip(), ctx))
                current = ""
                continue
        current = current + ch
    if current.strip():
        parts.append(_eval_single(current.strip(), ctx))
    return parts


def _eval_single(expr, ctx):
    """Evaluate a single expression (no pipes)."""
    expr = expr.strip()

    if not expr:
        return ""

    if expr == 'True':  return True
    if expr == 'False': return False
    if expr == 'None':  return None

    if (expr.startswith('"') and expr.endswith('"')) or (expr.startswith("'") and expr.endswith("'")):
        return expr[1:-1]

    try:
        return int(expr)
    except (ValueError, TypeError):
        pass

    try:
        return float(expr)
    except (ValueError, TypeError):
        pass

    if expr.startswith('not '):
        return not _eval_single(expr[4:].strip(), ctx)

    is_match = pcre2.match(r'^(.+?)\s+is\s+(not\s+)?(\w+)(?:\((.+)\))?$', expr)
    if is_match:
        lhs      = _eval_single(is_match.group(1), ctx)
        negated  = is_match.group(2) is not None
        testname = is_match.group(3)
        test_arg_str = is_match.group(4)
        is_match.free()
        test_fn = _TESTS.get(testname)
        if test_fn is None:
            raise _EvalError(format_str("Unknown test: '{testname}'"))
        if test_arg_str:
            test_arg = _eval_single(test_arg_str, ctx)
            result   = test_fn(lhs, test_arg)
        else:
            result = test_fn(lhs)
        return not result if negated else result

    in_match = pcre2.match(r'^(.+?)\s+(?:not\s+)?in\s+(.+)$', expr)
    if in_match:
        lhs = _eval_single(in_match.group(1), ctx)
        rhs = _eval_single(in_match.group(2), ctx)
        negated_in = ' not in ' in expr
        in_match.free()
        if isinstance(rhs, (list, str, dict)):
            result = lhs in rhs
        else:
            result = False
        return not result if negated_in else result

    for op in [' and ', ' or ', ' == ', ' != ', ' >= ', ' <= ', ' > ', ' < ']:
        idx = _find_op(expr, op)
        if idx != -1:
            lhs = _eval_single(expr[:idx].strip(), ctx)
            rhs = _eval_single(expr[idx + len(op):].strip(), ctx)
            if op == ' and ':  return lhs and rhs
            if op == ' or ':   return lhs or rhs
            if op == ' == ':   return lhs == rhs
            if op == ' != ':   return lhs != rhs
            if op == ' >= ':   return lhs >= rhs
            if op == ' <= ':   return lhs <= rhs
            if op == ' > ':    return lhs > rhs
            if op == ' < ':    return lhs < rhs

    tern = pcre2.match(r'^(.+?)\s+if\s+(.+?)\s+else\s+(.+)$', expr)
    if tern:
        true_val  = _eval_single(tern.group(1), ctx)
        cond      = _eval_single(tern.group(2), ctx)
        false_val = _eval_single(tern.group(3), ctx)
        tern.free()
        return true_val if cond else false_val

    if expr.startswith('[') and expr.endswith(']'):
        inner = expr[1:-1].strip()
        if not inner:
            return []
        items = _split_top_level(inner, ',')
        return [_eval_single(item.strip(), ctx) for item in items]

    if expr.startswith('{') and expr.endswith('}'):
        inner = expr[1:-1].strip()
        if not inner:
            return {}
        result = {}
        pairs = _split_top_level(inner, ',')
        for pair in pairs:
            colon = pair.find(':')
            if colon != -1:
                k = _eval_single(pair[:colon].strip(), ctx)
                v = _eval_single(pair[colon + 1:].strip(), ctx)
                result[k] = v
        return result

    func_match = pcre2.match(r'^([a-zA-Z_]\w*)\((.*)?\)$', expr)
    if func_match:
        fname   = func_match.group(1)
        args_s  = func_match.group(2) or ""
        func_match.free()
        fn = _GLOBALS.get(fname) or ctx.get(fname)
        if fn and callable(fn):
            args = _parse_filter_args(args_s, ctx) if args_s.strip() else []
            return fn(*args)

    if '[' in expr and expr.endswith(']'):
        bracket = _rfind(expr, '[')
        obj_part = expr[:bracket]
        key_part = expr[bracket + 1:-1]
        obj = _eval_single(obj_part, ctx)
        key = _eval_single(key_part, ctx)
        try:
            return obj[key]
        except (KeyError, IndexError, TypeError):
            return None

    if '.' in expr:
        parts = expr.split('.')
        obj   = _lookup(parts[0], ctx)
        for attr in parts[1:]:
            if obj is None:
                return None
            if isinstance(obj, dict):
                obj = obj.get(attr)
            else:
                try:
                    obj = getattr(obj, attr, None)
                except Exception:
                    obj = None
        return obj

    return _lookup(expr, ctx)


def _find_op(expr, op):
    """Find op in expr, ignoring occurrences inside strings or parens."""
    depth  = 0
    in_sq  = False
    in_dq  = False
    i      = 0
    op_len = len(op)

    while i < len(expr):
        ch = expr[i]
        if ch == "'" and not in_dq: in_sq = not in_sq
        elif ch == '"' and not in_sq: in_dq = not in_dq
        elif not in_sq and not in_dq:
            if ch in '([': depth = depth + 1
            elif ch in ')]': depth = depth - 1
            elif depth == 0 and expr[i:i + op_len] == op:
                return i
        i = i + 1
    return -1


def _split_top_level(s, delim):
    """Split s by delim at the top level (not inside brackets/strings)."""
    parts  = []
    depth  = 0
    in_sq  = False
    in_dq  = False
    cur    = ""
    for ch in s:
        if ch == "'" and not in_dq: in_sq = not in_sq
        elif ch == '"' and not in_sq: in_dq = not in_dq
        elif not in_sq and not in_dq:
            if ch in '([{': depth = depth + 1
            elif ch in ')]}': depth = depth - 1
            elif ch == delim and depth == 0:
                parts.append(cur)
                cur = ""
                continue
        cur = cur + ch
    parts.append(cur)
    return parts


def _lookup(name, ctx):
    """Look up name in ctx, then in _GLOBALS."""
    if name in ctx:
        return ctx[name]
    if name in _GLOBALS:
        return _GLOBALS[name]
    return None

# ==============================================================================
#  Tokenizer
# ==============================================================================

_TOK_TEXT     = 'text'
_TOK_EXPR     = 'expr'       # {{ ... }}
_TOK_BLOCK    = 'block'      # {% ... %}
_TOK_COMMENT  = 'comment'    # {# ... #}

_TOKEN_RE  = pcre2.compile(r'\{%[-]?\s*(.*?)\s*[-]?%\}|\{\{(.*?)\}\}|\{#(.*?)#\}', pcre2.DOTALL)
_BLOCK_RE  = pcre2.compile(r'^(\w+)(.*)?$')


def _tokenize(source):
    """
    Return a list of (type, content, line_offset) tuples for all tokens in source.
    Text between tags is yielded as _TOK_TEXT.
    """
    tokens = []
    pos   = 0
    source_bytes = source.encode('utf-8')
    for m in _TOKEN_RE.finditer(source):
        start, end = m.span()
        if start > pos:
            tokens.append((_TOK_TEXT, source_bytes[pos:start].decode('utf-8')))
        m.free()

        raw_bytes = source_bytes[start:end]
        raw = raw_bytes.decode('utf-8')
        if raw.startswith('{%'):
            inner = raw[2:-2].strip().lstrip('-').rstrip('-').strip()
            tokens.append((_TOK_BLOCK, inner))
        elif raw.startswith('{{'):
            inner = raw[2:-2].strip()
            tokens.append((_TOK_EXPR, inner))
        elif raw.startswith('{#'):
            inner = raw[2:-2].strip()
            tokens.append((_TOK_COMMENT, inner))
        pos = end

    if pos < len(source_bytes):
        tokens.append((_TOK_TEXT, source_bytes[pos:].decode('utf-8')))
    return tokens

# ==============================================================================
#  AST Nodes
# ==============================================================================

class _TextNode:
    def __init__(self, text):  self.text = text

class _ExprNode:
    def __init__(self, expr):  self.expr = expr

class _CommentNode:
    pass

class _IfNode:
    def __init__(self, branches, else_body):
        self.branches  = branches
        self.else_body = else_body   # list of nodes or None

class _ForNode:
    def __init__(self, var, iter_expr, body, else_body):
        self.var       = var         # "item" or "key, value"
        self.iter_expr = iter_expr
        self.body      = body        # list of nodes
        self.else_body = else_body   # list of nodes or None

class _SetNode:
    def __init__(self, name, expr):
        self.name = name
        self.expr = expr

class _BlockNode:
    def __init__(self, name, body):
        self.name = name
        self.body = body   # list of nodes

class _ExtendsNode:
    def __init__(self, parent):
        self.parent = parent   # template name string

class _IncludeNode:
    def __init__(self, name):
        self.name = name

class _RawNode:
    def __init__(self, text):
        self.text = text

# ==============================================================================
#  Parser
# ==============================================================================

class TemplateError(Exception):
    """Raised for syntax and runtime errors in templates."""
    pass


class _Parser:
    def __init__(self, tokens):
        self._tokens = tokens
        self._pos    = 0

    def _peek(self):
        if self._pos < len(self._tokens):
            return self._tokens[self._pos]
        return None

    def _consume(self):
        tok = self._tokens[self._pos]
        self._pos = self._pos + 1
        return tok

    def parse_body(self, stop_tags=None):
        """
        Parse tokens into a list of nodes, stopping when a block whose
        tag word is in stop_tags is encountered.
        Returns (nodes, tag_that_stopped) or (nodes, None) at end of input.
        """
        nodes    = []
        stop_tags = stop_tags or []

        while self._pos < len(self._tokens):
            tok = self._peek()
            if tok is None:
                break

            typ, content = tok[0], tok[1]

            if typ == _TOK_TEXT:
                self._consume()
                nodes.append(_TextNode(content))

            elif typ == _TOK_COMMENT:
                self._consume()
                nodes.append(_CommentNode())

            elif typ == _TOK_EXPR:
                self._consume()
                nodes.append(_ExprNode(content))

            elif typ == _TOK_BLOCK:
                self._consume()
                tag_word = content.split()[0] if content.split() else ""

                if stop_tags and tag_word in stop_tags:
                    return (nodes, tag_word, content)

                if tag_word == 'if':
                    nodes.append(self._parse_if(content))
                elif tag_word == 'for':
                    nodes.append(self._parse_for(content))
                elif tag_word == 'set':
                    nodes.append(self._parse_set(content))
                elif tag_word == 'block':
                    nodes.append(self._parse_block(content))
                elif tag_word == 'extends':
                    nodes.append(self._parse_extends(content))
                elif tag_word == 'include':
                    nodes.append(self._parse_include(content))
                elif tag_word == 'raw':
                    nodes.append(self._parse_raw())
                else:
                    pass

        return (nodes, None, "")

    def _parse_if(self, content):
        condition = content[2:].strip()
        branches  = [(condition, None)]
        final_else = None

        while True:
            body, stopper, stop_content = self.parse_body(['elif', 'else', 'endif'])

            if branches[-1][1] is None:
                branches[-1] = (branches[-1][0], body)

            if stopper == 'endif' or stopper is None:
                break
            elif stopper == 'else':
                else_body, _, _ = self.parse_body(['endif'])
                final_else = else_body
                break
            elif stopper == 'elif':
                elif_cond = stop_content[4:].strip()
                branches.append((elif_cond, None))

        return _IfNode(branches, final_else)

    def _parse_for(self, content):
        m = pcre2.match(r'^for\s+(.+?)\s+in\s+(.+)$', content)
        if not m:
            raise TemplateError(format_str("Invalid for tag: '{content}'"))
        var_part  = m.group(1).strip()
        iter_expr = m.group(2).strip()
        m.free()

        body, stopper, _ = self.parse_body(['else', 'endfor'])
        else_body = None
        if stopper == 'else':
            else_body, _, _ = self.parse_body(['endfor'])

        return _ForNode(var_part, iter_expr, body, else_body)

    def _parse_set(self, content):
        m = pcre2.match(r'^set\s+([a-zA-Z_]\w*)\s*=\s*(.+)$', content)
        if not m:
            raise TemplateError(format_str("Invalid set tag: '{content}'"))
        name = m.group(1)
        expr = m.group(2).strip()
        m.free()
        return _SetNode(name, expr)

    def _parse_block(self, content):
        m = pcre2.match(r'^block\s+(\w+)', content)
        if not m:
            raise TemplateError(format_str("Invalid block tag: '{content}'"))
        name = m.group(1)
        m.free()
        body, _, _ = self.parse_body(['endblock'])
        return _BlockNode(name, body)

    def _parse_extends(self, content):
        m = pcre2.match(r'''extends\s+['"](.+?)['"]''', content)
        if not m:
            raise TemplateError(format_str("Invalid extends tag: '{content}'"))
        parent = m.group(1)
        m.free()
        return _ExtendsNode(parent)

    def _parse_include(self, content):
        m = pcre2.match(r'''include\s+['"](.+?)['"]''', content)
        if not m:
            raise TemplateError(format_str("Invalid include tag: '{content}'"))
        name = m.group(1)
        m.free()
        return _IncludeNode(name)

    def _parse_raw(self):
        text = ""
        while self._pos < len(self._tokens):
            tok = self._consume()
            typ, content = tok[0], tok[1]
            if typ == _TOK_BLOCK and content.strip() == 'endraw':
                break
            if typ == _TOK_TEXT:
                text = text + content
            elif typ == _TOK_EXPR:
                text = text + '{{' + content + '}}'
            elif typ == _TOK_BLOCK:
                text = text + '{%' + content + '%}'
            elif typ == _TOK_COMMENT:
                text = text + '{#' + content + '#}'
        return _RawNode(text)


def _parse(source):
    tokens = _tokenize(source)
    parser = _Parser(tokens)
    nodes, _, _ = parser.parse_body()
    return nodes

# ==============================================================================
#  Renderer
# ==============================================================================

class _Renderer:
    def __init__(self, env, auto_escape=True):
        self._env        = env
        self._auto_escape = auto_escape

    def render_nodes(self, nodes, ctx):
        parts = []
        for node in nodes:
            parts.append(self._render_node(node, ctx))
        return "".join(parts)

    def _render_node(self, node, ctx):
        if isinstance(node, _TextNode):
            return node.text

        if isinstance(node, _CommentNode):
            return ""

        if isinstance(node, _RawNode):
            return node.text

        if isinstance(node, _ExprNode):
            try:
                value = _eval_expr(node.expr, ctx)
            except Exception as e:
                if self._env._undefined_errors:
                    raise TemplateError(format_str("Error evaluating '{{{{ {node.expr} }}}}': {e}"))
                value = ""
            return self._output(value)

        if isinstance(node, _IfNode):
            return self._render_if(node, ctx)

        if isinstance(node, _ForNode):
            return self._render_for(node, ctx)

        if isinstance(node, _SetNode):
            try:
                ctx[node.name] = _eval_expr(node.expr, ctx)
            except Exception:
                ctx[node.name] = None
            return ""

        if isinstance(node, _BlockNode):
            override = ctx.get('__blocks__', {}).get(node.name)
            body     = override if override is not None else node.body
            return self.render_nodes(body, ctx)

        if isinstance(node, _IncludeNode):
            return self._render_include(node, ctx)

        if isinstance(node, _ExtendsNode):
            return ""

        return ""

    def _output(self, value):
        if isinstance(value, _SafeString):
            return value._s
        s = _to_str(value)
        if self._auto_escape:
            return escape_html(s)
        return s

    def _render_if(self, node, ctx):
        for cond, body in node.branches:
            try:
                result = _eval_expr(cond, ctx)
            except Exception:
                result = False
            if result:
                return self.render_nodes(body, ctx)
        if node.else_body is not None:
            return self.render_nodes(node.else_body, ctx)
        return ""

    def _render_for(self, node, ctx):
        try:
            iterable = _eval_expr(node.iter_expr, ctx)
        except Exception:
            iterable = []

        if iterable is None:
            iterable = []

        items = list(iterable.items()) if isinstance(iterable, dict) else list(iterable)

        if not items:
            if node.else_body is not None:
                return self.render_nodes(node.else_body, ctx)
            return ""

        parts   = []
        n_items = len(items)
        var_part = node.var

        is_tuple = ',' in var_part
        if is_tuple:
            vnames = [v.strip() for v in var_part.split(',')]
        else:
            vnames = [var_part]

        for i in range(n_items):
            item = items[i]
            loop_ctx = dict(ctx)

            loop_ctx['loop'] = {
                'index':   i + 1,
                'index0':  i,
                'revindex':  n_items - i,
                'revindex0': n_items - i - 1,
                'first':   i == 0,
                'last':    i == n_items - 1,
                'length':  n_items,
                'depth':   (((ctx.get('loop', {}).get('depth', 0)) + 1) if (isinstance(ctx.get('loop'), dict)) else 1),
            }

            if is_tuple and isinstance(item, (list, tuple)) and len(item) >= len(vnames):
                for j in range(len(vnames)):
                    loop_ctx[vnames[j]] = item[j]
            elif is_tuple and len(vnames) == 2:
                loop_ctx[vnames[0]] = item[0] if isinstance(item, (list, tuple)) else item
                loop_ctx[vnames[1]] = item[1] if isinstance(item, (list, tuple)) else None
            else:
                loop_ctx[var_part] = item

            parts.append(self.render_nodes(node.body, loop_ctx))

        return "".join(parts)

    def _render_include(self, node, ctx):
        if self._env is None:
            return ""
        try:
            tpl = self._env.get_template(node.name)
            return tpl.render(ctx)
        except Exception as e:
            if self._env._undefined_errors:
                raise TemplateError(format_str("Include error '{node.name}': {e}"))
            return ""

# ==============================================================================
#  Template
# ==============================================================================

class Template:
    """
    A compiled template.  Create via Environment.from_string() or
    Environment.get_template(); don't instantiate directly.
    """

    def __init__(self, source, env, name="<string>"):
        self._source = source
        self._env    = env
        self._name   = name
        self._nodes  = _parse(source)

    def render(self, context=None, **kwargs):
        """
        Render the template and return the result as a str.

        render({'key': 'value'})
        render(key='value')
        """
        ctx = {}
        ctx.update(_GLOBALS)
        if context:
            ctx.update(context)
        ctx.update(kwargs)

        extends_node = None
        for node in self._nodes:
            if isinstance(node, _ExtendsNode):
                extends_node = node
                break

        if extends_node:
            return self._render_with_inheritance(extends_node, ctx)

        renderer = _Renderer(self._env, self._env._auto_escape)
        return renderer.render_nodes(self._nodes, ctx)

    def _render_with_inheritance(self, extends_node, ctx):
        """Render this template as a child of extends_node.parent."""
        blocks = {}
        for node in self._nodes:
            if isinstance(node, _BlockNode):
                blocks[node.name] = node.body

        parent_tpl = self._env.get_template(extends_node.parent)

        parent_ctx = dict(ctx)
        parent_ctx['__blocks__'] = blocks

        renderer = _Renderer(self._env, self._env._auto_escape)
        return renderer.render_nodes(parent_tpl._nodes, parent_ctx)

    def stream(self, context=None, **kwargs):
        """Return chunks of the rendered output as a list."""
        ctx = {}
        ctx.update(_GLOBALS)
        if context:
            ctx.update(context)
        ctx.update(kwargs)
        renderer = _Renderer(self._env, self._env._auto_escape)
        results = []
        for node in self._nodes:
            results.append(renderer._render_node(node, ctx))
        return results

    def __repr__(self):
        return format_str("<Template '{self._name}'>")

# ==============================================================================
#  Environment
# ==============================================================================

class Environment:
    """
    The central configuration object.

    env = Environment(
        loader=FileSystemLoader('./templates'),
        auto_escape=True,
        undefined_errors=False,
    )
    tpl = env.get_template('index.html')
    html = tpl.render(title='Hello', items=[1, 2, 3])

    Custom filters:
        env.filters['md5'] = lambda v: some_hash(v)

    Custom globals:
        env.globals['site_name'] = 'My App'
    """

    def __init__(self, loader=None, auto_escape=True, undefined_errors=False, trim_blocks=False, lstrip_blocks=False):
        self._loader           = loader
        self._auto_escape      = auto_escape
        self._undefined_errors = undefined_errors
        self._trim_blocks      = trim_blocks
        self._lstrip_blocks    = lstrip_blocks
        self._cache            = {}

        self.filters = dict(_FILTERS)
        self.globals = dict(_GLOBALS)
        self.tests   = dict(_TESTS)

    def add_filter(self, name, fn):
        """Register a custom filter function."""
        self.filters[name] = fn

    def add_global(self, name, value):
        """Register a template global variable or function."""
        self.globals[name] = value

    def add_test(self, name, fn):
        """Register a custom test function."""
        self.tests[name] = fn

    def get_template(self, name):
        """Load and return a Template by name, using the configured loader."""
        if name in self._cache:
            return self._cache[name]
        if self._loader is None:
            raise TemplateError(format_str("No loader configured (tried to load '{name}')"))
        source = self._loader.get_source(name)
        if self._trim_blocks:
            source = _apply_trim_blocks(source)
        if self._lstrip_blocks:
            source = _apply_lstrip_blocks(source)
        tpl = Template(source, self, name)
        self._cache[name] = tpl
        return tpl

    def from_string(self, source, name="<string>"):
        """Create a Template directly from a source string."""
        return Template(source, self, name)

    def invalidate_cache(self, name=None):
        """
        Clear the template cache.
        If name is given, only that template is evicted.
        """
        if name:
            if name in self._cache:
                del self._cache[name]
        else:
            self._cache = {}


def _apply_trim_blocks(source):
    """Remove the first newline after each block tag."""
    return pcre2.sub(r'(%\})\n', r'\1', source)


def _apply_lstrip_blocks(source):
    """Strip leading whitespace from lines that start a block tag."""
    return pcre2.sub(r'(?m)^[ \t]+(\{%)', r'\1', source)

# ==============================================================================
#  Loaders
# ==============================================================================

class BaseLoader:
    """Abstract base class for template loaders."""

    def get_source(self, name):
        raise NotImplementedError


class FileSystemLoader(BaseLoader):
    """
    Load templates from a directory (or list of directories).

    FileSystemLoader('./templates')
    FileSystemLoader(['./templates', './fallback'])
    """

    def __init__(self, search_path, encoding='utf-8'):
        if isinstance(search_path, str):
            self._paths   = [search_path]
        else:
            self._paths   = list(search_path)
        self._encoding = encoding

    def get_source(self, name):
        for base in self._paths:
            full = os.path.join(base, name)
            if os.path.isfile(full):
                return os.read_text(full)
        raise TemplateError(format_str("Template not found: '{name}'"))

    def list_templates(self):
        """Return a sorted list of all template names in the search paths."""
        names = []
        for base in self._paths:
            for dirpath, _, filenames in os.walk(base):
                for fname in filenames:
                    full = os.path.join(dirpath, fname)
                    rel  = full[len(base):].lstrip('/\\')
                    names.append(rel)
        names.sort()
        return names


class DictLoader(BaseLoader):
    """
    Load templates from a Python dict (useful for testing).

    DictLoader({'base.html': '...', 'child.html': '{% extends "base.html" %}'})
    """

    def __init__(self, mapping):
        self._mapping = mapping

    def get_source(self, name):
        if name not in self._mapping:
            raise TemplateError(format_str("Template not found: '{name}'"))
        return self._mapping[name]


class PrefixLoader(BaseLoader):
    """
    Dispatch to different loaders based on a prefix.

    PrefixLoader({
        'admin': FileSystemLoader('./admin_templates'),
        'user':  FileSystemLoader('./user_templates'),
    })
    # Loads 'admin/dashboard.html' from './admin_templates/dashboard.html'
    """

    def __init__(self, mapping, delimiter='/'):
        self._mapping   = mapping
        self._delimiter = delimiter

    def get_source(self, name):
        idx = name.find(self._delimiter)
        if idx == -1:
            raise TemplateError(format_str("Template name '{name}' has no prefix"))
        prefix = name[:idx]
        rest   = name[idx + 1:]
        loader = self._mapping.get(prefix)
        if loader is None:
            raise TemplateError(format_str("No loader for prefix '{prefix}'"))
        return loader.get_source(rest)


class ChoiceLoader(BaseLoader):
    """
    Try multiple loaders in order, returning the first match.

    ChoiceLoader([FileSystemLoader('./a'), FileSystemLoader('./b')])
    """

    def __init__(self, loaders):
        self._loaders = loaders

    def get_source(self, name):
        for loader in self._loaders:
            try:
                return loader.get_source(name)
            except TemplateError:
                pass
        raise TemplateError(format_str("Template not found in any loader: '{name}'"))

# ==============================================================================
#  Module-level convenience API
# ==============================================================================

_default_env = Environment(auto_escape=False)


def render_string(source, context=None, **kwargs):
    """
    Render a template source string with the given context.

    render_string('Hello {{ name }}!', name='World')
    → 'Hello World!'
    """
    tpl = _default_env.from_string(source)
    return tpl.render(context, **kwargs)


def render_file(path_str, context=None, **kwargs):
    """
    Read a template from disk and render it.
    Uses a throw-away FileSystemLoader; prefer Environment for production.
    """
    source = os.read_text(path_str)
    env    = Environment(auto_escape=True)
    tpl    = env.from_string(source, name=path_str)
    return tpl.render(context, **kwargs)