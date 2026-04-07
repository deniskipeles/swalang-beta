import gtk as pygtk

# This will be our callback function when the button is clicked.
def on_button_clicked(widget):
    print("Hello, GTK World!")
    # We can call methods on the widget passed to the callback
    widget.destroy() # This will close the window due to the 'destroy' signal

def on_window_destroy(widget):
    print("Window is closing. Quitting main loop.")
    pygtk.main_quit()

# --- Main application logic ---

# 1. Initialize GTK
# This must be the very first GTK call.
pygtk.init()

# 2. Create the main window
window = pygtk.Window(pygtk.WindowType.TOPLEVEL)
window.set_title("Pylearn GTK Test")
window.set_default_size(300, 200)

# 3. Connect the "destroy" signal of the window to our Pylearn function
window.connect("destroy", on_window_destroy)

# 4. Create a button
button = pygtk.Button("Click Me!")

# 5. Connect the "clicked" signal of the button to our Pylearn function
button.connect("clicked", on_button_clicked)

# 6. Add the button to the window
window.add(button)

# 7. Show all widgets
window.show_all()

# 8. Start the GTK main event loop.
# This is a blocking call and will run until pygtk.main_quit() is called.
print("Starting GTK main loop...")
pygtk.main()
print("GTK main loop has finished.")