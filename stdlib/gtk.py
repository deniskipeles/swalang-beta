# pylearn/stdlib/pygtk.py
"""
A Pylearn wrapper for the GTK 3 GUI toolkit, built using the Pylearn FFI.
This module provides a Pythonic, object-oriented interface to create
graphical user interfaces, with platform-aware library loading.
"""

import ffi
import sys # Import sys to check the platform

# --- Helper function for platform-aware library loading ---
def _load_library_with_fallbacks(base_name):
    """
    Tries to load a shared library using common platform-specific names.
    For example, for 'gtk-3', it will try 'libgtk-3-0.dll', 'libgtk-3.so.0', 'libgtk-3.dylib', etc.
    """
    platform = sys.platform # 'linux', 'darwin', 'win32'
    
    candidates = []
    if platform == 'win32':
        # Windows search order
        candidates = [
            format_str("lib{base_name}-0.dll"),
            format_str("{base_name}-0.dll"),
            format_str("{base_name}.dll"),
        ]
    elif platform == 'darwin':
        # macOS search order
        candidates = [
            format_str("lib{base_name}.0.dylib"),
            format_str("lib{base_name}.dylib"),
        ]
    else: # Assume Linux/other Unix-like
        # Linux search order
        candidates = [
            format_str("lib{base_name}.so.0"),
            format_str("lib{base_name}.so"),
        ]

    last_error = None
    for name in candidates:
        try:
            lib = ffi.CDLL(name)
            print(format_str("Successfully loaded {name}"))
            return lib
        except ffi.FFIError as e:
            last_error = e
    
    # If all candidates failed, raise the last error encountered
    raise last_error


# --- Load GTK and its dependency libraries ---
_glib = None
_gobject = None
_gdk = None
_gtk = None
GTK_AVAILABLE = False

try:
    # Order can be important. GLib is a core dependency.
    _glib = _load_library_with_fallbacks("glib-2.0")
    _gobject = _load_library_with_fallbacks("gobject-2.0")
    _gdk = _load_library_with_fallbacks("gdk-3")
    _gtk = _load_library_with_fallbacks("gtk-3")
    GTK_AVAILABLE = True
except ffi.FFIError as e:
    print(format_str("Warning: Failed to load one or more GTK libraries: {e}"))
    print("GUI functionality will not be available.")


# --- Global GTK Functions ---
_gtk_init = None
_gtk_main = None
_gtk_main_quit = None

if GTK_AVAILABLE:
    # void gtk_init(int *argc, char ***argv);
    # In Pylearn, we'll call it with NULL arguments.
    _gtk_init = _gtk.gtk_init([ffi.c_void_p, ffi.c_void_p], None)
    # void gtk_main(void);
    _gtk_main = _gtk.gtk_main([], None)
    # void gtk_main_quit(void);
    _gtk_main_quit = _gtk.gtk_main_quit([], None)


# --- Enums (mirrored as Pylearn classes with constants) ---
class WindowType:
    TOPLEVEL = 0

class Orientation:
    HORIZONTAL = 0
    VERTICAL = 1


# --- BaseWidget Class ---
class BaseWidget:
    """Base class for all GTK widgets in this wrapper."""
    def __init__(self, c_pointer):
        self._ptr = c_pointer 
        self._callbacks = {} 

    def show(self):
        _gtk.gtk_widget_show([ffi.c_void_p], None)(self._ptr)
    
    def show_all(self):
        _gtk.gtk_widget_show_all([ffi.c_void_p], None)(self._ptr)

    def connect(self, signal, pylearn_callback):
        # C Signature: gulong g_signal_connect(gpointer instance, const gchar *detailed_signal, GCallback c_handler, gpointer data);
        def c_callback_wrapper(widget_ptr, user_data_ptr):
            pylearn_callback(self)

        c_callback_ptr = ffi.callback(
            c_callback_wrapper,
            None,
            [ffi.c_void_p, ffi.c_void_p]
        )
        
        self._callbacks[signal] = c_callback_ptr

        _gobject.g_signal_connect_object(
            [ffi.c_void_p, ffi.c_char_p, ffi.c_void_p, ffi.c_void_p, ffi.c_int32],
            ffi.c_int64
        )(self._ptr, signal, c_callback_ptr, None, 0)
    
    def destroy(self):
        _gtk.gtk_widget_destroy([ffi.c_void_p], None)(self._ptr)


# --- Window Class ---
class Window(BaseWidget):
    def __init__(self, window_type=WindowType.TOPLEVEL):
        # GtkWidget* gtk_window_new(GtkWindowType type);
        c_ptr = _gtk.gtk_window_new([ffi.c_int32], ffi.c_void_p)(window_type)
        super().__init__(c_ptr)
        self.connect("destroy", lambda widget: main_quit())

    def set_title(self, title):
        _gtk.gtk_window_set_title([ffi.c_void_p, ffi.c_char_p], None)(self._ptr, title)

    def set_default_size(self, width, height):
        _gtk.gtk_window_set_default_size([ffi.c_void_p, ffi.c_int32, ffi.c_int32], None)(self._ptr, width, height)

    def add(self, widget):
        _gtk.gtk_container_add([ffi.c_void_p, ffi.c_void_p], None)(self._ptr, widget._ptr)


# --- Box Class ---
class Box(BaseWidget):
    def __init__(self, orientation=Orientation.VERTICAL, spacing=0):
        c_ptr = _gtk.gtk_box_new([ffi.c_int32, ffi.c_int32], ffi.c_void_p)(orientation, spacing)
        super().__init__(c_ptr)

    def pack_start(self, child_widget, expand, fill, padding):
        _gtk.gtk_box_pack_start(
            [ffi.c_void_p, ffi.c_void_p, ffi.c_int32, ffi.c_int32, ffi.c_int32], None
        )(self._ptr, child_widget._ptr, int(expand), int(fill), padding)

# --- Label Class ---
class Label(BaseWidget):
    def __init__(self, text):
        c_ptr = _gtk.gtk_label_new([ffi.c_char_p], ffi.c_void_p)(text)
        super().__init__(c_ptr)

# --- Button Class ---
class Button(BaseWidget):
    def __init__(self, label):
        c_ptr = _gtk.gtk_button_new_with_label([ffi.c_char_p], ffi.c_void_p)(label)
        super().__init__(c_ptr)


# --- Top-level Application Functions ---
def init():
    """Initializes the GTK library. Must be called before any other GTK function."""
    if GTK_AVAILABLE:
        _gtk_init(None, None)
    else:
        print("GTK not available, init() skipped.")

def main():
    """Starts the GTK main event loop. This function blocks until main_quit() is called."""
    if GTK_AVAILABLE:
        _gtk_main()
    else:
        print("GTK not available, main loop not started.")

def main_quit():
    """Stops the GTK main event loop."""
    if GTK_AVAILABLE:
        _gtk_main_quit()
    else:
        print("GTK not available, main_quit() skipped.")