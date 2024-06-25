#!/bin/bash
# This script generates a mock from a given interface
# Mandatory parameters needed:
#   interface name from an interface.go file
#   path to the directory of interface.go file, from the directory this script is called
#   path to the destination directory the mock will be created, from the directory this script is called
#
# Usage example: bash mockGenerator.sh EnableEpochsHandler ../../common ../../common

extractPackageName() {
  if [ "$mockDir" == "." ]; then
    mockDir=$(pwd)
  fi

  packageName=${mockDir##*"/"}
  # handle case when / is provided at the end of the path
  if [ ${#mockName} == 0 ]; then
    withoutLastSlash=${mockDir%"/"}
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
        echo -e "// $mockName -";
        echo "type $mockName struct {";
        } >> "$mockPath"
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

createMockStructure() {
  # navigate through all methods lines and create mock members with Called suffix and write them to the dest file
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
    echo "$structMember" >> "$mockPath"
  done
  # now mock struct is complete, close it
  echo -e "}\n" >> "$mockPath"
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

extractBasicParametersAndTypes() {
  # extract parameters from method line into:
  #   paramNames, which will be an array of strings used to call mock method
  #   paramTypes, which will be an array of strings exactly how the params types are. Eg. bool, error, uint32, etc.
  IFS=','
  read -ra ADDR <<< "$1"

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

extractParametersAndTypes() {
  parameters=$1

  # if there is no func into params, extract/append basic parameters and return
  if [[ ! "$parameters" == *"func"* ]]; then
    parametersUntilFirstBrace=${parameters%")"*}
    extractBasicParametersAndTypes "$parametersUntilFirstBrace"
    return
  fi

  # get the text before func
  untilFunc=${parameters%%"func("*}
  extractBasicParametersAndTypes "$untilFunc"

  # if last param type is not empty(ie it extracted some name for the func ptr), move last type to last name
  lastParamType=${paramTypes[-1]}
  if [ ! "$lastParamType" == "" ]; then
    paramNames[-1]=$lastParamType
  fi

  # get params of func ptr parameter
  fromFunc=${parameters#*"func("}
  paramsOfFunc=${fromFunc%%")"*}

  # extract return type of func
  fromParamsOfFunc=${fromFunc#*")"}
  returnTypeOfFunc=${fromParamsOfFunc%%","*}

  # if return type has any (, it means it has multiple return params
  if [[ "$returnTypeOfFunc" == *"("* ]]; then
    returnTypeOfFunc="${fromParamsOfFunc%%")"*})"
  fi

  # add the final param type, which is the entire func ptr signature
  paramTypes[-1]="func($paramsOfFunc)$returnTypeOfFunc"

  # call this method recursively in order to extract all possible func ptrs
  afterFunc=${fromParamsOfFunc#*"$returnTypeOfFunc, "}
  if [ "$returnTypeOfFunc" == "$afterFunc" ]; then
    # it means that the func ptr is the last parameter, so nothing left after func
    afterFunc=""
  else
    afterFunc=${afterFunc%")"*}
  fi

  extractParametersAndTypes "$afterFunc"
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
  { echo "mock.$mockField($stringParamNames)";
    echo "}";
    echo -e "}\n";
    } >> "$mockPath"
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
  { echo "return mock.$mockField($stringParamNames)";
  echo "}";
  } >> "$mockPath"

  # compute default values to return when mock member is not provided, separated by comma
  toReturn=""
  extractDefaultReturn

  # write the final return statement to file with default params and close the method
  { echo "return $toReturn";
    echo -e "}\n";
    } >> "$mockPath"
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
    { echo "return mock == nil";
      echo -e "}\n";
      } >> "$mockPath"
    return
  fi

  # add the check to mock member to not be nil
  echo "if mock.$mockField != nil {" >> "$mockPath"

  stringParamNames=""
  getStringParamNames

  # add return statement calling mock member
  # if there is no return type, add it without return
  # otherwise, return it with the provided params
  if [[ ${#returnTypesArr} == 0 ]]; then
    writeWithNoReturn
  else
    writeWithReturn
  fi
}

createMockMethods() {
  # navigate through all methods lines and:
  #   extract method name
  #   extract return types, used to handle the return
  #   extract parameters, used to call the mock member
  for method in "${methodsArr[@]}"
    do
      methodName=${method%%"("*}
      methodName=${methodName//[[:blank:]]/}
      # ignore commented lines
      [[ "$methodName" == "//"* ]] && continue

      removeCommentsFromMethodLine

      # extract from first ( until first )
      # this can lead to 2 cases:
      #   1. no func ptr param, which means rawParams will hold all parameters
      #   2. one or more func ptr params, which means rawParams will hold
      #   all parameters until the closing brace of the func ptr parameters
      rawParams=${method%")"*}
      rawParams=${rawParams#*"("}

      declare -a paramNames=()
      declare -a paramTypes=()

      extractParametersAndTypes "$rawParams"

      # extract final sequence of the methods parameters in order to properly extract its return types
      endingOfRawParam="()"
      if [ ${#paramNames} != 0 ]; then
        # remove previously appended ... for variable parameters
        lastParamName=${paramNames[-1]%%"."*}

        # if it's a placeholder, ignore param name. Otherwise, add it for better precision
        if [[ "$lastParamName" == "tmp"* ]]; then
          endingOfRawParam="${paramTypes[-1]})"
        else
          endingOfRawParam="$lastParamName ${paramTypes[-1]})"
        fi
      fi

      rawReturnTypesWithBraces=${method#*"$endingOfRawParam"}

      # clean fromParamsOfFunc
      fromParamsOfFunc=""

      # remove () in case of multiple return types
      rawReturnTypes=${rawReturnTypesWithBraces#*"("}
      rawReturnTypes=${rawReturnTypes%")"*}

      declare -a returnTypesArr=()
      extractReturnTypes

      # compute the mock member which will be called and write to the file:
      #   the comment
      #   the method signature
      # But first we compute the updated parameters, to avoid situation when param name is missing
      updatedParameters=""
      computeUpdatedParameters

      mockField=$methodName"Called"
      { echo "// $methodName -";
        echo "func (mock *$mockName) $methodName $updatedParameters $rawReturnTypesWithBraces {";
        } >> "$mockPath"

      createMethodBody
  done
}

generateMock() {
  interfaceName=$1
  filePath=$2"/interface.go"
  mockDir=$3

  [ ! -d "$2" ] && echo "Source directory for interface DOES NOT exists." && exit
  [ ! -f "$filePath" ] && echo "Source interface.go file DOES NOT exists." && exit
  [ ! -d "$mockDir" ] && echo "Destination directory DOES NOT exists." && exit

  extractPackageName

  mockName=$interfaceName"Mock"

  # make first char of the file name lowercase
  firstChar=${mockName::1}
  firstChar=${firstChar,,}

  lenOfMockName=${#mockName}
  mockFileName=$firstChar${mockName:1:$lenOfMockName}

  mockPath="$mockDir/$mockFileName.go"
  rm -rf "$mockPath"

  isInterfaceMethod=false
  declare -a methodsArr

  readInterfaceFile
  createMockStructure
  createMockMethods

  # go fmt file
  go fmt "$mockPath"
}

if [ $# -eq 3 ]; then
  generateMock "$@"
else
  echo "Please use the following format..."
  echo "bash mockGenerator.sh interface_name path_to_interface.go_dir path_to_mock_destionation_dir"
fi
