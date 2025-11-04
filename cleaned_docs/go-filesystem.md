# PocketBase Go Filesystem

This document explains how to interact with the filesystem in PocketBase using Go.

## Overview

PocketBase provides an abstraction layer for file storage, supporting both local filesystem and S3. The programmatic access is available through `app.NewFilesystem()`.

**Important:** Always call `Close()` on the filesystem instance and any file readers to prevent resource leaks.

---

## Reading Files

- **Single File:** Use `GetReader(key)` to retrieve a file's content. The `key` is typically `collectionId/recordId/filename`.
- **Multiple Files:** Use `List(prefix)` to get all files matching a path prefix.

### Example: Reading a Record's Avatar

```go
record, err := app.FindAuthRecordByEmail("users", "test@example.com")
if err != nil {
    return err
}

// Construct the full file key
avatarKey := record.BaseFilesPath() + "/" + record.GetString("avatar")

// Initialize the filesystem
fsys, err := app.NewFilesystem()
if err != nil {
    return err
}
defer fsys.Close()

// Get a reader for the file
r, err := fsys.GetReader(avatarKey)
if err != nil {
    return err
}
defer r.Close()

// Read the content
content := new(bytes.Buffer)
_, err = io.Copy(content, r)
if err != nil {
    return err
}
```

---

## Saving Files

While there are direct methods like `Upload()`, `UploadFile()`, and `UploadMultipart()`, file persistence for collection records is usually handled automatically when you save the record model.

### Example: Attaching a File to a Record

```go
record, err := app.FindRecordById("articles", "RECORD_ID")
if err != nil {
    return err
}

// Create a file from a local path
f, err := filesystem.NewFileFromPath("/local/path/to/file")

// Other file sources:
// - filesystem.NewFileFromBytes(data, name)
// - filesystem.NewFileFromURL(ctx, url)
// - filesystem.NewFileFromMultipart(mh)

// Set the new file (this can be a single file or a slice of files)
// Old files are automatically deleted on successful save.
record.Set("yourFileField", f)

err = app.Save(record)
if err != nil {
    return err
}
```

---

## Deleting Files

Direct deletion is possible with `Delete(key)`, but similar to saving, it's often handled automatically when updating a record.

### Example: Removing a File from a Record

```go
record, err := app.FindRecordById("articles", "RECORD_ID")
if err != nil {
    return err
}

// To reset a file field (deleting all associated files), set it to nil
record.Set("yourFileField", nil)

// To remove a specific file from a multi-file field, use the "-" modifier
record.Set("yourFileField-", "example_file.txt")

err = app.Save(record)
if err != nil {
    return err
}
```