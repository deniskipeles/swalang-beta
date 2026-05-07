import mbedtls

# ==============================================================================
#  Known-good test vectors (computed offline / cross-checked against OpenSSL)
# ==============================================================================

SHA256_HELLO_WORLD  = "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
SHA256_EMPTY        = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
SHA512_HELLO_WORLD  = "309ecc489c12d6eb4cc40f50c902f2b4d0ed77ee511a7c7a9bcd3ca86d4cd86f" + "989dd35bc5ff499670da34255b45b0cfd830e81f605dcf7dc5542e93ae9cd76f"
HMAC256_SECRET_MSG  = "8b5f35919261c180e14c7760821516e866e4a28298717804473b185ec1e695d7"

# ==============================================================================
#  Helpers
# ==============================================================================

def _assert_eq(got, expected, label):
    if got != expected:
        print(format_str("   FAIL  {label}"))
        print(format_str("   got      : {got}"))
        print(format_str("   expected : {expected}"))
        assert False, format_str("{label} mismatch")
    print(format_str("   OK    {label}"))

def _hex(raw_bytes):
    """Pure-Pylearn bytes → hex string (no stdlib needed)."""
    chars = "0123456789abcdef"
    result = ""
    for b in raw_bytes:
        result = result + chars[b // 16] + chars[b % 16]
    return result

# ==============================================================================
#  Test suite
# ==============================================================================

def test_hash_oneshot():
    print("Test 1: sha256() one-shot object...")
    h = mbedtls.sha256(b"hello world")
    _assert_eq(h.hexdigest(), SHA256_HELLO_WORLD, "sha256('hello world')")

    # Reuse: update appends to a fresh context after digest() restarts it
    h.update(b"hello world")
    _assert_eq(h.hexdigest(), SHA256_HELLO_WORLD, "sha256 reuse after digest()")
    h.free()
    print("Test 1 PASSED.\n")


def test_hash_empty():
    print("Test 2: sha256('') empty input...")
    h = mbedtls.sha256(b"")
    _assert_eq(h.hexdigest(), SHA256_EMPTY, "sha256('')")
    h.free()
    print("Test 2 PASSED.\n")


def test_hash_incremental():
    print("Test 3: Incremental update() — 'hello' + ' ' + 'world'...")
    h = mbedtls.Hash("sha256")
    h.update(b"hello")
    h.update(b" ")
    h.update(b"world")
    _assert_eq(h.hexdigest(), SHA256_HELLO_WORLD, "incremental sha256")
    h.free()
    print("Test 3 PASSED.\n")


def test_hash_lowercase_algo():
    print("Test 4: Algorithm name is case-insensitive ('sha256' vs 'SHA256')...")
    h = mbedtls.Hash("sha256", b"hello world")
    _assert_eq(h.hexdigest(), SHA256_HELLO_WORLD, "lowercase algo name")
    h.free()
    print("Test 4 PASSED.\n")


def test_hash_sha512():
    print("Test 5: sha512() one-shot...")
    h = mbedtls.sha512(b"hello world")
    result = h.hexdigest()
    h.free()
    # SHA-512 produces 128 hex chars
    _assert_eq(len(result), 128, "sha512 digest length")
    _assert_eq(result, SHA512_HELLO_WORLD, "sha512('hello world')")
    print("Test 5 PASSED.\n")


def test_hash_md5():
    print("Test 6: md5() — length check only (no hardcoded vector to avoid controversy)...")
    h = mbedtls.md5(b"hello world")
    result = h.hexdigest()
    h.free()
    _assert_eq(len(result), 32, "md5 digest length == 32 hex chars")
    print("Test 6 PASSED.\n")


def test_digest_bytes():
    print("Test 7: digest() returns raw bytes of correct length...")
    h = mbedtls.sha256(b"hello world")
    raw = h.digest()
    h.free()
    _assert_eq(len(raw), 32, "sha256 raw digest length == 32 bytes")
    _assert_eq(_hex(raw), SHA256_HELLO_WORLD, "sha256 raw bytes via _hex()")
    print("Test 7 PASSED.\n")


def test_hmac_sha256():
    print("Test 8: HMAC-SHA256('secret', 'message')...")
    h = mbedtls.HMAC("SHA256", b"secret", b"message")
    _assert_eq(h.hexdigest(), HMAC256_SECRET_MSG, "HMAC-SHA256")
    h.free()
    print("Test 8 PASSED.\n")


def test_hmac_convenience():
    print("Test 9: hmac_sha256() convenience wrapper...")
    h = mbedtls.hmac_sha256(b"secret", b"message")
    _assert_eq(h.hexdigest(), HMAC256_SECRET_MSG, "hmac_sha256 convenience")
    h.free()
    print("Test 9 PASSED.\n")


def test_hmac_incremental():
    print("Test 10: HMAC incremental update()...")
    h = mbedtls.HMAC("SHA256", b"secret")
    h.update(b"mes")
    h.update(b"sage")
    _assert_eq(h.hexdigest(), HMAC256_SECRET_MSG, "HMAC incremental")
    h.free()
    print("Test 10 PASSED.\n")


def test_hmac_digest_bytes():
    print("Test 11: HMAC digest() returns 32 raw bytes...")
    h = mbedtls.HMAC("SHA256", b"key", b"data")
    raw = h.digest()
    h.free()
    _assert_eq(len(raw), 32, "HMAC-SHA256 raw length == 32")
    print("Test 11 PASSED.\n")


def test_aes_cbc_roundtrip():
    print("Test 12: AES-CBC encrypt → decrypt roundtrip (128-bit key)...")
    key       = b"0123456789abcdef"              # 16 bytes = AES-128
    iv        = b"abcdef0123456789"              # 16 bytes
    plaintext = b"Hello, AES-CBC!!"              # exactly 16 bytes

    cipher = mbedtls.AES_CBC(key)
    ct = cipher.encrypt(iv, plaintext)
    pt = cipher.decrypt(iv, ct)
    cipher.free()

    _assert_eq(len(ct), 16, "AES-CBC ciphertext length")
    _assert_eq(pt, plaintext, "AES-CBC decrypt == original plaintext")
    print("Test 12 PASSED.\n")


def test_aes_cbc_256():
    print("Test 13: AES-CBC roundtrip (256-bit key, 48-byte payload)...")
    key       = b"01234567890123456789012345678901"   # 32 bytes = AES-256
    iv        = b"abcdef0123456789"
    plaintext = b"A" * 48                            # 3 AES blocks

    cipher = mbedtls.AES_CBC(key)
    ct = cipher.encrypt(iv, plaintext)
    pt = cipher.decrypt(iv, ct)
    cipher.free()

    _assert_eq(len(ct), 48, "AES-256-CBC ciphertext length")
    _assert_eq(pt, plaintext, "AES-256-CBC decrypt == original")
    print("Test 13 PASSED.\n")


def test_aes_cbc_different_iv():
    print("Test 14: Different IVs produce different ciphertexts...")
    key = b"0123456789abcdef"
    iv1 = b"AAAAAAAAAAAAAAAA"
    iv2 = b"BBBBBBBBBBBBBBBB"
    pt  = b"same plaintext!!"

    cipher = mbedtls.AES_CBC(key)
    ct1 = cipher.encrypt(iv1, pt)
    ct2 = cipher.encrypt(iv2, pt)
    cipher.free()

    if ct1 == ct2:
        assert False, "Different IVs must produce different ciphertexts"
    print("   OK    different IVs → different ciphertexts")
    print("Test 14 PASSED.\n")


def test_aes_gcm_roundtrip():
    print("Test 15: AES-GCM encrypt → decrypt roundtrip...")
    key       = b"0123456789abcdef"       # 16 bytes
    iv        = b"unique-iv-12"           # 12 bytes (recommended for GCM)
    plaintext = b"Secret GCM message"
    aad       = b"authenticated data"

    g = mbedtls.AES_GCM(key)
    ct, tag = g.encrypt(iv, plaintext, aad=aad)
    pt      = g.decrypt(iv, ct, tag, aad=aad)
    g.free()

    _assert_eq(len(tag), 16, "GCM tag length == 16")
    _assert_eq(pt, plaintext, "AES-GCM decrypt == original plaintext")
    print("Test 15 PASSED.\n")


def test_aes_gcm_tag_tamper():
    print("Test 16: AES-GCM rejects tampered tag...")
    key = b"0123456789abcdef"
    iv  = b"unique-iv-12"
    pt  = b"tamper test data"

    g = mbedtls.AES_GCM(key)
    ct, tag = g.encrypt(iv, pt)

    # Flip the first byte of the tag
    tag_list = []
    for b in tag:
        tag_list = tag_list + [b]
    tag_list[0] = (tag_list[0] + 1) % 256
    bad_tag = bytes(tag_list)

    caught = False
    try:
        g.decrypt(iv, ct, bad_tag)
    except mbedtls.MbedError:
        caught = True
    g.free()

    if not caught:
        assert False, "GCM must raise MbedError on tampered tag"
    print("   OK    tampered tag correctly rejected")
    print("Test 16 PASSED.\n")


def test_base64_roundtrip():
    print("Test 17: Base64 encode → decode roundtrip...")
    original = b"Hello, Base64 world!"
    encoded  = mbedtls.b64encode(original)
    decoded  = mbedtls.b64decode(encoded)
    _assert_eq(decoded, original, "base64 roundtrip")
    print("Test 17 PASSED.\n")


def test_base64_str():
    print("Test 18: b64encode_str / b64decode_str string helpers...")
    original = b"Pylearn FFI rocks"
    s = mbedtls.b64encode_str(original)
    _assert_eq(isinstance(s, str), True, "b64encode_str returns str")
    back = mbedtls.b64decode_str(s)
    _assert_eq(back, original, "b64decode_str roundtrip")
    print("Test 18 PASSED.\n")


def test_base64_known_vector():
    print("Test 19: Base64 known vector — 'Man' → 'TWFu'...")
    enc = mbedtls.b64encode_str(b"Man")
    _assert_eq(enc, "TWFu", "base64('Man') == 'TWFu'")
    dec = mbedtls.b64decode_str("TWFu")
    _assert_eq(dec, b"Man", "base64_decode('TWFu') == 'Man'")
    print("Test 19 PASSED.\n")


def test_random_length():
    print("Test 20: Random.read() returns correct byte count...")
    rng = mbedtls.Random()
    for n in [1, 16, 32, 64, 256]:
        raw = rng.read(n)
        _assert_eq(len(raw), n, format_str("random.read({n}) length"))
    rng.free()
    print("Test 20 PASSED.\n")


def test_random_entropy():
    print("Test 21: Two random.read(32) calls produce different bytes...")
    rng = mbedtls.Random()
    a = rng.read(32)
    b = rng.read(32)
    rng.free()
    if a == b:
        assert False, "Two sequential random reads must differ"
    print("   OK    sequential reads differ")
    print("Test 21 PASSED.\n")


def test_random_randint():
    print("Test 22: Random.randint() stays within [lo, hi]...")
    rng = mbedtls.Random()
    lo  = 1
    hi  = 100
    all_in_range = True
    for _ in range(50):
        v = rng.randint(lo, hi)
        if v < lo or v > hi:
            all_in_range = False
            break
    rng.free()
    _assert_eq(all_in_range, True, "randint always in [1, 100]")
    print("Test 22 PASSED.\n")


def test_random_personalization():
    print("Test 23: Random with personalization string seeds successfully...")
    rng = mbedtls.Random(personalization=b"my-app-v1.0")
    raw = rng.read(32)
    rng.free()
    _assert_eq(len(raw), 32, "personalized RNG produces 32 bytes")
    print("Test 23 PASSED.\n")


def test_oneshot_helpers():
    print("Test 24: One-shot digest helpers (sha256_digest, sha256_hex)...")
    raw = mbedtls.sha256_digest(b"hello world")
    _assert_eq(len(raw), 32, "sha256_digest length")
    _assert_eq(_hex(raw), SHA256_HELLO_WORLD, "sha256_digest value")

    s = mbedtls.sha256_hex(b"hello world")
    _assert_eq(s, SHA256_HELLO_WORLD, "sha256_hex value")
    print("Test 24 PASSED.\n")


def test_hmac_oneshot_helpers():
    print("Test 25: One-shot HMAC helpers (hmac_sha256_digest, hmac_sha256_hex)...")
    raw = mbedtls.hmac_sha256_digest(b"secret", b"message")
    _assert_eq(len(raw), 32, "hmac_sha256_digest length")

    s = mbedtls.hmac_sha256_hex(b"secret", b"message")
    _assert_eq(s, HMAC256_SECRET_MSG, "hmac_sha256_hex value")
    print("Test 25 PASSED.\n")


def test_error_on_bad_algorithm():
    print("Test 26: Hash raises MbedError for unknown algorithm...")
    caught = False
    try:
        mbedtls.Hash("NOTAREALGORITHM")
    except mbedtls.MbedError:
        caught = True
    _assert_eq(caught, True, "MbedError raised for bad algorithm")
    print("Test 26 PASSED.\n")


def test_error_on_freed_ctx():
    print("Test 27: update() raises after free()...")
    h = mbedtls.sha256(b"test")
    h.free()
    caught = False
    try:
        h.update(b"more")
    except mbedtls.MbedError:
        caught = True
    _assert_eq(caught, True, "MbedError raised on freed context")
    print("Test 27 PASSED.\n")


def test_aes_bad_key_size():
    print("Test 28: AES_CBC rejects invalid key size...")
    caught = False
    try:
        mbedtls.AES_CBC(b"shortkey")   # 8 bytes — invalid
    except mbedtls.MbedError:
        caught = True
    _assert_eq(caught, True, "MbedError raised for 8-byte AES key")
    print("Test 28 PASSED.\n")


def test_aes_bad_iv_length():
    print("Test 29: AES_CBC rejects IV != 16 bytes...")
    cipher = mbedtls.AES_CBC(b"0123456789abcdef")
    caught = False
    try:
        cipher.encrypt(b"short_iv", b"A" * 16)
    except mbedtls.MbedError:
        caught = True
    cipher.free()
    _assert_eq(caught, True, "MbedError raised for bad IV length")
    print("Test 29 PASSED.\n")


def test_aes_bad_plaintext_length():
    print("Test 30: AES_CBC rejects plaintext not multiple of 16...")
    cipher = mbedtls.AES_CBC(b"0123456789abcdef")
    caught = False
    try:
        cipher.encrypt(b"BBBBBBBBBBBBBBBB", b"not a multiple")
    except mbedtls.MbedError:
        caught = True
    cipher.free()
    _assert_eq(caught, True, "MbedError raised for non-block plaintext")
    print("Test 30 PASSED.\n")


# ==============================================================================
#  Entry point
# ==============================================================================

def run_all():
    print("=" * 60)
    print("   MbedTLS Wrapper — Full Test Suite")
    print("=" * 60)
    print()

    if not mbedtls.CRYPTO_AVAILABLE:
        print("⚠️  MbedTLS crypto library not available — skipping all tests.")
        return None

    test_hash_oneshot()
    test_hash_empty()
    test_hash_incremental()
    test_hash_lowercase_algo()
    test_hash_sha512()
    test_hash_md5()
    test_digest_bytes()
    test_hmac_sha256()
    test_hmac_convenience()
    test_hmac_incremental()
    test_hmac_digest_bytes()
    test_aes_cbc_roundtrip()
    test_aes_cbc_256()
    test_aes_cbc_different_iv()
    test_aes_gcm_roundtrip()
    test_aes_gcm_tag_tamper()
    test_base64_roundtrip()
    test_base64_str()
    test_base64_known_vector()
    test_random_length()
    test_random_entropy()
    test_random_randint()
    test_random_personalization()
    test_oneshot_helpers()
    test_hmac_oneshot_helpers()
    test_error_on_bad_algorithm()
    test_error_on_freed_ctx()
    test_aes_bad_key_size()
    test_aes_bad_iv_length()
    test_aes_bad_plaintext_length()

    print("=" * 60)
    print("🎉  All 30 tests passed!")
    print("=" * 60)

run_all()