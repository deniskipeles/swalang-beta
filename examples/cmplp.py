# == cmplp.py ==
x=2
print(x)

import os

dir = os.getcwd()
file_path = 'output.txt'
file = open(file_path,'w')
file.write('Hello World >>> AND HIGH')
file.close()

file = open(file_path,'r')  
print(file.read())

# print(os.walk(dir+'/build'))

# print(os.remove('hello'))