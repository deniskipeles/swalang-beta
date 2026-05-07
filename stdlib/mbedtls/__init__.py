"""
MbedTLS wrapper for Pylearn/Swalang.
Covers: Hashing, HMAC, AES-CBC, AES-GCM, Base64, CSPRNG, TLS, X.509
"""

import ffi
import sys

# ==============================================================================
#  Constants
# ==============================================================================

MBEDTLS_DECRYPT         = 0
MBEDTLS_ENCRYPT         = 1
MBEDTLS_CIPHER_ID_AES   = 2
MBEDTLS_SSL_IS_CLIENT   = 0
MBEDTLS_SSL_IS_SERVER   = 1
MBEDTLS_SSL_TRANSPORT_STREAM   = 0
MBEDTLS_SSL_TRANSPORT_DATAGRAM = 1
MBEDTLS_SSL_PRESET_DEFAULT     = 0

# Opaque struct sizes — padded generously so we never underestimate
_CTX_MD         = 256
_CTX_AES        = 512
_CTX_GCM        = 512
_CTX_ENTROPY    = 8192
_CTX_CTR_DRBG   = 2048
_CTX_SSL        = 8192
_CTX_SSL_CONFIG = 8192
_CTX_X509_CRT   = 8192

# ==============================================================================
#  Library Loading
# ==============================================================================

class MbedError(Exception):
    pass

def _try_load(names):
    for name in names:
        try:
            return ffi.CDLL(name)
        except ffi.FFIError:
            pass
    return None

# On Linux, crypto symbols live in libtfpsacrypto (= libmbedcrypto in this
# mbedtls release). TLS symbols are in libmbedtls, X.509 in libmbedx509.
# On Windows, everything is merged into a single libmbedtls.dll.

_base = "bin"

if sys.platform == "linux":
    _lib_crypto = _try_load([
        _base + "/x86_64-linux/mbedtls/libtfpsacrypto.so",
        "libtfpsacrypto.so",
        _base + "/x86_64-linux/mbedtls/libmbedcrypto.so",
        "libmbedcrypto.so",
    ])
    _lib_x509 = _try_load([
        _base + "/x86_64-linux/mbedtls/libmbedx509.so",
        "libmbedx509.so",
    ])
    _lib_tls = _try_load([
        _base + "/x86_64-linux/mbedtls/libmbedtls.so",
        "libmbedtls.so",
    ])
elif sys.platform == "windows":
    _lib_tls = _try_load([
        _base + "/x86_64-windows-gnu/mbedtls/libmbedtls.dll",
        "libmbedtls.dll",
    ])
    _lib_crypto = _lib_tls
    _lib_x509   = _lib_tls
else:
    _lib_crypto = _try_load(["libmbedcrypto.dylib"])
    _lib_x509   = _try_load(["libmbedx509.dylib"])
    _lib_tls    = _try_load(["libmbedtls.dylib"])

CRYPTO_AVAILABLE = _lib_crypto is not None
TLS_AVAILABLE    = _lib_tls    is not None
X509_AVAILABLE   = _lib_x509   is not None

# ==============================================================================
#  Function Bindings — Crypto / Hash / HMAC
# ==============================================================================

