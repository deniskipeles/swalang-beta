# examples/hv_test.py
import hv

@hv.server.route("/hello", "GET")
def hello(req, resp):
    return resp.status(200).body("Hello, HV from Pylearn!")

if __name__ == "__main__":
    hv.server.on("request").start()