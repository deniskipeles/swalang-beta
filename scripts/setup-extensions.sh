#!/bin/bash
# ./scripts/setup-extensions.sh
# Create the .extensions directory if it doesn't exist
mkdir -p .extensions

# List of foundational C libraries (name repo_url)
declare -A LIBRARIES=(
  [zlib]="https://github.com/madler/zlib.git"
  [libcurl]="https://github.com/curl/curl.git"
  [libbz2]="https://sourceware.org/git/bzip2.git"
  [xz]="https://github.com/xz-mirror/xz.git"
  [libffi]="https://github.com/libffi/libffi.git"
  [libuv]="https://github.com/libuv/libuv.git"
  [mbedtls]="https://github.com/Mbed-TLS/mbedtls.git"
  [openssl]="https://github.com/openssl/openssl.git"
  [cjson]="https://github.com/DaveGamble/cJSON.git"
  [jansson]="https://github.com/akheron/jansson.git"
  [libyaml]="https://github.com/yaml/libyaml.git"
  [sqlite]="https://github.com/sqlite/sqlite.git"
  [libpng]="https://github.com/glennrp/libpng.git"
  [libjpeg-turbo]="https://github.com/libjpeg-turbo/libjpeg-turbo.git"
  [libgit2]="https://github.com/libgit2/libgit2.git"
  [tinyxml2]="https://github.com/leethomason/tinyxml2.git"
  [libarchive]="https://github.com/libarchive/libarchive.git"
  [pcre2]="https://github.com/PhilipHazel/pcre2.git"
  [oniguruma]="https://github.com/kkos/oniguruma.git"
  [libevent]="https://github.com/libevent/libevent.git"
  [mongoose]="https://github.com/cesanta/mongoose.git"
  [yyjson]="https://github.com/ibireme/yyjson.git"
)

echo "Cloning foundational C libraries as submodules..."

for lib in "${!LIBRARIES[@]}"; do
  target=".extensions/$lib"
  url="${LIBRARIES[$lib]}"
  if [ -d "$target" ]; then
    echo "✔ $lib already exists at $target"
  else
    echo "➕ Adding $lib..."
    git submodule add "$url" "$target"
  fi
done

echo "✅ Done! Now run:"
echo "   git submodule update --init --recursive"
