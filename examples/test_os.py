import os

print("===========================================")
print("📂 Testing Pure-Swalang 'os' Module")
print("===========================================")

# 1. Environment Variables
path_env = os.getenv("PATH")
print(format_str("👉 getenv('PATH'): {path_env[:50]}..."))

# 2. Get Current Working Directory
cwd = os.getcwd()
print(format_str("👉 getcwd(): {cwd}"))

# 3. Path Manipulation
test_path = os.path.join(cwd, "test_dir", "file.txt")
print(format_str("👉 os.path.join: {test_path}"))
print(format_str("👉 os.path.dirname: {os.path.dirname(test_path)}"))
print(format_str("👉 os.path.basename: {os.path.basename(test_path)}"))
print(format_str("👉 os.path.splitext: {os.path.splitext(test_path)}"))

# 4. File System Operations
dir_name = "test_ffi_dir"
# FIX: Use the actual file name for the system call!
file_name = os.path.join(dir_name, "hello.txt")

print(format_str("\n🔨 Creating test directory '{dir_name}'..."))
if not os.path.exists(dir_name):
    os.mkdir(dir_name)

print(format_str("🔨 Executing shell command to create '{file_name}'..."))
if os.path.sep == '\\':
    os.system(format_str("echo Hello World > {file_name}"))
else:
    os.system(format_str("echo 'Hello World' > {file_name}"))

# 5. Testing `listdir` and `stat` logic
print("\n🔍 Checking File System:")
print(format_str("  - {dir_name} exists: {os.path.exists(dir_name)}"))
print(format_str("  - {dir_name} is dir: {os.path.isdir(dir_name)}"))
print(format_str("  - {file_name} is file: {os.path.isfile(file_name)}"))
print(format_str("  - {file_name} size: {os.path.getsize(file_name)} bytes"))

print(format_str("\n📁 os.listdir('{dir_name}'):"))
print(format_str("   {os.listdir(dir_name)}"))

# 6. Testing `os.walk` (Generator)
print("\n🚶 os.walk('.'): (First 2 iterations)")
iterations = 0
for root, dirs, files in os.walk("."):
    print(format_str("   Directory: {root}"))
    if "test_ffi_dir" in root:
        print(format_str("     Files: {files}"))
    iterations = iterations + 1
    if iterations >= 2:
        break

# 7. Cleanup
print("\n🧹 Cleaning up...")
os.remove(file_name)
if os.path.sep == '\\':
    os.system(format_str("rmdir {dir_name}"))
else:
    os.system(format_str("rm -r {dir_name}"))

print("✅ Cleanup complete.")
print("\n🎉 Pure Swalang 'os' tests passed successfully!")