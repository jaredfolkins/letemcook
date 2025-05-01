#!/bin/sh

set -e

find . -type f \( -name "*.go" -o -name "*.templ" \) -print0 | xargs -0 perl -i -ne 'print unless /^\s*\/\/(?!go:embed)/';

echo "Comment stripping complete." 