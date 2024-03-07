import os
import pathlib
import subprocess

GOPATH = os.getenv("GOPATH", "")

for file in pathlib.Path(f"{GOPATH}/pkg/mod/github.com/multiversx").rglob("libwasmer_darwin_amd64.dylib"):
    print(f"Running install_name_tool on: {file}")
    subprocess.check_output(f"sudo install_name_tool -id @rpath/libwasmer_darwin_amd64.dylib {file}", shell=True)
