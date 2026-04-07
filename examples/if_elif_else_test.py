# examples/if_elif_else.py

# --- IF/ELIF/ELSE TEST ---
print("--- If/Elif/Else Test ---")
val = 15

if val < 10:
    print("val is less than 10")
elif val < 20:
    print("val is less than 20 but not less than 10") # Should print this
else:
    print("val is 20 or greater")

val = 5
if val < 10:
    print("val is less than 10") # Should print this
elif val < 20:
    print("val is less than 20 but not less than 10")
else:
    print("val is 20 or greater")

val = 25
if val < 10:
    print("val is less than 10")
elif val < 20:
    print("val is less than 20 but not less than 10")
else:
    print("val is 20 or greater") # Should print this

# Test without else
val = 100
if val == 50:
    print("val is 50")
elif val == 75:
    print("val is 75")
# Should print nothing here

print("--- End If/Elif/Else Test ---")
# --- END IF/ELIF/ELSE TEST ---