if CRYPTO_AVAILABLE:
    _md_info_from_string = _lib_crypto.mbedtls_md_info_from_string(
        [ffi.c_char_p], ffi.c_void_p)
    _md_get_size = _lib_crypto.mbedtls_md_get_size(
        [ffi.c_void_p], ffi.c_uchar)
    _md_init   = _lib_crypto.mbedtls_md_init(  [ffi.c_void_p], None)
    _md_free   = _lib_crypto.mbedtls_md_free(  [ffi.c_void_p], None)
    _md_setup  = _lib_crypto.mbedtls_md_setup( [ffi.c_void_p, ffi.c_void_p, ffi.c_int32], ffi.c_int32)
    _md_starts = _lib_crypto.mbedtls_md_starts([ffi.c_void_p], ffi.c_int32)
    _md_update = _lib_crypto.mbedtls_md_update([ffi.c_void_p, ffi.c_char_p, ffi.c_uint64], ffi.c_int32)
    _md_finish = _lib_crypto.mbedtls_md_finish([ffi.c_void_p, ffi.c_char_p], ffi.c_int32)
    _md_oneshot = _lib_crypto.mbedtls_md(
        [ffi.c_void_p, ffi.c_char_p, ffi.c_uint64, ffi.c_char_p], ffi.c_int32)

    _aes_init       = _lib_crypto.mbedtls_aes_init(   [ffi.c_void_p], None)
    _aes_free       = _lib_crypto.mbedtls_aes_free(   [ffi.c_void_p], None)
    _aes_setkey_enc = _lib_crypto.mbedtls_aes_setkey_enc(
        [ffi.c_void_p, ffi.c_char_p, ffi.c_uint32], ffi.c_int32)
    _aes_setkey_dec = _lib_crypto.mbedtls_aes_setkey_dec(
        [ffi.c_void_p, ffi.c_char_p, ffi.c_uint32], ffi.c_int32)
    _aes_crypt_cbc  = _lib_crypto.mbedtls_aes_crypt_cbc(
        [ffi.c_void_p, ffi.c_int32, ffi.c_uint64,
         ffi.c_char_p, ffi.c_char_p, ffi.c_char_p], ffi.c_int32)

    _gcm_init    = _lib_crypto.mbedtls_gcm_init(   [ffi.c_void_p], None)
    _gcm_free    = _lib_crypto.mbedtls_gcm_free(   [ffi.c_void_p], None)
    _gcm_setkey  = _lib_crypto.mbedtls_gcm_setkey(
        [ffi.c_void_p, ffi.c_int32, ffi.c_char_p, ffi.c_uint32], ffi.c_int32)
    _gcm_crypt_and_tag = _lib_crypto.mbedtls_gcm_crypt_and_tag(
        [ffi.c_void_p, ffi.c_int32, ffi.c_uint64,
         ffi.c_char_p, ffi.c_uint64,
         ffi.c_char_p, ffi.c_uint64,
         ffi.c_char_p, ffi.c_char_p,
         ffi.c_uint64, ffi.c_char_p], ffi.c_int32)
    _gcm_auth_decrypt = _lib_crypto.mbedtls_gcm_auth_decrypt(
        [ffi.c_void_p, ffi.c_uint64,
         ffi.c_char_p, ffi.c_uint64,
         ffi.c_char_p, ffi.c_uint64,
         ffi.c_char_p, ffi.c_uint64,
         ffi.c_char_p, ffi.c_char_p], ffi.c_int32)

    _b64_encode = _lib_crypto.mbedtls_base64_encode(
        [ffi.c_char_p, ffi.c_uint64, ffi.c_void_p,
         ffi.c_char_p, ffi.c_uint64], ffi.c_int32)
    _b64_decode = _lib_crypto.mbedtls_base64_decode(
        [ffi.c_char_p, ffi.c_uint64, ffi.c_void_p,
         ffi.c_char_p, ffi.c_uint64], ffi.c_int32)

    _entropy_init = _lib_crypto.mbedtls_entropy_init([ffi.c_void_p], None)
    _entropy_free = _lib_crypto.mbedtls_entropy_free([ffi.c_void_p], None)
    _drbg_init    = _lib_crypto.mbedtls_ctr_drbg_init([ffi.c_void_p], None)
    _drbg_free    = _lib_crypto.mbedtls_ctr_drbg_free([ffi.c_void_p], None)
    _drbg_seed    = _lib_crypto.mbedtls_ctr_drbg_seed(
        [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p,
         ffi.c_char_p, ffi.c_uint64], ffi.c_int32)
    _drbg_random  = _lib_crypto.mbedtls_ctr_drbg_random(
        [ffi.c_void_p, ffi.c_char_p, ffi.c_uint64], ffi.c_int32)
    _entropy_func = _lib_crypto.mbedtls_entropy_func(
        [ffi.c_void_p, ffi.c_char_p, ffi.c_uint64], ffi.c_int32)

