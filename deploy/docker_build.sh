#!/bin/bash

printf "Building %s:%s and pusing to %s" ${AP_NAME} ${AP_VERSION_TAG} ${AP_PROJECT}

docker build -t ${AP_NAME}:${AP_VERSION_TAG} -f Dockerfile .
docker tag ${AP_NAME}:${AP_VERSION_TAG} gcr.io/${AP_PROJECT}/${AP_NAME}:${AP_VERSION_TAG}
docker push gcr.io/${AP_PROJECT}/${AP_NAME}:${AP_VERSION_TAG}
