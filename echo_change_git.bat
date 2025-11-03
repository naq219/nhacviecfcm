#!/bin/bash

# File xuất ra
output="all_changed_files.txt"
> "$output"  # xóa nội dung cũ nếu có

# Lấy danh sách file đã thay đổi (modified, staged)
files=$(git diff --name-only)

for f in $files; do
    if [ -f "$f" ]; then
        echo "-----start $f----" >> "$output"
        cat "$f" >> "$output"
        echo "-----end $f----" >> "$output"
    fi
done

echo "Done! All changed files are in $output"
