# 1. fetch sources
cd /tmp
wget https://github.com/libffi/libffi/releases/download/v3.4.6/libffi-3.4.6.tar.gz
tar -xf libffi-3.4.6.tar.gz
cd libffi-3.4.6

# 2. configure for Windows 64-bit
./configure --host=x86_64-w64-mingw32 \
            --prefix=/usr/x86_64-w64-mingw32 \
            --disable-shared --enable-static

# 3. build & install
make -j$(nproc)
sudo make install