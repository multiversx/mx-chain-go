#!/usr/bin/env bash

packages=("node" "keygenerator")

platforms=("windows/amd64" "darwin/amd64" "linux/amd64" "linux/arm64")
BASEDIR=$(pwd)/cross_build
APP_VER=$(git describe --tags --long --dirty)
echo "##teamcity[testSuiteStarted name='Cross_Build']"
for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    echo "Building $output_name"
    for package_name in "${packages[@]}"
    do
        output_name=$BASEDIR/$GOOS/$GOARCH/$package_name
        if [ $GOOS = "windows" ]; then
            output_name+='.exe'
        fi

        echo "##teamcity[testStarted name='$package_name-$platform' captureStandardOutput='true']"
        pushd "cmd/$package_name"
        env GOOS=$GOOS GOARCH=$GOARCH go build -o "$output_name" -a -i -v -ldflags="-X main.appVersion=$APP_VER"
        if [ $? -ne 0 ]; then
            echo 'An error has occurred!'
            echo "##teamcity[testFailed type='comparisonFailure' name='$package_name-$platform']"
        fi
            echo "##teamcity[testFinished name='$package_name-$platform']"
        popd

	pushd "$BASEDIR/$GOOS/$GOARCH/"
	  tar czvf "$BASEDIR/elrond_""$APP_VER""_""$GOOS""_""$GOARCH"".tgz" *
	popd
    done
done
echo "##teamcity[testSuiteFinished name='Cross_Build']"
