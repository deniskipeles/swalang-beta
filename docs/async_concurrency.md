# Async & Concurrency

Swalang provides built-in support for asynchronous programming using the `async` and `await` keywords, similar to Python's `asyncio`.

## Async Functions

Define an asynchronous function using the `async` keyword:

```python
async def my_coroutine():
    print("Start")
    await aio.sleep(1)
    print("End")
```

## Running Async Code

If you define a function named `main_program` as `async`, the Swalang interpreter will automatically boot its asyncio engine and run it.

```python
async def main_program():
    print("Hello from async world!")

# The interpreter will run this automatically if it's async
```

## The `aio` Module

The `aio` module (and sometimes `asyncio` alias) provides tools for cooperative multitasking.

### `aio.sleep(delay)`

Suspends the current coroutine for the specified number of seconds without blocking the entire interpreter thread.

### `aio.gather(*coroutines)`

Runs multiple coroutines concurrently and waits for all of them to finish. It returns a list of the results.

```python
import aio

async def task(id, delay):
    await aio.sleep(delay)
    return f"Task {id} done"

async def main_program():
    results = await aio.gather(task(1, 0.5), task(2, 1.0))
    print(results) # ['Task 1 done', 'Task 2 done']
```

### Concurrent HTTP Requests

Using `asyncio.gather` to perform multiple HTTP requests concurrently:

```python
import aio
import http

async def fetch_url(url):
    print(f"Fetching {url}...")
    response = http.get(url)
    return response.json()

async def main_program():
    urls = [
        "https://jsonplaceholder.typicode.com/todos/1",
        "https://jsonplaceholder.typicode.com/todos/2"
    ]
    tasks = [fetch_url(url) for url in urls]
    results = await aio.gather(*tasks)
    for res in results:
        print(res.get("title"))
```

## Generators and `yield`

Swalang also supports generators using the `yield` keyword.

```python
def count_up_to(n):
    i = 1
    while i <= n:
        yield i
        i += 1

for num in count_up_to(5):
    print(num)
```
