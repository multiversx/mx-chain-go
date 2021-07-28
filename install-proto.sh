
PATH_GITHUB_GOGO="${GOPATH}"/src/github.com/gogo
if [ ! -d "${PATH_GITHUB_GOGO}" ]
then
  mkdir "${PATH_GITHUB_GOGO}}"
fi
cd "${PATH_GITHUB_GOGO}"


if [ ! -d "protobuf" ]
then
  echo "Cloning gogo/protobuf..."
  git clone https://github.com/gogo/protobuf.git
  git fetch origin pull/659/head:casttypewith
fi

cd protobuf

git checkout casttypewith

PROTO_COMPILER=protoc-3.17.3-linux-x86_64.zip
TEMP_FOLDER_NAME=temp-001
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

cd "${GOPATH}"/src/github.com/ElrondNetwork

if [ ! -d "protobuf" ]
then
  echo "Cloning ElrondNetwork/protobuf..."
  git clone https://github.com/ElrondNetwork/protobuf.git
fi

echo "Building protoc-gen-gogoslick binary..."
cd protobuf/protoc-gen-gogoslick
go build
sudo cp -f protoc-gen-gogoslick /usr/bin/

echo "Done."

