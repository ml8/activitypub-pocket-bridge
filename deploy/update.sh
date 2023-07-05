#!/bin/bash

gcloud compute instances update-container ${AP_VM} \
  --zone ${AP_ZONE} \
  --container-image gcr.io/${AP_PROJECT}/${AP_NAME}:${AP_VERSION_TAG} \
  --container-arg="-prod" \
  --container-arg="-pocketAppKey=${AP_POCKET_KEY}" \
  --container-arg="-certDir=/disks/data-disk/state" \
  --container-arg="-db=/disks/data-disk/state" \
  --container-arg="-postInterval=60m" \
  --container-arg="-logtostderr"
  
