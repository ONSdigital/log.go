#!/usr/bin/env bash

path=
directory=false

print_usage() {
  echo ""
  echo "Update Logger Script"
  echo ""
  echo ""
  echo "Script is to help replace go-ns logging with log.go logging. This script will update the package in use and attempt to refactor current logging into the new structure."
  echo ""
  echo " * Updates logging library in file or directory"
  echo " * Attempts to update as many variations of old logs to new logs"
  echo " * Script is unable to update logs which traverse multiple lines"
  echo ""
  echo "Options are:"
  echo ""
  echo "      OPTION    DESCRIPTION                                 EXAMPLE ('' does NOT indicate default value)"
  echo "        -p      The path to file or directory from \$HOME    \"go/src/github.com/myNewService\""
  echo "        -d      Boolean flag, set if path is to directory   No value is needed, it defaults to false if not set"
  echo "        -h      Help                                        No value is needed, returns description of flags"
  echo ""
  echo "Examples:"
  echo ""
  echo " * ./edit-logs.sh -d -p /go-projects/search-api/handlers"
  echo " * ./edit-logs.sh -p /go-projects/search-api/handlers/handlers.go"
  echo ""
}

while getopts 'hdp:' flag; do
  case "${flag}" in
    p) path="${OPTARG}"
    ;;
    d) directory='true'
    ;;
    h) print_usage
       exit 0
    ;;
    \?) echo "ERROR: Unknown option $OPTARG" >&2
        echo "use -h flag for help"
        exit 1
    ;;
    *) print_usage >&2
       exit 1
       ;;
  esac
done

home=$HOME

if [[ -z $path ]]; then
  echo "ERROR: var not set [-p path]" >&2
  exit 1
fi

# path=${path:?ERROR: var not set [-p path]}

update_file(){
    echo "update ${1}"

    echo "Update log library"
    echo "=============================================="
    perl -i -p -e 's!github.com/ONSdigital/go-ns/log!github.com/ONSdigital/log.go/log!g' $1
    echo "done"


    echo "Capture and replace logs with err set"
    echo "=============================================="
    perl -i -p -e 's/log.(?:ErrorCtx|ErrorC|Error)\((?:ctx, )?("[^"]+"||[A-Za-z)-9]+)?(?:, )?(?:err)(?:, nil)*(?:, )*(, logData|, log.Data{"[^"]+": .+)*}?\)/log.Event(ctx, ${1}, log.Error(err)${2})/g' $1
    echo "done"


    echo "Capture and replace error logs with log.Data containing error"
    echo "=============================================="
    perl -i -p -e 's/log.(?:ErrorCtx|InfoCtx|DebugCtx|Error|Debug|Info)\((?:ctx, )*("[^"]+")*(?:nil)*(?:, )*(log.Data{("[^"]+": .+?, )?("error": err)(?:, )?("[^"]+": .+)*})\)/log.Event(ctx, ${1}, log.Error(err), log.Data{${3}${5}})/g' $1
    echo "done"


    echo "Capture and replace logs with log.Data or logData or not at all if both missing (handles nil)"
    echo "=============================================="
    perl -i -p -e 's/log.(?:ErrorCtx|InfoCtx|DebugCtx|Error|Debug|Info)\((?:ctx, )?("[^"]+")*(?:nil)*(?:, )*(?:nil)?(logData|log.Data\{"[^"]+": .+)*}?\)/log.Event(ctx, ${1}, ${2})/g' $1
    echo "finished updating file: $1"
}

if [[ -n $path ]]; then
    fullpath=$home$path
      
    if [[ $directory == true ]]; then
      echo "path to directory: ${fullpath}"
      for filename in "$fullpath"/*.go; do
        update_file $filename
      done
    else
      update_file $fullpath
    fi

    echo "done\n"
fi
