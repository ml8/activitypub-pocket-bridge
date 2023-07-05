#!/bin/bash

gcloud compute instances create-with-container ${AP_VM} \
  --zone ${AP_ZONE} \
  --container-image gcr.io/${AP_PROJECT}/${AP_NAME}:${AP_VERSION_TAG} \
  --machine-type e2-micro \
  --tags http-allow,http-server,https-server \
  --address ${AP_IP} \
  --disk name=${AP_DATA_DISK},mode=rw \
  --container-mount-disk mount-path="/disks/data-disk",name=${AP_DATA_DISK},mode=rw \
  --container-arg="-prod" \
  --container-arg="-certDir=/disks/data-disk/state" \
  --container-arg="-pocketAppKey=${AP_POCKET_KEY}" \
  --container-arg="-db=/disks/data-disk/state" \
  --container-arg="-postInterval=60m" \
  --container-arg="-logtostderr"
  
