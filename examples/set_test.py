# examples/set_test.py

z = set({1,2,3,4})
z.add(1)
z.add(2)
z.add(3)
z.add(4)
print(z)
print(type(z))

x = {
    1,
    2,
    3,
    4,
    5,
    6,
    
}
print(x)
print(type(x))  

y = {
    "hello":"there",
    "name":"denis",
    "animals":{
        "dog":"bark",
        "cat":"meow",
        "horse":"neigh",
        "others":[
            [1,2,3],
            [4,5,6]
        ],
    },
    "array":[
        1,
        2,
        3,
        4,
        5,
        6,
    ]
}
print(y)
print(type(y))

