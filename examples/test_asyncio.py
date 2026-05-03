import asyncio
import time

async def fetch_data(id, delay):
    print(format_str("Task {id}: Starting fetch..."))
    await asyncio.sleep(delay)
    print(format_str("Task {id}: Done!"))
    return format_str("Data from {id}")

async def main_program():
    print("🚀 Starting Asyncio Test!")
    start = time.time()
    
    # Run tasks concurrently!
    # Even though task 1 takes 2 seconds, the whole script finishes in 2 seconds, not 3!
    results = await asyncio.gather(
        fetch_data(1, 2.0),
        fetch_data(2, 1.0)
    )
    
    end = time.time()
    
    print(format_str("\n✅ All tasks finished in {end - start} seconds!"))
    print(format_str("Results: {results}"))

