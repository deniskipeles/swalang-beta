# pylearn/stdlib/asyncio.py

"""
A Pylearn implementation of the asyncio event loop API.

This module provides infrastructure for writing single-threaded concurrent
code using coroutines, multiplexing I/O access over sockets and other
resources, and running networked clients and servers.
"""

# Import the native Go implementation, which is registered as "aio".
# This provides the low-level bridge to the Go event loop.
# from pylearn_importlib import load_module_from_path
# _aio = load_module_from_path("aio")
import aio as _aio

# --- Public API ---

class Task:
    """A Future-like object that runs a Pylearn coroutine."""
    def __init__(self):
        # Users should not instantiate Task directly.
        raise TypeError("Tasks cannot be created directly. Use asyncio.create_task().")

def create_task(coro):
    """
    Schedule the execution of a coroutine object.
    Returns a Task object.
    """
    # The native _create_task function does all the work:
    # 1. Validates 'coro' is an async function.
    # 2. Creates a Go function to execute the Pylearn coroutine's body.
    # 3. Schedules the Go function on the event loop, getting an AsyncResult.
    # 4. Wraps everything in a Pylearn Task object and returns it.
    return _aio._create_task(coro)

def run(coro):
    """
    Run a coroutine and return its result.

    This function runs the given coroutine, taking care of managing the
    event loop and finalizing asynchronous generators.

    This function cannot be called when another event loop is running in
    the same thread.
    """
    task = create_task(coro)
    # The native _run_and_wait function blocks until the task is
    # complete and returns the result or raises the exception.
    return _aio._run_and_wait(task)

def sleep(delay):
    """
    Block for `delay` seconds.

    This is a coroutine.
    """
    # Simply pass through to the native sleep implementation.
    return _aio.sleep(delay)

# def gather(*aws):
#     """
#     Run awaitable objects in the `aws` sequence concurrently.

#     If any awaitable in `aws` is a coroutine, it is automatically
#     scheduled as a Task.

#     Returns a Task that waits on all provided awaitables.
#     """
#     # TODO: Automatically wrap coroutines in create_task if they aren't already tasks.
#     # For now, assumes all arguments are already awaitable (Tasks or native AsyncResults).
#     return _aio.gather(*aws)
def gather(*aws):
    print("Inside asyncio.gather, received args:", aws) # DEBUG
    result = _aio.gather(*aws)
    print("Native _aio.gather returned:", result) # DEBUG
    return result