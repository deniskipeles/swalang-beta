# pylearn/stdlib/tinyxml2.py

import ffi
import sys

# ==============================================================================
#  Load the C++ -> C Bridge Library
# ==============================================================================

def _load_library_with_fallbacks(base_name):
    platform = sys.platform
    candidates = []
    if platform == 'windows':
        candidates = [base_name + '.dll']
    elif platform == 'darwin':
        candidates = ['lib' + base_name + '.dylib']
    else: # Linux
        candidates = ['lib' + base_name + '.so']

    last_error = None
    for name in candidates:
        try:
            # Note: You may need to adjust the path if the .so/.dll is not
            # in the current directory or a standard system path.
            return ffi.CDLL(name)
        except ffi.FFIError as e:
            last_error = e
    raise last_error

_lib = None
TINYXML2_AVAILABLE = False
try:
    # <<< THIS IS THE KEY CHANGE >>>
    # We are now loading our compiled C bridge, not the raw C++ library.
    _lib = _load_library_with_fallbacks("tinyxml2_bridge")
    TINYXML2_AVAILABLE = True
except ffi.FFIError as e:
    print(format_str("Warning: Failed to load tinyxml2_bridge library: {e}"))
    print("TinyXML2 functionality will not be available.")

# ==============================================================================
#  Define C Function Signatures from the Shim
# ==============================================================================

if TINYXML2_AVAILABLE:
    # Document functions
    _XMLDocument_New = _lib.XMLDocument_New([], ffi.c_void_p)
    _XMLDocument_Delete = _lib.XMLDocument_Delete([ffi.c_void_p], None)
    _XMLDocument_Parse = _lib.XMLDocument_Parse([ffi.c_void_p, ffi.c_char_p], ffi.c_int32)
    _XMLDocument_FirstChildElement = _lib.XMLDocument_FirstChildElement([ffi.c_void_p, ffi.c_char_p], ffi.c_void_p)

    # Element functions
    _XMLElement_Name = _lib.XMLElement_Name([ffi.c_void_p], ffi.c_char_p)
    _XMLElement_GetText = _lib.XMLElement_GetText([ffi.c_void_p], ffi.c_char_p)
    _XMLElement_Attribute = _lib.XMLElement_Attribute([ffi.c_void_p, ffi.c_char_p], ffi.c_char_p)
    _XMLElement_FirstChildElement = _lib.XMLElement_FirstChildElement([ffi.c_void_p, ffi.c_char_p], ffi.c_void_p)
    _XMLElement_NextSiblingElement = _lib.XMLElement_NextSiblingElement([ffi.c_void_p, ffi.c_char_p], ffi.c_void_p)

# ==============================================================================
#  Pythonic Wrapper Classes (No changes from here on)
# ==============================================================================

XML_SUCCESS = 0

class XMLError(Exception):
    pass

class XMLElement:
    def __init__(self, ptr):
        self._ptr = ptr

    @property
    def name(self):
        """The name of the XML element tag."""
        # No change needed here as _XMLElement_Name returns a C string
        return ffi.string_at(_XMLElement_Name(self._ptr))

    @property
    def text(self):
        """The inner text of the element, or None if it has no text."""
        text_ptr = _XMLElement_GetText(self._ptr)
        return ffi.string_at(text_ptr) if text_ptr else None

    def first_child_element(self, name=""):
        """
        Returns the first child element with the given name.
        If name is empty, returns the first child element.
        """
        # <<< FIX: Remove .encode() >>>
        child_ptr = _XMLElement_FirstChildElement(self._ptr, name)
        return XMLElement(child_ptr) if child_ptr.Address != 0 else None

    def next_sibling_element(self, name=""):
        """
        Returns the next sibling element with the given name.
        If name is empty, returns the next sibling element.
        """
        # <<< FIX: Remove .encode() >>>
        sibling_ptr = _XMLElement_NextSiblingElement(self._ptr, name)
        return XMLElement(sibling_ptr) if sibling_ptr.Address != 0 else None

    def attribute(self, name):
        """Returns the value of an attribute, or None if it doesn't exist."""
        # <<< FIX: Remove .encode() >>>
        attr_ptr = _XMLElement_Attribute(self._ptr, name)
        return ffi.string_at(attr_ptr) if attr_ptr else None

    def __repr__(self):
        return format_str("<XMLElement {self.name}>")

class XMLDocument:
    # ... (init and free are unchanged) ...
    def __init__(self):
        self._ptr = _XMLDocument_New()
        if self._ptr.Address == 0:
            raise MemoryError("Failed to create XMLDocument")

    def free(self):
        if self._ptr:
            _XMLDocument_Delete(self._ptr)
            self._ptr = None

    def parse(self, xml_text):
        """
        Parses an XML string.
        Raises XMLError on failure.
        """
        if not isinstance(xml_text, str):
            raise TypeError("parse() argument must be a string")
        
        # <<< FIX: Remove .encode() >>>
        res = _XMLDocument_Parse(self._ptr, xml_text)
        
        if res != XML_SUCCESS:
            raise XMLError(format_str("XML parsing failed with error code: {res}"))

    def first_child_element(self, name=""):
        """Returns the root element of the document."""
        # <<< FIX: Remove .encode() >>>
        root_ptr = _XMLDocument_FirstChildElement(self._ptr, name)
        return XMLElement(root_ptr) if root_ptr.Address != 0 else None

    def __repr__(self):
        return "<XMLDocument>"