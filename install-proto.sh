
PATH_GITHUB_GOGO="${GOPATH}"/src/github.com/gogo
if [ ! -d "${PATH_GITHUB_GOGO}" ]
then
  mkdir -p "${PATH_GITHUB_GOGO}"
fi
cd "${PATH_GITHUB_GOGO}"


if [ ! -d "protobuf" ]
then
  echo "Cloning gogo/protobuf..."
  git clone https://github.com/gogo/protobuf.git
fi

cd protobuf
DRY_RUN_RESULT=$(git fetch --dry-run origin pull/659/head:casttypewith 2>&1)
if [[ "${DRY_RUN_RESULT}" != "fatal"* ]]
then
  git fetch origin pull/659/head:casttypewith
fi

git checkout casttypewith

PROTO_COMPILER=protoc-3.17.3-linux-x86_64.zip
TEMP_FOLDER_NAME=temp-proto-001
TEMP_LOCATION=~/temp/${TEMP_FOLDER_NAME}
mkdir -p ${TEMP_LOCATION}
cd "${TEMP_LOCATION}"
echo "Downloading protobuf compiler v3.17.3 ..."
wget -q https://github.com/protocolbuffers/protobuf/releases/download/v3.17.3/${PROTO_COMPILER}
unzip -q "${PROTO_COMPILER}" -d ./
sudo cp -rf include/google /usr/include
sudo cp -f bin/protoc /usr/bin
sudo chmod +x /usr/bin/protoc

echo "Removing temporally files ..."
rm -r ${TEMP_LOCATION}

cd "${GOPATH}"/src/github.com/multiversx

if [ ! -d "protobuf" ]
then
  echo "Cloning multiversx/protobuf..."
  git clone https://github.com/multiversx/protobuf.git
fi

echo "Building protoc-gen-gogoslick binary..."
cd protobuf/protoc-gen-gogoslick
go build
sudo cp -f protoc-gen-gogoslick /usr/bin/

echo "Done."