if TLS_AVAILABLE:
    _ssl_init        = _lib_tls.mbedtls_ssl_init(        [ffi.c_void_p], None)
    _ssl_free        = _lib_tls.mbedtls_ssl_free(        [ffi.c_void_p], None)
    _ssl_cfg_init    = _lib_tls.mbedtls_ssl_config_init( [ffi.c_void_p], None)
    _ssl_cfg_free    = _lib_tls.mbedtls_ssl_config_free( [ffi.c_void_p], None)
    _ssl_cfg_defaults = _lib_tls.mbedtls_ssl_config_defaults(
        [ffi.c_void_p, ffi.c_int32, ffi.c_int32, ffi.c_int32], ffi.c_int32)
    _ssl_conf_authmode = _lib_tls.mbedtls_ssl_conf_authmode(
        [ffi.c_void_p, ffi.c_int32], None)
    _ssl_conf_ca_chain = _lib_tls.mbedtls_ssl_conf_ca_chain(
        [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], None)
    try:
        _ssl_conf_rng  = _lib_tls.mbedtls_ssl_conf_rng([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], None)
    except ffi.FFIError:
        _ssl_conf_rng = None
    _ssl_setup     = _lib_tls.mbedtls_ssl_setup(
        [ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
    _ssl_set_hostname = _lib_tls.mbedtls_ssl_set_hostname(
        [ffi.c_void_p, ffi.c_char_p], ffi.c_int32)
    _ssl_set_bio   = _lib_tls.mbedtls_ssl_set_bio(
        [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p,
         ffi.c_void_p, ffi.c_void_p], None)
    _ssl_handshake = _lib_tls.mbedtls_ssl_handshake(
        [ffi.c_void_p], ffi.c_int32)
    _ssl_read      = _lib_tls.mbedtls_ssl_read(
        [ffi.c_void_p, ffi.c_char_p, ffi.c_uint64], ffi.c_int32)
    _ssl_write     = _lib_tls.mbedtls_ssl_write(
        [ffi.c_void_p, ffi.c_char_p, ffi.c_uint64], ffi.c_int32)
    _ssl_close     = _lib_tls.mbedtls_ssl_close_notify(
        [ffi.c_void_p], ffi.c_int32)
    _ssl_get_verify_result = _lib_tls.mbedtls_ssl_get_verify_result(
        [ffi.c_void_p], ffi.c_uint32)

if X509_AVAILABLE:
    _x509_init      = _lib_x509.mbedtls_x509_crt_init(  [ffi.c_void_p], None)
    _x509_free      = _lib_x509.mbedtls_x509_crt_free(  [ffi.c_void_p], None)
    _x509_parse     = _lib_x509.mbedtls_x509_crt_parse(
        [ffi.c_void_p, ffi.c_char_p, ffi.c_uint64], ffi.c_int32)
    _x509_parse_file = _lib_x509.mbedtls_x509_crt_parse_file(
        [ffi.c_void_p, ffi.c_char_p], ffi.c_int32)
    _x509_verify    = _lib_x509.mbedtls_x509_crt_verify(
        [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p,
         ffi.c_char_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
    _x509_info      = _lib_x509.mbedtls_x509_crt_info(
        [ffi.c_char_p, ffi.c_uint64, ffi.c_char_p, ffi.c_void_p], ffi.c_int32)

# ==============================================================================
#  Helpers
# ==============================================================================

def _check(ret, label):
    if ret != 0:
        raise MbedError(format_str("{label} failed with code {ret:#010x}"))

def _require_crypto():
    if not CRYPTO_AVAILABLE:
        raise MbedError("MbedTLS crypto library not available")

def _require_tls():
    if not TLS_AVAILABLE:
        raise MbedError("MbedTLS TLS library not available")

def _require_x509():
    if not X509_AVAILABLE:
        raise MbedError("MbedTLS X.509 library not available")

def _to_bytes(data):
    if isinstance(data, str):
        return data.encode("utf-8")
    return data

# ==============================================================================
#  Hash
# ==============================================================================

class Hash:
    """
    Streaming hash context. Mirrors hashlib's interface.

    Usage:
        h = Hash("SHA256")
        h.update(b"hello")
        print(h.hexdigest())

        # one-shot
        h2 = Hash("SHA256", b"hello world")
        digest = h2.digest()
        h2.free()
    """
    def __init__(self, algorithm, data=b''):
        _require_crypto()
        self.name = algorithm.upper()
        self._freed = False

        self._info = _md_info_from_string(self.name.encode("utf-8"))
        if self._info.Address == 0:
            raise MbedError(format_str("Unknown hash algorithm: '{self.name}'"))

        size_val = _md_get_size(self._info)
        self.digest_size = int(str(size_val))
        self._ctx = ffi.malloc(_CTX_MD)

        try:
            _md_init(self._ctx)
            _check(_md_setup(self._ctx, self._info, 0), "md_setup")
            _check(_md_starts(self._ctx), "md_starts")
            if data:
                self.update(data)
        except MbedError:
            self.free()
            raise

    def update(self, data):
        if self._freed:
            raise MbedError("Hash context already freed")
        b = _to_bytes(data)
        _check(_md_update(self._ctx, b, len(b)), "md_update")

    def digest(self):
        if self._freed:
            raise MbedError("Hash context already freed")
        out = ffi.malloc(self.digest_size)
        try:
            _check(_md_finish(self._ctx, out), "md_finish")
            result = ffi.buffer_to_bytes(out, self.digest_size)
            # Restart so the object is reusable
            _check(_md_starts(self._ctx), "md_starts (restart)")
            return result
        finally:
            ffi.free(out)

    def hexdigest(self):
        raw = self.digest()
        hex_chars = "0123456789abcdef"
        result = ""
        for b in raw:
            result = result + hex_chars[b // 16] + hex_chars[b % 16]
        return result

    def copy(self):
        """Return a new Hash with identical state by re-hashing is not
        supported; use the one-shot helpers for cloning patterns."""
        raise MbedError("copy() not supported — use a new Hash instance")

    def free(self):
        if not self._freed and self._ctx:
            _md_free(self._ctx)
            ffi.free(self._ctx)
            self._ctx = None
            self._freed = True

# ==============================================================================
#  HMAC (Pure Swalang Implementation using Hash)
# ==============================================================================

class HMAC:
    """
    HMAC context implemented in pure Swalang using the MbedTLS Hash primitive.
    Bypasses MbedTLS 4.0/PSA MAC API instability.

    Usage:
        h = HMAC("SHA256", b"secret-key")
        h.update(b"message")
        print(h.hexdigest())
        h.free()
    """
    def __init__(self, algorithm, key, data=b''):
        _require_crypto()
        self.algorithm = algorithm.upper()
        self._freed = False
        key = _to_bytes(key)

        if self.algorithm in ["SHA512", "SHA384"]:
            self.block_size = 128
        else:
            self.block_size = 64

        # Initialize hash context to get digest size and validate algorithm
        temp_hash = Hash(self.algorithm)
        self.digest_size = temp_hash.digest_size
        temp_hash.free()

        # Format the key
        if len(key) > self.block_size:
            key_hash = Hash(self.algorithm, key)
            key = key_hash.digest()
            key_hash.free()
            
        if len(key) < self.block_size:
            pad_len = self.block_size - len(key)
            zero_pad = b'\x00' * pad_len
            key = key + zero_pad

        ipad_list = []
        opad_list = []
        for i in range(self.block_size):
            ipad_list.append(key[i] ^ 0x36)
            opad_list.append(key[i] ^ 0x5c)
            
        self._ipad = bytes(ipad_list)
        self._opad = bytes(opad_list)
        
        # Start the inner hash
        self._inner_hash = Hash(self.algorithm)
        self._inner_hash.update(self._ipad)
        
        if data:
            self.update(data)

    def update(self, data):
        if self._freed:
            raise MbedError("HMAC context already freed")
        self._inner_hash.update(_to_bytes(data))

    def digest(self):
        if self._freed:
            raise MbedError("HMAC context already freed")
            
        inner_digest = self._inner_hash.digest()
        
        outer_hash = Hash(self.algorithm)
        outer_hash.update(self._opad)
        outer_hash.update(inner_digest)
        result = outer_hash.digest()
        outer_hash.free()
        
        # Restart the inner hash for subsequent uses (matching MbedTLS behavior)
        self._inner_hash.update(self._ipad)
        
        return result

    def hexdigest(self):
        raw = self.digest()
        hex_chars = "0123456789abcdef"
        result = ""
        for b in raw:
            result = result + hex_chars[b // 16] + hex_chars[b % 16]
        return result

    def free(self):
        if not self._freed:
            self._inner_hash.free()
            self._freed = True

# ==============================================================================
#  AES-CBC Cipher
# ==============================================================================

class AES_CBC:
    """
    AES in CBC mode.  Key must be 16, 24, or 32 bytes (128/192/256-bit).
    IV must be exactly 16 bytes.  Input length must be a multiple of 16.

    Usage:
        cipher = AES_CBC(key)
        ct = cipher.encrypt(iv, plaintext)
        pt = cipher.decrypt(iv, ct)
        cipher.free()
    """
    def __init__(self, key):
        _require_crypto()
        self._key  = _to_bytes(key)
        self._bits = len(self._key) * 8
        if self._bits not in [128, 192, 256]:
            raise MbedError("AES key must be 16, 24 or 32 bytes")
        self._ctx  = ffi.malloc(_CTX_AES)
        self._freed = False
        _aes_init(self._ctx)

    def encrypt(self, iv, plaintext):
        if self._freed:
            raise MbedError("AES context already freed")
        iv = _to_bytes(iv)
        pt = _to_bytes(plaintext)
        if len(iv) != 16:
            raise MbedError("IV must be 16 bytes")
        if len(pt) % 16 != 0:
            raise MbedError("Plaintext length must be a multiple of 16 bytes")

        iv_buf = ffi.malloc(16)
        out    = ffi.malloc(len(pt))
        try:
            ffi.memcpy(iv_buf, iv, 16)
            _check(_aes_setkey_enc(self._ctx, self._key, self._bits), "aes_setkey_enc")
            _check(_aes_crypt_cbc(self._ctx, MBEDTLS_ENCRYPT, len(pt), iv_buf, pt, out), "aes_crypt_cbc(enc)")
            return ffi.buffer_to_bytes(out, len(pt))
        finally:
            ffi.free(iv_buf)
            ffi.free(out)

    def decrypt(self, iv, ciphertext):
        if self._freed:
            raise MbedError("AES context already freed")
        iv = _to_bytes(iv)
        ct = _to_bytes(ciphertext)
        if len(iv) != 16:
            raise MbedError("IV must be 16 bytes")
        if len(ct) % 16 != 0:
            raise MbedError("Ciphertext length must be a multiple of 16 bytes")

        iv_buf = ffi.malloc(16)
        out    = ffi.malloc(len(ct))
        try:
            ffi.memcpy(iv_buf, iv, 16)
            _check(_aes_setkey_dec(self._ctx, self._key, self._bits), "aes_setkey_dec")
            _check(_aes_crypt_cbc(self._ctx, MBEDTLS_DECRYPT, len(ct), iv_buf, ct, out), "aes_crypt_cbc(dec)")
            return ffi.buffer_to_bytes(out, len(ct))
        finally:
            ffi.free(iv_buf)
            ffi.free(out)

    def free(self):
        if not self._freed and self._ctx:
            _aes_free(self._ctx)
            ffi.free(self._ctx)
            self._ctx = None
            self._freed = True

# ==============================================================================
#  AES-GCM Cipher
# ==============================================================================

class AES_GCM:
    """
    AES in GCM mode (authenticated encryption).
    Key must be 16, 24, or 32 bytes.
    Returns (ciphertext, tag) from encrypt() and plaintext from decrypt().

    Usage:
        g = AES_GCM(key)
        ct, tag = g.encrypt(iv, plaintext, aad=b'')
        pt = g.decrypt(iv, ct, tag, aad=b'')
        g.free()
    """
    def __init__(self, key):
        _require_crypto()
        self._key  = _to_bytes(key)
        self._bits = len(self._key) * 8
        if self._bits not in [128, 192, 256]:
            raise MbedError("AES-GCM key must be 16, 24 or 32 bytes")
        self._ctx  = ffi.malloc(_CTX_GCM)
        self._freed = False
        _gcm_init(self._ctx)
        _check(_gcm_setkey(self._ctx, MBEDTLS_CIPHER_ID_AES, self._key, self._bits), "gcm_setkey")

    def encrypt(self, iv, plaintext, aad=b'', tag_len=16):
        if self._freed:
            raise MbedError("GCM context already freed")
        iv  = _to_bytes(iv)
        pt  = _to_bytes(plaintext)
        aad = _to_bytes(aad)
        out = ffi.malloc(len(pt))
        tag = ffi.malloc(tag_len)
        try:
            _check(_gcm_crypt_and_tag(self._ctx, MBEDTLS_ENCRYPT, len(pt), iv, len(iv), aad, len(aad), pt, out, tag_len, tag), "gcm_crypt_and_tag")
            return (ffi.buffer_to_bytes(out, len(pt)), ffi.buffer_to_bytes(tag, tag_len))
        except Exception as e:
            pass
        finally:
            ffi.free(out)
            ffi.free(tag)

    def decrypt(self, iv, ciphertext, tag, aad=b''):
        if self._freed:
            raise MbedError("GCM context already freed")
        iv  = _to_bytes(iv)
        ct  = _to_bytes(ciphertext)
        tag = _to_bytes(tag)
        aad = _to_bytes(aad)
        out = ffi.malloc(len(ct))
        try:
            _check(_gcm_auth_decrypt(self._ctx, len(ct), iv, len(iv), aad, len(aad), tag, len(tag), ct, out), "gcm_auth_decrypt")
            return ffi.buffer_to_bytes(out, len(ct))
        except Exception as e:
            pass
        finally:
            ffi.free(out)

    def free(self):
        if not self._freed and self._ctx:
            _gcm_free(self._ctx)
            ffi.free(self._ctx)
            self._ctx = None
            self._freed = True

# ==============================================================================
#  Base64
# ==============================================================================

def b64encode(data):
    """Encode bytes to Base64 bytes."""
    _require_crypto()
    data = _to_bytes(data)
    # Worst-case output: ceil(len/3)*4 + 1
    out_size = ((len(data) + 2) // 3) * 4 + 1
    out = ffi.malloc(out_size)
    olen_buf = ffi.malloc(8)  # size_t
    try:
        _check(_b64_encode(out, out_size, olen_buf, data, len(data)), "base64_encode")
        # Read olen from buffer
        olen_bytes = ffi.buffer_to_bytes(olen_buf, 8)
        olen = 0
        for i in range(8):
            olen = olen + olen_bytes[i] * (256 ** i)
        return ffi.buffer_to_bytes(out, olen)
    finally:
        ffi.free(out)
        ffi.free(olen_buf)

def b64decode(data):
    """Decode Base64 bytes to raw bytes."""
    _require_crypto()
    data = _to_bytes(data)
    out_size = len(data)  # decoded is always <= encoded
    out = ffi.malloc(out_size)
    olen_buf = ffi.malloc(8)
    try:
        _check(_b64_decode(out, out_size, olen_buf, data, len(data)), "base64_decode")
        olen_bytes = ffi.buffer_to_bytes(olen_buf, 8)
        olen = 0
        for i in range(8):
            olen = olen + olen_bytes[i] * (256 ** i)
        return ffi.buffer_to_bytes(out, olen)
    finally:
        ffi.free(out)
        ffi.free(olen_buf)

def b64encode_str(data):
    """Encode bytes to a Base64 string."""
    return b64encode(data).decode("ascii")

def b64decode_str(s):
    """Decode a Base64 string to bytes."""
    return b64decode(s.encode("ascii"))

# ==============================================================================
#  CSPRNG
# ==============================================================================

class Random:
    """
    Cryptographically secure random number generator (CTR-DRBG + entropy).

    Usage:
        rng = Random()
        key = rng.read(32)   # 32 random bytes
        rng.free()
    """
    def __init__(self, personalization=b''):
        _require_crypto()
        self._freed = False
        self._entropy = ffi.malloc(_CTX_ENTROPY)
        self._ctx     = ffi.malloc(_CTX_CTR_DRBG)
        try:
            _entropy_init(self._entropy)
            _drbg_init(self._ctx)
            pers = _to_bytes(personalization)
            _check(_drbg_seed(self._ctx, _entropy_func, self._entropy, pers, len(pers)), "ctr_drbg_seed")
        except MbedError:
            self.free()
            raise

    def read(self, n):
        """Return n cryptographically random bytes."""
        if self._freed:
            raise MbedError("Random context already freed")
        buf = ffi.malloc(n)
        try:
            _check(_drbg_random(self._ctx, buf, n), "ctr_drbg_random")
            return ffi.buffer_to_bytes(buf, n)
        finally:
            ffi.free(buf)

    def randint(self, lo, hi):
        """Return a random integer in [lo, hi]."""
        span = hi - lo + 1
        raw  = self.read(8)
        val  = 0
        for i in range(8):
            val = val + raw[i] * (256 ** i)
        return lo + (val % span)

    def free(self):
        if not self._freed:
            if self._ctx:
                _drbg_free(self._ctx)
                ffi.free(self._ctx)
                self._ctx = None
            if self._entropy:
                _entropy_free(self._entropy)
                ffi.free(self._entropy)
                self._entropy = None
            self._freed = True

# ==============================================================================
#  X.509 Certificate Store
# ==============================================================================

class CertStore:
    """
    Holds one or more X.509 certificates (CA chain or leaf certs).

    Usage:
        ca = CertStore()
        ca.load_pem(open("ca-bundle.pem", "rb").read())
        ca.free()
    """
    def __init__(self):
        _require_x509()
        self._freed = False
        self._ctx   = ffi.malloc(_CTX_X509_CRT)
        _x509_init(self._ctx)

    def load_pem(self, pem_data):
        """Load a PEM-encoded certificate or chain."""
        data = _to_bytes(pem_data)
        # mbedtls requires a null terminator for PEM
        if data[-1] != 0:
            data = data + b'\x00'
        _check(_x509_parse(self._ctx, data, len(data)), "x509_crt_parse")

    def load_file(self, path):
        """Load a certificate from a file path."""
        _check(_x509_parse_file(self._ctx, path.encode("utf-8")), "x509_crt_parse_file")

    def info(self):
        """Return a human-readable string describing the certificate chain."""
        buf = ffi.malloc(4096)
        try:
            _x509_info(buf, 4096, b"  ", self._ctx)
            raw = ffi.buffer_to_bytes(buf, 4096)
            # Trim at first null byte
            text = ""
            for b in raw:
                if b == 0:
                    break
                text = text + chr(b)
            return text
        finally:
            ffi.free(buf)

    def free(self):
        if not self._freed and self._ctx:
            _x509_free(self._ctx)
            ffi.free(self._ctx)
            self._ctx = None
            self._freed = True

# ==============================================================================
#  TLS Client Context
# ==============================================================================

class TLSClient:
    """
    TLS client context for wrapping a raw TCP socket file descriptor.

    Usage:
        rng  = Random()
        ca   = CertStore()
        ca.load_pem(ca_pem_bytes)

        tls = TLSClient(rng, ca, hostname="example.com")
        tls.set_fd(sock_fd)      # raw OS file descriptor
        tls.handshake()
        tls.write(b"GET / HTTP/1.0\r\n\r\n")
        data = tls.read(4096)
        tls.close()
        tls.free()
        ca.free()
        rng.free()
    """

    # MBEDTLS_SSL_VERIFY_REQUIRED = 2
    VERIFY_NONE     = 0
    VERIFY_OPTIONAL = 1
    VERIFY_REQUIRED = 2

    def __init__(self, rng, ca_store=None, hostname=None,  verify=2):
        _require_tls()
        self._freed    = False
        self._rng      = rng
        self._hostname = hostname
        self._ssl      = ffi.malloc(_CTX_SSL)
        self._cfg      = ffi.malloc(_CTX_SSL_CONFIG)

        try:
            _ssl_init(self._ssl)
            _ssl_cfg_init(self._cfg)

            _check(_ssl_cfg_defaults(self._cfg, MBEDTLS_SSL_IS_CLIENT, MBEDTLS_SSL_TRANSPORT_STREAM, MBEDTLS_SSL_PRESET_DEFAULT), "ssl_config_defaults")

            # Hook in our CSPRNG
            if _ssl_conf_rng:
                _ssl_conf_rng(self._cfg, _drbg_random, rng._ctx)

            # Auth mode
            _ssl_conf_authmode(self._cfg, verify)

            # CA chain for peer verification
            if ca_store is not None:
                _ssl_conf_ca_chain(self._cfg, ca_store._ctx, ffi.c_void_p(0))

            _check(_ssl_setup(self._ssl, self._cfg), "ssl_setup")

            if hostname is not None:
                _check(_ssl_set_hostname(self._ssl, hostname.encode("utf-8")), "ssl_set_hostname")
        except MbedError:
            self.free()
            raise

    def set_bio(self, send_fn, recv_fn, ctx_ptr=None):
        """
        Attach custom send/recv callbacks (ffi callback objects).
        For fd-based sockets use set_fd() instead.
        """
        p = ctx_ptr if ctx_ptr is not None else ffi.c_void_p(0)
        _ssl_set_bio(self._ssl, p, send_fn, recv_fn, ffi.c_void_p(0))

    def set_fd(self, fd):
        """
        Use a raw OS socket file descriptor for I/O.
        Requires mbedtls_net_set_fd or the built-in net send/recv.
        """
        # mbedtls ships mbedtls_net_send and mbedtls_net_recv which take
        # the fd (cast to void*) as the context pointer.
        _net_send = _lib_tls.mbedtls_net_send([ffi.c_void_p, ffi.c_char_p, ffi.c_uint64], ffi.c_int32)
        _net_recv = _lib_tls.mbedtls_net_recv([ffi.c_void_p, ffi.c_char_p, ffi.c_uint64], ffi.c_int32)
        _ssl_set_bio(self._ssl, ffi.c_void_p(fd), _net_send, _net_recv, ffi.c_void_p(0))

    def handshake(self):
        """Perform the TLS handshake. Call after set_fd/set_bio."""
        ret = _ssl_handshake(self._ssl)
        if ret != 0:
            raise MbedError(format_str("TLS handshake failed: {ret:#010x}"))

    def get_verify_result(self):
        """Return the peer certificate verification flags (0 = ok)."""
        return _ssl_get_verify_result(self._ssl)

    def write(self, data):
        """Send data over TLS. Returns number of bytes written."""
        data = _to_bytes(data)
        ret  = _ssl_write(self._ssl, data, len(data))
        if ret < 0:
            raise MbedError(format_str("TLS write failed: {ret:#010x}"))
        return ret

    def read(self, size=4096):
        """Read up to size bytes from TLS stream."""
        buf = ffi.malloc(size)
        try:
            ret = _ssl_read(self._ssl, buf, size)
            if ret < 0:
                raise MbedError(format_str("TLS read failed: {ret:#010x}"))
            return ffi.buffer_to_bytes(buf, ret)
        finally:
            ffi.free(buf)

    def close(self):
        """Send TLS close_notify alert."""
        _ssl_close(self._ssl)

    def free(self):
        if not self._freed:
            if self._ssl:
                _ssl_free(self._ssl)
                ffi.free(self._ssl)
                self._ssl = None
            if self._cfg:
                _ssl_cfg_free(self._cfg)
                ffi.free(self._cfg)
                self._cfg = None
            self._freed = True

# ==============================================================================
#  Convenience One-Shot Functions
# ==============================================================================

def _oneshot(algorithm, data):
    h = Hash(algorithm, _to_bytes(data))
    try:
        return h.digest()
    finally:
        h.free()

def _oneshot_hex(algorithm, data):
    h = Hash(algorithm, _to_bytes(data))
    try:
        return h.hexdigest()
    finally:
        h.free()

def sha256(data=b''):
    return Hash("SHA256", _to_bytes(data))

def sha512(data=b''):
    return Hash("SHA512", _to_bytes(data))

def sha1(data=b''):
    return Hash("SHA1", _to_bytes(data))

def md5(data=b''):
    return Hash("MD5", _to_bytes(data))

def sha256_digest(data):
    return _oneshot("SHA256", data)

def sha512_digest(data):
    return _oneshot("SHA512", data)

def sha256_hex(data):
    return _oneshot_hex("SHA256", data)

def sha512_hex(data):
    return _oneshot_hex("SHA512", data)

def hmac_sha256(key, data=b''):
    return HMAC("SHA256", key, _to_bytes(data))

def hmac_sha512(key, data=b''):
    return HMAC("SHA512", key, _to_bytes(data))

def hmac_sha256_digest(key, data):
    h = HMAC("SHA256", key, _to_bytes(data))
    try:
        return h.digest()
    finally:
        h.free()

def hmac_sha256_hex(key, data):
    h = HMAC("SHA256", key, _to_bytes(data))
    try:
        return h.hexdigest()
    finally:
        h.free()

def random_bytes(n):
    """Return n cryptographically random bytes using a one-shot RNG."""
    rng = Random()
    try:
        return rng.read(n)
    finally:
        rng.free()

# Backwards-compat alias matching the original test signature: new_hmac(key, data, algo)
def new_hmac(key, data=b'', algorithm="SHA256"):
    return HMAC(algorithm, key, _to_bytes(data))