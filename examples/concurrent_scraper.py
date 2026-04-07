import time
import http
import re
import asyncio

# A list of websites to scrape.
# Note: For a real project, use sites that permit scraping.
# These examples are for demonstration.
URLS = [
    "https://example.com/",
    "https://www.google.com/nonexistent-page-to-show-error-handling", # This will fail
    "https://www.iana.org/domains/reserved",
    "https://deniskipeles.com/", # A personal site, usually simple
]

# A simple regex to find the <title> tag.
# Note: Using regex for HTML parsing is fragile. A real project
# should use a proper HTML parsing library. This is for demonstration.
TITLE_REGEX = re.compile("<title>(.*?)</title>")

async def fetch_and_parse(url):
    """
    An asynchronous function (a coroutine) to fetch a URL,
    extract its title, and print the result.
    """
    print(format_str("Starting fetch for {url}"))
    try:
        # 'await' pauses this function, allowing the event loop to run
        # other tasks while we wait for the network response.
        response = await http.get(url, {"timeout": 10}) # 10-second timeout

        # Once the response is ready, this function resumes.
        if response.status_code == 200:
            match = TITLE_REGEX.search(response.text)
            if match:
                # The .group(1) method gets the content of the first capturing group.
                title = match.group(1)
                print(format_str("SUCCESS: {url} -> Title: {title.strip()}"))
            else:
                print(format_str("NO TITLE: Could not find title for {url}"))
        else:
            print(format_str("HTTP ERROR: {url} -> Status: {response.status_code}"))
    
    except Exception as e:
        # Our native http client will raise errors for timeouts or connection issues.
        print(format_str("FETCH FAILED: {url} -> Error: {e}"))


async def main_program():
    """The main entry point for our asynchronous program."""
    
    # --- Part 1: Run tasks sequentially to see the 'slow' way ---
    print("--- Running tasks sequentially ---")
    start_time_seq = time.time()
    for url in URLS:
        await fetch_and_parse(url) # await each task one by one
    end_time_seq = time.time()
    print(format_str("Sequential execution took: {end_time_seq - start_time_seq:.2f} seconds\n"))

    # --- Part 2: Run tasks concurrently using asyncio ---
    print("--- Running tasks concurrently with asyncio.gather ---")
    start_time_concurrent = time.time()

    # Create a list of tasks. `create_task` schedules the coroutine
    # to run on the event loop as soon as possible. It does NOT block.
    tasks = []
    for url in URLS:
        task = asyncio.create_task(fetch_and_parse(url))
        tasks.append(task)

    # `asyncio.gather` is a convenient way to wait for all the
    # scheduled tasks to complete.
    await asyncio.gather(*tasks)

    end_time_concurrent = time.time()
    print(format_str("Concurrent execution took: {end_time_concurrent - start_time_concurrent:.2f} seconds"))

# asyncio.run() #is the entry point that starts the event loop,
# runs our main coroutine, and shuts the loop down.
# Our interpreter's main() function will call and await this.