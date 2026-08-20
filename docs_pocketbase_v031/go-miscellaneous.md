# PocketBase Go Miscellaneous

This document covers miscellaneous features and utilities in PocketBase using Go.

## app.Store()

`app.Store()` provides a concurrent-safe, in-memory key-value store that persists for the lifetime of the application process. This is useful for caching, storing configuration flags, or sharing data between requests.

### Common Methods

- **Get(key):** Retrieves a value by key.
- **Set(key, value):** Stores a value with a key.
- **GetOrSet(key, setFunc):** Retrieves a value, or sets it using the provided function if it doesn't exist. The setter is invoked only once.

### Example

```go
app.Store().Set("example", 123)
v1 := app.Store().Get("example").(int) // 123

v2 := app.Store().GetOrSet("example2", func() any {
    // This setter is invoked only once unless "example2" is removed
    return 456
}).(int) // 456
```

**Note:** The application store is used internally by PocketBase with keys prefixed by `pb*`. Modifying these keys or calling `RemoveAll()`/`Reset()` can cause unintended side effects. For advanced use, consider creating your own store with `store.New[K, T](nil)`.

---

## Security Helpers

The `security` package provides various security utilities. The most commonly used are:

### Generating Random Strings

- `security.RandomString(length)` - Generates a cryptographically secure random string using the default alphabet.
- `security.RandomStringWithAlphabet(length, alphabet)` - Generates a random string using a custom alphabet.

```go
secret := security.RandomString(10) // e.g. a35Vdb10Z4
secret := security.RandomStringWithAlphabet(5, "1234567890") // e.g. 33215
```

### Constant-Time String Comparison

- `security.Equal(string1, string2)` - Compares two strings in constant time, preventing timing attacks.

```go
isEqual := security.Equal(hash1, hash2)
```

### AES Encrypt/Decrypt

- `security.Encrypt(data, key)` - Encrypts data using AES-256-CBC. The key must be a random 32-character string.
- `security.Decrypt(encryptedData, key)` - Decrypts data encrypted with `security.Encrypt`.

```go
const key = "KaNom0KbaT2i0PoOfJQGd34R3NVf6cRQ" // Must be a random 32-character string
encrypted, err := security.Encrypt([]byte("test"), key)
if err != nil {
    return err
}
decrypted := security.Decrypt(encrypted, key) // []byte("test")
```