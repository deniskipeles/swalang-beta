# Pylearn script
# from concurrent.futures import ThreadPoolExecutor
import concurrent
ThreadPoolExecutor = concurrent.futures.ThreadPoolExecutor
import time # Pylearn's standard time module
import http # Pylearn's http client module

def my_task(task_id, duration):
    print(format_str("Task {task_id}: Starting at {time.time()}, will sleep for {duration}s"))
    time.sleep(duration) # Pylearn's blocking sleep
    result = format_str("Task {task_id}: Finished after {duration}s at {time.time()}")
    print(result)
    return result

def fetch_url_task(url):
    print(format_str("Fetching {url} in a thread..."))
    # http.get is synchronous here, but the whole task runs in a separate Go goroutine
    response = http.get(url)
    if response and response.status_code == 200:
        print(format_str("Fetched {url} successfully in thread."))
        obj = {"url": url, "title": response.json().get("title", "N/A (json error?)")}
        return obj
    else:
        status = "N/A"
        if response: 
            status = response.status_code
            print(format_str("Failed to fetch {url} in thread, status: {status}"))
        obj = {"url": url, "error": format_str("Failed with status {status}")}
        return obj



print("--- ThreadPoolExecutor Test ---")

# Create a thread pool
# 'with' statement for executors is common in Python for ensuring shutdown.
# Pylearn needs to support 'with' and the context manager protocol (__enter__, __exit__)
# on ThreadPoolExecutor for this exact syntax.
# For now, we'll manually call shutdown.
executor = ThreadPoolExecutor(5)
print(format_str("Executor created: {executor}"))

try:
    # Submit tasks
    future1 = executor.submit(my_task, "A", 2.0)
    future2 = executor.submit(my_task, "B", 1.0)
    future_http1 = executor.submit(fetch_url_task, "https://jsonplaceholder.typicode.com/todos/1")
    future_http2 = executor.submit(fetch_url_task, "https://jsonplaceholder.typicode.com/todos/2")


    print(format_str("Task A submitted, future: {future1}"))
    print(format_str("Task B submitted, future: {future2}"))
    print(format_str("HTTP Task 1 submitted, future: {future_http1}"))
    print(format_str("HTTP Task 2 submitted, future: {future_http2}"))

    # Get results (this will block until the task is done)
    print("\nWaiting for results...")
    result1 = future1.result() # Blocks
    print(format_str("Result from Task A: {result1}"))

    result2 = future2.result() # Blocks if not already done
    print(format_str("Result from Task B: {result2}"))
    
    http_result1 = future_http1.result()
    print(format_str("Result from HTTP Task 1: {http_result1}"))
    
    http_result2 = future_http2.result()
    print(format_str("Result from HTTP Task 2: {http_result2}"))

except Exception as e:
    print(format_str("An error occurred: {e}"))
# finally:
print("Shutting down executor...")
executor.shutdown(wait=True) # Wait for all tasks to complete before exiting
print("Executor shutdown complete.")