mkdir "${GOPATH}"/src/github.com/gogo

cd "${GOPATH}"/src/github.com/gogo

git clone https://github.com/gogo/protobuf.git

cd protobuf

git fetch origin pull/659/head:casttypewith

git checkout casttypewith

PROTO_COMPILER=protoc-3.17.3-linux-x86_64.zip
TEMP_FOLDER_NAME=temp-001

mkdir ~/${TEMP_FOLDER_NAME}

cd ~/${TEMP_FOLDER_NAME}

wget https://github.com/protocolbuffers/protobuf/releases/download/v3.17.3/${PROTO_COMPILER}
unzip ${PROTO_COMPILER}

sudo cp -rf include/google /usr/include

sudo cp -f bin/protoc /usr/bin

sudo chmod +x /usr/bin/protoc

rm -r ~/${TEMP_FOLDER_NAME}

cd "${GOPATH}"/src/github.com/ElrondNetwork

git clone https://github.com/ElrondNetwork/protobuf.git

cd protobuf/protoc-gen-gogoslick

go build

sudo cp protoc-gen-gogoslick /usr/bin/


