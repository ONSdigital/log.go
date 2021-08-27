#!/bin/bash
# 
# This is intended to help with upgrading apps with the correct versions of
# their mod deps.  
#
# When ran in an app's directory it should display deps containing old (pre-v2)
# logging code and open the release page for the deps github repos in the
# desktop browser to help with finding the new version of the dep to use in the
# repo.
#
# linux may need: alias open='xdg-open &> /dev/null'

go mod vendor

for i in $(grep -r log.go/log vendor/github.com/ONSdigital/ | grep -v v2 | awk -F "/" '{print "https://" $2"/"$3"/"$4"/releases"}'|sort -u); do
    echo $i
    open $i &> /dev/null
done

rm -rf vendor
