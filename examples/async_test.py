# examples/async_test.py
import http
import asyncio as aio  # Import the new module
import time # Pylearn's standard (blocking) time

tasks = [i for i in range(3)]
print(tasks)

urls = [
    "https://jsonplaceholder.typicode.com/todos/5",
    "https://jsonplaceholder.typicode.com/todos/6",
    "https://jsonplaceholder.typicode.com/todos/7"
]

# Basic usage
for char in enumerate("abc"):
    print(char)  # 0 a, 1 b, 2 c

# With start parameter  
for i, item in enumerate(["x", "y", "z"], start=1):
    print(i, item)  # 1 x, 2 y, 3 z

async def simple_task(id, delay):
    print(format_str("Task {id} starting, will sleep for {delay}s"))
    await aio.sleep(delay) # Use the cooperative sleep
    result = format_str("Task {id} finished after {delay}s at {time.time()}")
    print(result)
    return result

async def fetch_url_task(url):
    print(format_str("Fetching {url}..."))
    response = await http.async_get(url) # Clean call using standard http module
    if response and response.status_code == 200:
        return {"url": url, "title": response.json().get("title", "N/A")}
    else:
        status = "N/A"
        if response: 
            status = response.status_code
        print(format_str("Failed to fetch {url}, status: {status}"))
        return {"url": url, "error": format_str("Failed with status {status}")}

async def main_program():
    print(format_str("Main program started at {time.time()}"))

    # Test aio.sleep
    await aio.sleep(0.5)
    print(format_str("Resumed after 0.5s sleep at {time.time()}"))

    # Test aio.gather with simple tasks
    print("\nGathering simple sleep tasks...")
    task1_promise = simple_task("A", 1.0)
    task2_promise = simple_task("B", 0.5)
    
    results = await aio.gather(task1_promise, task2_promise)
    print("Simple tasks gathered. Results:")
    for res in results:
        print(format_str("- {res}"))

    # Test aio.gather with HTTP fetch tasks
    print("\nGathering HTTP fetch tasks...")
    fetch_promise1 = fetch_url_task("https://jsonplaceholder.typicode.com/todos/5")
    fetch_promise2 = fetch_url_task("https://jsonplaceholder.typicode.com/todos/6")
    fetch_promise3 = fetch_url_task("https://jsonplaceholder.typicode.com/todos/7")

    http_results = await aio.gather(fetch_promise1, fetch_promise2, fetch_promise3)
    print("HTTP tasks gathered. Results:")
    for res in http_results:
        print(format_str("- {res}")) # These will be Pylearn Dicts

    return "Main program completed."

aio.run(main_program())