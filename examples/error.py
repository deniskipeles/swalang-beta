# examples/error.py 

def error(value):
    if value == 1:
        raise Exception('One')
    elif value == 2:
        raise Exception('Two')
    elif value == 3:
        raise Exception('Three')
    elif value == 4:
        raise Exception('Four')
    else:
        print('no error here')
try:
    e = error(1)
except Exception as e:
    print(e)

# --- Exception Classes ---
class CustomError(Exception):
    """Base exception for curl errors"""
    def __init__(self, code, message):
        self.code = code
        self.message = message
        super().__init__(format_str("Curl error {code}: {message}"))

def custom(val):
    if val == 1:
        raise CustomError(1,"Number One")
    else:
        print("No Number")

try:
    e = custom(1)
except Exception as e:
    print(e)

def issues_will_happen(val):
    raise Exception(val)

issues_will_happen("HEY!!!!")