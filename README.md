Hacking/learning ActivityPub and Pocket APIs... 

_Warning: this code is a horrible mess just hacked together as a
proof-of-concept..._

## Goal

> Implement a focused subset of the activitypub spec that allows a mastodon user to
> follow a pocket username as @username@pocket-activitypub-service.social and for
> the pocket-activitypub-service.social to federate pocket saved articles to the
> user's mastodon instance.

## Building/Running

Requires:

  * A GCP project with a data disk (optional) and an ip address (required) provisioned
  * the following environment variables set:
    
    ```
    AP_NAME        # name of the app (for gcr.io)
    AP_PROJECT     # name of GCP project
    AP_VM          # name of the VM
    AP_ZONE        # zone of the VM
    AP_IP          # IP in project
    AP_DATA_DISK   # disk in project (optional: don't supply -db flag)
    AP_POCKET_KEY  # pocket API key
    AP_VERSION_TAG # docker version tag (e.g., latest)
    ```
  * A domain (supplied with `-domain`) pointing to `AP_IP`

The scripts `deploy/docker_build.sh` builds and pushes the container to
`gcr.io`, while `deploy/deploy.sh` and `deploy/update.sh` create and update the
VM, respectively.

To link a pocket account, visit `MY_DOMAIN/pocket/register`. Then, once the
account is linked, you can follow your `username@MY_DOMAIN` from any mastodon
(or other activitypub federated service).

* Link to live instance:
  [hq.jerry.business](https://hq.jerry.business/pocket/register)

For example, the following two accounts are two pocket accounts I linked (with
not much saved in them...):

  * `langma@hq.jerry.business`
  * `m@hq.jerry.business`
