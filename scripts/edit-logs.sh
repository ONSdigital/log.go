#!/usr/bin/env bash

path=
me=$(basename $0)

print_usage() {
    echo ""
    echo "Update Logger Script"
    echo ""
    echo ""
    echo "$me helps replace go-ns logging with log.go logging. This script will update the package in use and attempt to refactor current logging into the new structure."
    echo ""
    echo " * Updates logging library in given file or directory"
    echo " * Attempts to update as many variations of old logs to new logs"
    echo " * Unable to update logs which traverse multiple lines"
    echo ""
    echo "      OPTION              DESCRIPTION         EXAMPLE (NOT default value)"
    echo "        -p <file_or_dir>  File or directory   \"go/src/github.com/myNewService\""
    echo "        -h                Show this help"
    echo ""
    echo "Example usage:"
    echo ""
    echo " * Add the script to your PATH, with one of:"
    echo "   * PATH=\$GOPATH/src/github.com/ONSdigital/log.go/scripts:\$PATH"
    echo "   * cp -ip scripts/$me ~/bin/$me"
    echo " * Run against chosen file or directory:"
    echo "   * $me -p handlers/handlers.go"
    echo "   * $me -p handlers"
    echo ""
}

while getopts 'hp:' flag; do
    case "${flag}" in
        (p) path="${OPTARG}"
            ;;
        (h) print_usage
            exit 0
            ;;
       (\?) echo "ERROR: Unknown option $OPTARG" >&2
            echo "use -h flag for help"
            exit 1
            ;;
        (*) print_usage >&2
            exit 1
            ;;
    esac
done

if [[ -z $path ]]; then
  echo "ERROR: required flag not used [-p path]" >&2
  exit 1
fi

banner() {
    echo "$@"
    echo "===================================================="
}

update_file(){
    banner "update ${1}"

    banner "Update log library"
    perl -i -p -e 's!github.com/ONSdigital/go-ns/log!github.com/ONSdigital/log.go/log!g' $1
    echo "done"


    banner "Capture and replace logs with err set"
    perl -i -p -e 's/log.(?:ErrorCtx|ErrorC|Error)\((?:ctx, )?("[^"]+"||[A-Za-z)-9]+)?(?:, )?(?:err)(?:, nil)*(?:, )*(, logData|, log.Data{"[^"]+": .+)*}?\)/log.Event(ctx, ${1}, log.Error(err)${2})/g' $1
    echo "done"


    banner "Capture and replace error logs with log.Data containing error"
    perl -i -p -e 's/log.(?:ErrorCtx|InfoCtx|DebugCtx|Error|Debug|Info)\((?:ctx, )*("[^"]+")*(?:nil)*(?:, )*(log.Data{("[^"]+": .+?, )?("error": err)(?:, )?("[^"]+": .+)*})\)/log.Event(ctx, ${1}, log.Error(err), log.Data{${3}${5}})/g' $1
    echo "done"


    banner "Capture and replace logs with log.Data or logData or not at all if both missing (handles nil)"
    perl -i -p -e 's/log.(?:ErrorCtx|InfoCtx|DebugCtx|Error|Debug|Info)\((?:ctx, )?("[^"]+")*(?:nil)*(?:, )*(?:nil)?(logData|log.Data\{"[^"]+": .+)*}?\)/log.Event(ctx, ${1}, ${2})/g' $1
    echo "finished updating file: $1"
}

if [[ -d $path ]]; then
    banner "path to directory: ${path}"
    for filename in "$path"/*.go; do
        update_file $filename
    done
else
    update_file $path
fi

banner "done"
