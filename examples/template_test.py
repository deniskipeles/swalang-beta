# examples/template_test.py

# Import the template module
import template
print("hi")
# Method 1: Using Go template syntax (default)
tmpl = template.new("hello")
tmpl.parse("Hello {{.name}}!")
result = tmpl.execute({"name": "World"})
print(1,result)

# Method 2: Using Jinja2-like syntax
jinja_tmpl = template.new("jinja_hello", "pylearn_custom")
jinja_tmpl.parse("Hello {{ name }}!")
result = jinja_tmpl.execute({"name": "World"})
print(2,result)

# Method 3: Quick render from string
result = template.render("Hello {{ name }}!", {"name": "World"}, "pylearn_custom")
print(3,result)

# Method 4: Load and render from file
result = template.render_file("hello.html", {"name": "World"}, "pylearn_custom", "./examples/templates")
print(4,result)

# Advanced Jinja2-like features:
advanced_template = '''
{% for item in items %}
    <li>{{ item.name|upper }} - {{ item.price|default("N/A", True) }}</li>
{% endfor %}

{% if user.is_admin %}
    <p>Welcome, admin!</p>
{% elif user.is_authenticated %}
    <p>Welcome, {{ user.name }}!</p>
{% else %}
    <p>Please log in.</p>
{% endif %}

'''

data = {
    "items": [
        {"name": "Apple", "price": 1.50},
        {"name": "Banana", "price1": 0.75}
    ],
    "user": {
        "name": "John",
        "is_authenticated": True,
        "is_admin": False
    }
}

result = template.render(advanced_template, data, "pylearn_custom")

print("last",result)
