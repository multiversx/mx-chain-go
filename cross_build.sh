#!/usr/bin/env bash

package=node
package_split=(${package//\// })
package_name=${package_split[-1]}

platforms=("windows/amd64" "darwin/amd64" "linux/amd64" "linux/arm64")
BASEDIR=$(pwd)/cross_build
APP_VER=$(git describe --tags --long --dirty)
echo "##teamcity[testSuiteStarted name='Cross_Build']"
for platform in "${platforms[@]}"
do
    echo "##teamcity[testStarted name='$platform' captureStandardOutput='true']"
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name=$BASEDIR/$GOOS/$GOARCH/$package_name
    echo "Building $output_name"
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi

    pushd cmd/node
    env GOOS=$GOOS GOARCH=$GOARCH go build -o $output_name -a -i -v -ldflags="-X main.appVersion=$APP_VER"
    if [ $? -ne 0 ]; then
        echo 'An error has occurred!'
	echo "##teamcity[testFailed type='comparisonFailure' name='$platform']"
    fi
    popd
    echo "##teamcity[testFinished name='$platform']"
done
echo "##teamcity[testSuiteFinished name='Cross_Build']"
