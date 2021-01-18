#!/usr/bin/env python3
import sys
import ozone.runner

if len(sys.argv) < 2:
    print("Arg error. Usage: oz <command>")
    exit(0)

secret_filename = os.getenv('SECRETS_FILE')
start_index = 1
if secret_filename == None:
    secret_filename = sys.argv[1]
    print("No SECRETS_FILE set. Using: " + secret_filename)
    start_index = 2

print("hello")