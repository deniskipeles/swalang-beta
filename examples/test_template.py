import template
import os

print("===========================================")
print("🎨 Testing Swalang Template Engine")
print("===========================================")

# ------------------------------------------------------------------------------
# 1. Setup Mock Template Loader (DictLoader)
# ------------------------------------------------------------------------------
print("1. Initializing Template Environment...")

template_store = {
    "base.html": r"""<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>{% block title %}Default Portal{% endblock %}</title>
</head>
<body style="font-family: sans-serif; background: #fafafa; padding: 20px;">
    <header style="border-bottom: 2px solid #333; padding-bottom: 10px;">
        <h2>Swalang Web Portal</h2>
    </header>
    <main style="margin: 20px 0;">
        {% block content %}{% endblock %}
    </main>
    <footer style="border-t: 1px solid #ccc; padding-top: 10px; font-size: 12px; color: #666;">
        &copy; 2026 Swa Foundation. Built with Swalang.
    </footer>
</body>
</html>""",

    "dashboard.html": r"""{% extends "base.html" %}

{% block title %}Meneja wa Mradi — {{ jina | title }}{% endblock %}

{% block content %}
    <h3>Habari, {{ jina | title }}!</h3>
    <p>Hali ya mfumo: <strong>{{ hali | upper | default("HAIJULIKANI") }}</strong></p>

    <h4>Orodha ya Majukumu:</h4>
    <ul>
    {% for jukumu in majukumu %}
        <li>
            Kazi {{ loop.index }}: {{ jukumu | strip | title }}
            {% if loop.first %} <span style="color: green;">(Kipaumbele)</span>{% endif %}
            {% if loop.last %} <span style="color: blue;">(Mwisho)</span>{% endif %}
        </li>
    {% else %}
        <li>Hakuna majukumu kwa sasa.</li>
    {% endfor %}
    </ul>

    {# Test variable assignments #}
    {% set kizingiti = 100 %}
    {% set alama = 120 %}
    <p>Matokeo: {% if alama >= kizingiti %}<strong>Umepita!</strong>{% else %}Umeshindwa.{% endif %}</p>

    {# Test raw escaping #}
    <h4>Testing Security Escaping:</h4>
    <p>Escaped Value: {{ usalama }}</p>
    <p>Safe Value: {{ usalama | safe }}</p>

    {# Test raw block literal #}
    <h4>Jinja2 Syntax Example:</h4>
    <pre><code>{% raw %}
        {% for item in items %}
            <li>{{ item }}</li>
        {% endfor %}
    {% endraw %}</code></pre>
{% endblock %}"""
}

# Instantiate the environment
loader = template.DictLoader(template_store)
env = template.Environment(loader=loader, auto_escape=True, undefined_errors=True)

# ------------------------------------------------------------------------------
# 2. Render Template
# ------------------------------------------------------------------------------
print("\n2. Rendering 'dashboard.html' with dynamic context...")

context = {
    "jina": "amani kemboi",
    "hali": "salama",
    "majukumu": [
        "  kujenga parser ya swalang  ",
        "kuandaa vipimo (unit tests)",
        "  kupeleka sandbox production  "
    ],
    "usalama": "<script>alert('dangerous xss!')</script>"
}

try:
    tpl = env.get_template("dashboard.html")
    output = tpl.render(context)
    print("\n👉 Rendered Output:")
    print("----------------------------------------------------------------------")
    print(output)
    print("----------------------------------------------------------------------")
    
    # Assertions to verify correct parsing/rendering
    assert "Meneja Wa Mradi — Amani Kemboi" in output, "Title casing filter failed"
    assert "SALAMA" in output, "Upper filter failed"
    assert "Kazi 1: Kujenga Parser Ya Swalang" in output, "Combined filter + loop indexing failed"
    assert "&lt;script&gt;" in output, "Auto-escaping failed"
    assert "<script>alert" in output, "Raw/safe filter failed"
    print("✅ All basic rendering assertions passed successfully.")
except Exception as e:
    print(format_str("❌ Render failed: {e}"))
    raise e

# ------------------------------------------------------------------------------
# 3. Test Loop Else block
# ------------------------------------------------------------------------------
print("\n3. Testing loop {% else %} block with empty array...")

context_empty = {
    "jina": "msanidi",
    "majukumu": []
}

try:
    output_empty = tpl.render(context_empty)
    assert "Hakuna majukumu kwa sasa." in output_empty, "Loop else block failed"
    print("✅ Loop else assertion passed successfully.")
except Exception as e:
    print(format_str("❌ Empty loop test failed: {e}"))
    raise e

# ------------------------------------------------------------------------------
# 4. Test Filters and Global Functions
# ------------------------------------------------------------------------------
print("\n4. Testing built-in filters & global functions in isolation...")

# Test simple strings
s_out = template.render_string("Join: {{ items | join(', ') }}", items=["moja", "mbili", "tatu"])
assert s_out == "Join: moja, mbili, tatu", "Join filter failed"

# Test truncation
t_out = template.render_string("Truncate: {{ text | truncate(15) }}", text="Kupanga ni kuchagua siku zote")
assert t_out == "Truncate: Kupanga ni ku...", "Truncation filter failed"

# Test mathematical globals
math_out = template.render_string("Max: {{ max(10, 50) }} | Min: {{ min(10, 50) }}", {})
assert math_out == "Max: 50 | Min: 10", "Math globals failed"

print("✅ Filter and global assertions passed successfully.")
print("\n🎉 All template engine verification tests completed successfully!")