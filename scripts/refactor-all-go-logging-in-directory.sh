#!/bin/bash

# Runs sed script refactor-go-logging.sh against all .go files found in TARGET_DIR given as argument $1
# NB: will edit all files in place!

realpath ()                                                                                                                                                                                   
{                                                                                                                                                                                             
    f=$@;                                                                                                                                                                                     
    if [ -d "$f" ]; then                                                                                                                                                                      
        base="";                                                                                                                                                                              
        dir="$f";                                                                                                                                                                             
    else                                                                                                                                                                                      
        base="/$(basename "$f")";                                                                                                                                                             
        dir=$(dirname "$f");                                                                                                                                                                  
    fi;                                                                                                                                                                                       
    dir=$(cd "$dir" && /bin/pwd);                                                                                                                                                             
    echo "$dir$base"                                                                                                                                                                          
}

THIS_DIR=$(realpath $(dirname $0))
TARGET_DIR=$1
find ${TARGET_DIR} -type f -name "*.go" -print0 | while read -d $'\0' file
do
  ${THIS_DIR}/refactor-go-logging.sh < ${file} > ${file}.new && mv ${file}.new ${file}
done

