#!/bin/bash

unset HTTP_PROXY
/usr/bin/curl --silent http://localhost:5005/v2/_catalog | jq -r '.repositories[]' | 
    while read -r image; do
        /usr/bin/curl --silent http://localhost:5005/v2/${image}/tags/list
    done
