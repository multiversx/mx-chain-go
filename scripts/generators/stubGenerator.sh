#!/bin/bash
# This script generates a stub from a given interface
# Mandatory parameters needed:
#   interface name from an interface.go file
#   path to the directory of interface.go file, from the directory this script is called
#   path to the destination directory the stub will be created, from the directory this script is called
#
# Usage example: bash stubGenerator.sh EnableEpochsHandler ../../common ../../common

extractPackageName() {
  if [ "$stubDir" == "." ]; then
    stubDir=$(pwd)
  fi

  packageName=${stubDir##*"/"}
  # handle case when / is provided at the end of the path
  if [ ${#stubName} == 0 ]; then
    withoutLastSlash=${stubDir%"/"}
    packageName=${withoutLastSlash##*"/"}
  fi
}

readInterfaceFile() {
  # read each file line until the specified interface is found
  # when the declaration of the interface is found, start reading all methods until }
  interfaceFound=false
  while IFS= read -r line
  do
    if [[ "$line" == *"type $interfaceName interface"* ]]; then
      { echo -e "package $packageName\n";
        echo -e "// $stubName -";
        echo "type $stubName struct {";
        } >> "$stubPath"
      isInterfaceMethod=true
      interfaceFound=true
      continue
    fi

    # continue reading until interface is found
    [ $isInterfaceMethod == false ] && continue

    #skip empty lines inside interface
    lineWithNoSpaces=${line//[[:blank:]]/}
    [[ ${#lineWithNoSpaces} == 0 ]] && continue

    # if end of interface declaration is hit, reset the flag and break the while loop
    if [[ "$line" == "}" ]]; then
      isInterfaceMethod=false
      break
    fi

    # append methods lines to methodsArr which is an array with strings of each method line
    if $isInterfaceMethod; then
      # if there is no bracket, it means that this interface extends another one
      # ignore it for the moment
      [[ "$line" != *"("* ]] && continue

      methodsArr+=("$line")
    fi
  done < "$filePath"

  [ $interfaceFound == false ] && echo "Interface $interfaceName DOES NOT exists in $filePath." && exit
}

removeCommentsFromMethodLine() {
  if [[ "$method" == *"//"* ]]; then
    method=${method%%"//"*}
  fi
}

createStubStructure() {
  # navigate through all methods lines and create stub members with Called suffix and write them to the dest file
  for method in "${methodsArr[@]}"
  do
    [[ $method == *"IsInterfaceNil"* ]] && continue

    methodName=${method%%"("*}
    methodName=${methodName//[[:blank:]]/}
    # ignore commented lines
    [[ "$methodName" == "//"* ]] && continue

    removeCommentsFromMethodLine

    replacementStr=$methodName"Called func("
    pattern="$methodName("
    structMember=${method//"$pattern"/"$replacementStr"}
    echo "$structMember" >> "$stubPath"
  done
  # now stub struct is complete, close it
  echo -e "}\n" >> "$stubPath"
}

extractReturnTypes() {
  # extract return types from method line into returnTypesArr array
  IFS=", "
  read -ra ADDR <<< "$rawReturnTypes"
  if [[ ${#ADDR[@]} != 0 ]]; then
    for i in "${ADDR[@]}"; do
      returnTypesArr+=("$i")
    done
  fi
}

extractParametersAndTypes() {
  # extract parameters from method line into:
  #   paramNames, which will be an array of strings used to call stub method
  #   paramTypes, which will be an array of strings exactly how the params types are. Eg. bool, error, uint32, etc.
  IFS=','
  read -ra ADDR <<< "$rawParams"

  if [[ ${#ADDR[@]} != 0 ]]; then
    for i in "${!ADDR[@]}"; do
      IFS=' '
      read -r paramName paramType optParamType<<< "${ADDR[$i]}"

      # handle variable number of arguments
      if [[ "$paramType" == *"..."* ]]; then
        paramName+="..."
      fi

      # handle missing param name
      # eg. (string, string)
      # in this case param name will be tmp concatenated with the index of the param
      if [ ${#paramType} == 0 ]; then
        paramType=$paramName
        paramName=tmp$i
      fi

      # handle chan param type
      if [ "$paramType" == "chan" ] && [ ${#optParamType} != 0 ]; then
        paramType+=" "$optParamType
      fi

      if [ "$paramName" == "chan" ]; then
        paramType=$paramName" "$paramType
        paramName=tmp$i
      fi

      paramNames+=("$paramName")
      paramTypes+=("$paramType")
    done
  fi
}

computeUpdatedParameters() {
  updatedParameters+="("
  for i in "${!paramNames[@]}"; do
    updatedParameters+="${paramNames[$i]%%"."*}" # until first dot, in case there are variable params
    updatedParameters+=" "
    updatedParameters+="${paramTypes[$i]}"
    updatedParameters+=", "
  done

  if [[ ${#updatedParameters} != 1 ]]; then
    # remove the last ", " added with last param
    updatedParameters=${updatedParameters::-2}
  fi

  updatedParameters+=")"
}

writeWithNoReturn() {
  { echo "stub.$stubField($stringParamNames)";
    echo "}";
    echo -e "}\n";
    } >> "$stubPath"
}

extractDefaultReturn() {
  for i in "${returnTypesArr[@]}"; do
    case $i in
      bool)
        toReturn+="false, "
        ;;
      string | core.PeerID)
        toReturn+="\"\", "
        ;;
      int | uint32 | uint64 | int32 | int64 | float | float32 | float64 | byte)
        toReturn+="0, "
        ;;
      *)
        toReturn+="nil, "
        ;;
    esac
  done

  if [[ ${#toReturn} != 0 ]]; then
    # remove the last ", " added with last param
    toReturn=${toReturn::-2}
  fi
}

writeWithReturn() {
  { echo "return stub.$stubField($stringParamNames)";
  echo "}";
  } >> "$stubPath"

  # compute default values to return when stub member is not provided, separated by comma
  toReturn=""
  extractDefaultReturn

  # write the final return statement to file with default params and close the method
  { echo "return $toReturn";
    echo -e "}\n";
    } >> "$stubPath"
}

getStringParamNames() {
  for i in "${paramNames[@]}"; do
    stringParamNames+=$i
    stringParamNames+=", "
  done

  if [[ ${#stringParamNames} != 0 ]]; then
    # remove the last ", " added with last param
    stringParamNames=${stringParamNames::-2}
  fi
}

createMethodBody() {
  # if method is IsInterfaceNil, write special return and return
  if [[ $methodName == *"IsInterfaceNil"* ]]; then
    { echo "return stub == nil";
      echo -e "}\n";
      } >> "$stubPath"
    return
  fi

  # add the check to stub member to not be nil
  echo "if stub.$stubField != nil {" >> "$stubPath"

  stringParamNames=""
  getStringParamNames

  # add return statement calling stub member
  # if there is no return type, add it without return
  # otherwise, return it with the provided params
  if [[ ${#returnTypesArr} == 0 ]]; then
    writeWithNoReturn
  else
    writeWithReturn
  fi
}

createStubMethods() {
  # navigate through all methods lines and:
  #   extract method name
  #   extract return types, used to handle the return
  #   extract parameters, used to call the stub member
  for method in "${methodsArr[@]}"
    do
      methodName=${method%%"("*}
      methodName=${methodName//[[:blank:]]/}
      # ignore commented lines
      [[ "$methodName" == "//"* ]] && continue

     removeCommentsFromMethodLine

      rawReturnTypesWithBraces=${method#*")"}
      rawReturnTypesWithBraces=${rawReturnTypesWithBraces%"{"*}
      # remove () in case of multiple return types
      rawReturnTypes=${rawReturnTypesWithBraces#*"("}
      rawReturnTypes=${rawReturnTypes%")"*}

      declare -a returnTypesArr=()
      extractReturnTypes

      rawParams=${method%%")"*}
      rawParams=${rawParams##*"("}

      declare -a paramNames=()
      declare -a paramTypes=()
      extractParametersAndTypes

      # compute the stub member which will be called and write to the file:
      #   the comment
      #   the method signature
      # But first we compute the updated parameters, to avoid situation when param name is missing
      updatedParameters=""
      computeUpdatedParameters

      stubField=$methodName"Called"
      { echo "// $methodName -";
        echo "func (stub *$stubName) $methodName $updatedParameters $rawReturnTypesWithBraces {";
        } >> "$stubPath"

      createMethodBody
  done
}

generateStub() {
  interfaceName=$1
  filePath=$2"/interface.go"
  stubDir=$3

  [ ! -d "$2" ] && echo "Source directory for interface DOES NOT exists." && exit
  [ ! -f "$filePath" ] && echo "Source interface.go file DOES NOT exists." && exit
  [ ! -d "$stubDir" ] && echo "Destination directory DOES NOT exists." && exit

  extractPackageName

  stubName=$interfaceName"Stub"

  # make first char of the file name lowercase
  firstChar=${stubName::1}
  firstChar=${firstChar,,}

  lenOfStubName=${#stubName}
  stubFileName=$firstChar${stubName:1:$lenOfStubName}

  stubPath="$stubDir/$stubFileName.go"
  rm -rf "$stubPath"

  isInterfaceMethod=false
  declare -a methodsArr

  readInterfaceFile
  createStubStructure
  createStubMethods

  # go fmt file
  go fmt "$stubPath"
}

if [ $# -eq 3 ]; then
  generateStub "$@"
else
  echo "Please use the following format..."
  echo "bash stubGenerator.sh interface_name path_to_interface.go_dir path_to_stub_destionation_dir"
fi
