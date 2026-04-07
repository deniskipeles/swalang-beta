# examples/time_test.py

import time
import os

print(f("Current Pylearn time: {time.time()}"))

print("Sleeping for 1.5 seconds...")
start_time = time.time()
time.sleep(1.5) # Pylearn time.sleep
end_time = time.time()
print(f("Slept for approximately {end_time - start_time} seconds."))

print(os.listdir()) # Just to see some other output




