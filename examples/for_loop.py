import os

print(os.listdir())

y=0
while y < 10:
    print("y1 =", y)
    if y == 8:
        break
    y = y+1


x = 0
while x < 20:
    print('x=',x)
    if x % 2 == 0:
        x = x + 3
        continue
    if x % 3 == 0:
        x = x + 5
    x = x + 1

for xx in range(20):
    print("xx=",xx)
    if xx == 8:
        break


