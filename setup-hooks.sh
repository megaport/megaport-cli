#!/bin/sh

# Copy the pre-commit hook to the .git/hooks directory
cp hooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit