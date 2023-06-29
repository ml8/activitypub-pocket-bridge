Hacking/learning ActivityPub and Pocket APIs... 

## Goal

> Implement a focused subset of the activitypub spec that allows a mastodon user to
> follow a pocket username as @username@pocket-activitypub-service.social and for
> the pocket-activitypub-service.social to federate pocket saved articles to the
> user's mastodon instance.

## Status

__So far:__

* a golang library for authenticating and retrieving pocket saves 
* an implementation of webfinger
  ([required](https://docs.joinmastodon.org/spec/webfinger/) by mastodon
  integration) for the service
* the [actor](https://www.w3.org/TR/activitypub/#actor-objects) object
  [retrieve](https://www.w3.org/TR/activitypub/#retrieving-objects) API

__Next steps are to:__

* implement the [follow
  activity](https://www.w3.org/TR/activitypub/#follow-activity-inbox) (mastodon
  <-> pocket-activitypub-service.social)
  post saves to the follower
  ([delivery](https://www.w3.org/TR/activitypub/#delivery))
* figure out whatever else I need to do for mastodon to work ok with my
  cobbled-together code
