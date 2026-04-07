# examples/test_template.py

import template
import os

async def main_program(): # <-- Make the function async
    """
    This program demonstrates the features of the new Pylearn template engine.
    """
    print("--- Pylearn Template Engine Demonstration ---")
    
    # Simple helper for running tests
    def run_test(name, result, expected):
        # Normalize whitespace for consistent comparison
        norm_result = " ".join(result.strip().split())
        norm_expected = " ".join(expected.strip().split())

        if norm_result == norm_expected:
            print(f"   ✅ Test Passed: {name}")
            
        else:
            print(f"   ❌ TEST FAILED: {name}")
            print(f"        Expected: '{norm_expected}'")
            print(f"        Got:      '{norm_result}'")
    # try:
    #     # --- Setup for file-based tests (include, extends) ---
    #     # Create a temporary directory and template files
    #     os.mkdir("temp_templates")
    #     with open("temp_templates/header.html", "w") as f:
    #         f.write("<h1>{{ title }}</h1>")
        
    #     with open("temp_templates/base.html", "w") as f:
    #         f.write("<!DOCTYPE html><html><head><title>{% block title %}Default Title{% endblock %}</title></head>\n<body><div id='content'>{% block content %}{% endblock %}</div></body></html>")
    # except Exception as e:
    #     print(e)
    # --- 1. Simple Variable Substitution ---
    print("\n[1] Testing simple variable substitution...")
    simple_tpl = "Hello, {{ name }}! Welcome to Pylearn templates."
    context = {"name": "World"}
    output = template.render(simple_tpl, context)
    run_test("Simple Variable", output, "Hello, World! Welcome to Pylearn templates.")
    
    # --- 2. Testing `if/elif/else` blocks ---
    print("\n[2] Testing {% if/elif/else %} blocks...")
    if_else_tpl = """{% if score > 90 %}Grade: A{% elif score > 80 %}Grade: B{% else %}Grade: C{% endif %}"""
    
    context_a = {"score": 95}
    output_a = template.render(if_else_tpl, context_a).strip()
    run_test("If-True Branch (Grade A)", output_a, "Grade: A")

    context_b = {"score": 85}
    output_b = template.render(if_else_tpl, context_b).strip()
    run_test("Elif Branch (Grade B)", output_b, "Grade: B")

    context_c = {"score": 75}
    output_c = template.render(if_else_tpl, context_c).strip()
    run_test("Else Branch (Grade C)", output_c, "Grade: C")

    # --- 3. Testing `for` loop with `range` function ---
    print("\n[3] Testing {% for ... %} loop with range() function...")
    for_tpl = "<ul>{% for i in range(3) %}<li>Item {{ i }}</li>{% endfor %}</ul>"
    output_for = template.render(for_tpl, {})
    run_test("For Loop with Range", output_for, "<ul><li>Item 0</li><li>Item 1</li><li>Item 2</li></ul>")

    # --- 4. Testing Filters ---
    print("\n[4] Testing filters (`| upper`)...")
    filter_tpl = "This is a test: {{ message | upper }}"
    filter_context = {"message": "hello world"}
    output_filter = template.render(filter_tpl, filter_context)
    run_test("Upper Filter", output_filter, "This is a test: HELLO WORLD")

    # --- 5. Testing `raw` block ---
    print("\n[5] Testing {% raw %} block...")
    raw_tpl = "{% raw %}This will not be rendered: {{ name }}{% endraw %}"
    output_raw = template.render(raw_tpl, {"name": "Should not appear"})
    run_test("Raw Block", output_raw, "This will not be rendered: {{ name }}")

    # --- 6. Testing `include` tag ---
    print("\n[6] Testing {% include %} tag...")
    include_tpl = '<div id="header">{% include "temp_templates/header.html" %}</div>'
    include_context = {"title": "My Included Header"}
    output_include = template.render(include_tpl, include_context)
    run_test("Include Tag", output_include, '<div id="header"><h1>My Included Header</h1></div>')
    
    # --- 7. Testing `extends` and `block` tags ---
    print("\n[7] Testing {% extends %} and {% block %} tags...")
    extends_tpl = """
    {% extends "temp_templates/base.html" %}
    {% block title %}My Page Title{% endblock %}
    {% block content %}<p>This is the content of my page.</p>{% endblock %}
    """
    output_extends = template.render(extends_tpl, {})
    expected_extends = """
    <!DOCTYPE html><html><head><title>My Page Title</title></head>
    <body><div id='content'><p>This is the content of my page.</p></div></body></html>
    """
    run_test("Extends and Block Tags", output_extends, expected_extends)

    # --- 8. Testing the Object-Oriented API ---
    print("\n[8] Testing object-oriented API...")
    tpl_string = "User: {{ user.name }} (ID: {{ user.id }})"
    compiled_template = template.from_string(tpl_string)
    user_data = {"user": {"name": "Charlie", "id": 101}}
    oo_output = compiled_template.render(user_data)
    run_test("OO API", oo_output, "User: Charlie (ID: 101)")
    
    # --- 9. Testing error handling ---
    print("\n[9] Testing error handling...")
    try:
        template.from_string("Hello, {{ user.name ")
        run_test("Syntax Error", "no error raised", "TemplateError")
    except template.TemplateError as e:
        print(f"   Successfully caught expected syntax error: {e}")
        run_test("Syntax Error", "TemplateError raised", "TemplateError raised")

    output_missing_var = template.render("Hello, {{ missing_variable }}", {})
    run_test("Missing Variable", output_missing_var, "Hello, ")

    # --- Cleanup ---
    # os.remove("temp_templates/header.html")
    # os.remove("temp_templates/base.html")
    # os.rmdir("temp_templates")

    print("\n--- All template demonstrations complete. ---")