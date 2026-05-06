# `lvgl` Module Reference

The `lvgl` module provides high-level bindings to the LVGL (Light and Versatile Graphics Library) version 9.

## Global Functions

- `init()`: Initializes the LVGL library.
- `tick_inc(ms)`: Informs LVGL that `ms` milliseconds have passed.
- `timer_handler()`: Processes LVGL timers, UI updates, and input reading. Returns the time until the next timer needs to run.

## Constants

### Alignments
- `LV_ALIGN_DEFAULT`
- `LV_ALIGN_TOP_LEFT`, `LV_ALIGN_TOP_MID`, `LV_ALIGN_TOP_RIGHT`
- `LV_ALIGN_BOTTOM_LEFT`, `LV_ALIGN_BOTTOM_MID`, `LV_ALIGN_BOTTOM_RIGHT`
- `LV_ALIGN_LEFT_MID`, `LV_ALIGN_RIGHT_MID`, `LV_ALIGN_CENTER`

### Input Device Types
- `LV_INDEV_TYPE_NONE`
- `LV_INDEV_TYPE_POINTER`
- `LV_INDEV_TYPE_KEYPAD`
- `LV_INDEV_TYPE_BUTTON`
- `LV_INDEV_TYPE_ENCODER`

### Input Device States
- `LV_INDEV_STATE_RELEASED`
- `LV_INDEV_STATE_PRESSED`

### Event Codes
- `LV_EVENT_PRESSED`
- `LV_EVENT_CLICKED`
- `LV_EVENT_VALUE_CHANGED`

## Classes

### `Display(w, h)`
Registers a display driver for LVGL.
- `ptr`: C pointer to the display object.
- `set_buffers(buf1_ptr, buf2_ptr, buf_size_bytes, render_mode=LV_DISPLAY_RENDER_MODE_PARTIAL)`: Configures draw buffers.
- `set_flush_cb(py_callback)`: Sets the callback function used to copy the rendered image to the display.
- `flush_ready()`: Informs LVGL that flushing is complete.

### `InputDevice(type_=LV_INDEV_TYPE_POINTER)`
Registers an input device.
- `ptr`: C pointer to the input device object.
- `set_read_cb(py_callback)`: Sets the callback function used to read the input device state.
- `write_pointer_data(data_ptr, x, y, state)`: (Static method) Helper to write pointer data into memory.

### `Screen` (Static Methods)
- `active()`: Returns the C pointer to the currently active screen.

### `Button(parent_ptr)`
Creates a button widget.
- `ptr`: C pointer to the button object.
- `align(alignment, x_ofs=0, y_ofs=0)`: Aligns the button.
- `set_size(w, h)`: Sets the dimensions.
- `add_event_cb(py_callback, event_filter)`: Adds an event callback.

### `Label(parent_ptr)`
Creates a label widget.
- `ptr`: C pointer to the label object.
- `set_text(text)`: Sets the label text.
- `align(alignment, x_ofs=0, y_ofs=0)`: Aligns the label.

## Structs

- `Area`: Fields `x1`, `y1`, `x2`, `y2` (all `c_int32`).
