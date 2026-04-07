# examples/tuple_test.py
single_line_tuple = (1, 2, 3)
print(single_line_tuple)
print(type(single_line_tuple))

multiline_tuple = (
    1,
    2,
    3,
    4,
    5,
    (
        6,
        7,
        8
    )
)

print(multiline_tuple)
print(type(multiline_tuple))


x=(1,2,3)
y=(1,2,3)
print(x,y)
print(y.count(2))

x = multiline_tuple
for i in x:
    print(i)

