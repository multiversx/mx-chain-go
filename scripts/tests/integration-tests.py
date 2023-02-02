import os
import subprocess
import sys

NUM_CHUNKS = int(8)


def main():
    integration_test_path = "../../integrationTests"
    sub_folders = [f.path for f in os.scandir(integration_test_path) if f.is_dir()]

    chunk_size = int(len(sub_folders) / NUM_CHUNKS)

    chunked_list = list()

    for i in range(0, len(sub_folders), chunk_size):
        chunked_list.append(sub_folders[i:i + chunk_size])

    chunk_index = 0
    if len(sys.argv) > 1:
        chunk_index = sys.argv[1]

    current_chunk = chunked_list[int(chunk_index)]
    packages = ""
    for i in range(0, len(current_chunk), 1):
        packages = packages + current_chunk[i] + "/..."
        if i != len(current_chunk)-1:
            packages += " "

    print("running integration tests packages:", packages)
    process = subprocess.Popen(["go", "test", "-parallel", "2"] + packages.split(), stdout=subprocess.PIPE, stderr=subprocess.PIPE, bufsize=1, universal_newlines=True)
    stdout, stderr = process.communicate()

    if process.returncode != 0:
        print("Error:", stderr)
    else:
        print(stdout)

    process.wait()


if __name__ == "__main__":
    main()
