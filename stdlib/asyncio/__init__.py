import uv

class Future:
    def __init__(self):
        self.done = False
        self.result = None
        self._callbacks = []

    def set_result(self, result):
        if self.done:
            return None
        self.result = result
        self.done = True
        for cb in self._callbacks:
            cb(self)

    def add_done_callback(self, cb):
        if self.done:
            cb(self)
        else:
            self._callbacks.append(cb)

    def __await__(self):
        # A future yields itself to the task manager, pausing the coroutine
        yield self
        return self.result

class Task(Future):
    def __init__(self, coro, loop):
        self.done = False
        self.result = None
        self._callbacks = []
        
        self.coro = coro
        self.loop = loop
        # Start executing the coroutine immediately
        self._step(None)

    def _step(self, value):
        try:
            result = self.coro.send(value)
            
            # If it yielded a Future, wait for it
            if hasattr(result, "add_done_callback"):
                result.add_done_callback(self._wakeup)
            # If it yielded another Coroutine, wrap it in a Task and wait
            elif hasattr(result, "send"):
                t = Task(result, self.loop)
                t.add_done_callback(self._wakeup)
            else:
                # If it yielded a plain value, just re-step immediately
                self._step(result)
                
        except StopIteration as e:
            # FIX: In Pylearn, the return value of a generator 
            # is stored in the 'value' attribute of the StopIteration error instance.
            val = None
            try:
                val = e.value
            except Exception:
                pass
            self.set_result(val)
        except Exception as e:
            print(format_str("Task Crash: {e}"))
            self.set_result(None)

    def _wakeup(self, future):
        # When the awaited future finishes, wake up the coroutine with the result
        self._step(future.result)


_loop = None

def get_event_loop():
    global _loop
    if _loop is None:
        _loop = uv.Loop()
    return _loop

def sleep(delay):
    fut = Future()
    loop = get_event_loop()
    timer = uv.Timer(loop)
    
    def on_timeout():
        timer.stop()
        timer.close()
        fut.set_result(None)
        
    # Start the native libuv timer
    timer.start(on_timeout, int(delay * 1000), 0)
    return fut

def create_task(coro):
    loop = get_event_loop()
    return Task(coro, loop)

def gather(*coros):
    fut = Future()
    loop = get_event_loop()
    
    results = []
    for i in range(len(coros)):
        results.append(None)
        
    state = [len(coros)] # Mutable list trick to keep state in closures

    if state[0] == 0:
        fut.set_result(results)
        return fut

    def make_done_callback(i):
        def cb(task):
            results[i] = task.result
            state[0] = state[0] - 1
            if state[0] == 0:
                fut.set_result(results)
        return cb

    for i in range(len(coros)):
        t = Task(coros[i], loop)
        t.add_done_callback(make_done_callback(i))

    return fut

def run(coro):
    loop = get_event_loop()
    task = Task(coro, loop)
    
    # Start the libuv event loop. It blocks until all tasks/timers finish!
    loop.run()
    
    global _loop
    _loop.close()
    _loop = None
    
    return task.result
