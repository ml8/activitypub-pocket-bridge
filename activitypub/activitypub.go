package activitypub

import (
	"net/http"
	"strings"

	"github.com/golang/glog"
	"github.com/gorilla/mux"

	"github.com/ml8/ap-bot/pocket"
	"github.com/ml8/ap-bot/util"
)

const (
	WebFingerUrl               = "/.well-known/webfinger"
	ActorUrlPrefix             = "/user"
	ActorUrlTemplate           = ActorUrlPrefix + "/{account}"     // /user/{account}
	ActorCollectionUrlTemplate = ActorUrlTemplate + "/{collection" // /user/{account}/{collection}
)

type ActivityPub interface {
	WebFingerHandler(w http.ResponseWriter, r *http.Request)
	ActorHandler(w http.ResponseWriter, r *http.Request)
}

type ResourceMap struct {
	BaseUrl string
}

type activitypub struct {
	Pocket    pocket.Pocket
	Resources ResourceMap
}

func Init(p pocket.Pocket, resources ResourceMap) ActivityPub {
	pub := &activitypub{Pocket: p, Resources: resources}
	return pub
}

func (p *activitypub) userBaseUrl(name string) string {
	return p.Resources.BaseUrl + ActorUrlPrefix + "/" + name
}

func (p *activitypub) userFeatureUrl(feature, name string) string {
	return p.userBaseUrl(name) + "/" + feature
}

func parseResourceString(s string) (name string, resType string, url string) {
	var delim string
	if strings.Contains(s, "://") {
		delim = "://"
	} else if strings.Contains(s, ":") {
		delim = ":"
	} else {
		return
	}
	splt := strings.Split(s, delim)
	if len(splt) != 2 {
		return
	}
	resType = splt[0]
	name = splt[1]
	name, _ = strings.CutPrefix(name, "@")
	splt = strings.Split(name, "@")
	if len(splt) != 2 {
		return
	}
	name = splt[0]
	url = splt[1]
	return
}

func (p *activitypub) WebFingerHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("resource")
	glog.Infof("Got query %v", query)
	name, resType, _ := parseResourceString(query)
	if name == "" {
		glog.Errorf("No name found in query %v", query)
		util.ErrorResponse(w, http.StatusInternalServerError, "")
		return
	}
	if resType != "acct" {
		glog.Errorf("Unknown resource type %v", resType)
		util.ErrorResponse(w, http.StatusInternalServerError, "")
		return
	}

	// Is user valid?
	if !p.Pocket.IsLoggedIn(name) {
		glog.Errorf("User %v not logged in", name)
		util.ErrorResponse(w, http.StatusNotFound, "")
		return
	}

	// Provide response
	resp := &WebFingerNode{}
	resp.Subject = query
	resp.Aliases = nil
	// Using that WebFinger response, Mastodon will check the following:
	//   - The subject is present
	//   - The links array contains a link with rel of self and type of either:
	//        application/ld+json; profile="https://www.w3.org/ns/activitystreams", or
	//        application/activity+json
	//       - The href for this link resolves to an ActivityPub actor
	//
	// Using that ActivityPub actor representation (which may be provided directly, without
	// the initial WebFinger request), Mastodon will do the following:
	//   - Take preferredUsername and the hostname of the actor’s server
	//   - Construct an acct: URI using that username and domain
	//   - Make a Webfinger request for that resource
	resp.Links = append(resp.Links, WebFingerLink{
		Rel:  "self",
		Type: "application/activity+json",
		Href: p.Resources.BaseUrl + ActorUrlPrefix + "/" + name,
	})
	util.JsonResponse(w, http.StatusOK, resp)
	return
}

func (p *activitypub) ActorHandler(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["account"]
	glog.Infof("Actor for %v", name)
	if !p.Pocket.IsLoggedIn(name) {
		glog.Errorf("User %v is not logged in", name)
		util.ErrorResponse(w, http.StatusNotFound, "")
		return
	}
	actor := Actor{
		ID:       p.userBaseUrl(name),
		Type:     "Person",
		Inbox:    p.userFeatureUrl("inbox", name),
		Outbox:   p.userFeatureUrl("outbox", name),
		Approves: false,
	}
	wrapped := WrapActor(&actor)
	glog.Infof("ActorResponse: %v", wrapped)
	util.JsonResponse(w, http.StatusOK, wrapped)
}
