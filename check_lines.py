#!/usr/bin/env python3
import sys

import sys
if len(sys.argv) > 1:
    filepath = sys.argv[1]
else:
    filepath = r"D:\\vista-app\\internal\\app\\core\\config_loader.go"
violations = []

with open(filepath, 'r', encoding='utf-8') as f:
    for i, line in enumerate(f, 1):
        line_len = len(line.rstrip('\r\n'))
        if line_len > 79:
            violations.append(f"Line {i}: {line_len} chars")

if violations:
    print("LINE LENGTH VIOLATIONS:")
    for v in violations:
        print(f"  [X] {v}")
    sys.exit(1)
else:
    print("[OK] All lines comply with 79 character limit")
    sys.exit(0)
