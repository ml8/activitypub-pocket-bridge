package activitypub

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/ml8/ap-bot/util"
)

func (p *activitypub) getOrAddUser(name string) (user *User, exists bool) {
	p.Lock()
	defer p.Unlock()
	user, exists = p.Users[name]
	if !exists {
		key, _ := rsa.GenerateKey(rand.Reader, 2048)
		user = &User{
			Name:       name,
			Followers:  nil,
			PrivateKey: key,
		}
		p.Users[name] = user
		p.Persist()
	}
	return
}

func (p *activitypub) publicKeyForUser(u *User) *PublicKey {
	return &PublicKey{
		KeyId: p.userBaseUrl(u.Name) + "#main-key",
		Owner: p.userBaseUrl(u.Name),
		Key:   u.encodePublicKey(),
	}
}

func (p *activitypub) actorForUser(user *User) (a *Actor) {
	name := user.Name
	a = &Actor{
		ID:            p.userBaseUrl(name),
		Type:          "Person",
		Inbox:         p.userFeatureUrl("inbox", name),
		Outbox:        p.userFeatureUrl("outbox", name),
		Approves:      false,
		PreferredName: name,
		Summary:       fmt.Sprintf("Pocket user %v", name),
		Discoverable:  true,
		Name:          name,
		Followers:     p.userFeatureUrl("followers", name),
		PublicKey:     *p.publicKeyForUser(user),
	}
	return
}

func (p *activitypub) WebFingerHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("resource")
	glog.Infof("Got query %v", query)
	name, resType, _ := parseResourceString(query)
	if name == "" {
		util.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("No name found in query %v", query))
		return
	}
	if resType != "acct" {
		util.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Unknown resource type %v", resType))
		return
	}

	// Is user valid?
	if !p.Pocket.IsLoggedIn(name) {
		util.ErrorResponse(w, http.StatusNotFound, fmt.Sprintf("User %v not logged in", name))
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
		util.ErrorResponse(w, http.StatusNotFound, fmt.Sprintf("User %v is not logged in", name))
		return
	}
	user, _ := p.getOrAddUser(name)
	actor := p.actorForUser(user)
	wrapped := ActorContext{Actor: *actor, Context: SecurityContext()}
	glog.Infof("ActorResponse: %v", wrapped)
	util.JsonResponse(w, http.StatusOK, wrapped)
}

func (p *activitypub) CollectionHandler(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["account"]
	collection := mux.Vars(r)["collection"]
	if !p.Pocket.IsLoggedIn(name) {
		util.ErrorResponse(w, http.StatusNotFound, fmt.Sprintf("User %v is not logged in", name))
		return
	}
	user, _ := p.getOrAddUser(name)
	handler, ok := p.Handlers[collection]
	if !ok {
		util.ErrorResponse(w, http.StatusNotFound, fmt.Sprintf("Collection %v not found for user %v", collection, user))
		return
	}
	handler(user, w, r)
}

func (p *activitypub) FollowersHandler(user *User, w http.ResponseWriter, r *http.Request) {
	followers := NewCollection(p.userFeatureUrl("followers", user.Name), false)
	f := func(f string) error {
		followers.AddItem(f)
		return nil
	}
	user.forEachFollower(f)
	resp := CollectionContext{Collection: *followers, Context: DefaultContext()}
	util.JsonResponse(w, http.StatusOK, resp)
}

func (p *activitypub) UnfollowActivityHandler(user *User, activity *Activity, w http.ResponseWriter, r *http.Request) {
	user.delFollower(string(activity.Actor))
	p.Lock()
	p.Persist()
	p.Unlock()
	util.JsonResponse(w, http.StatusOK, "")
}

func (p *activitypub) FollowActivityHandler(user *User, activity *Activity, w http.ResponseWriter, r *http.Request) {
	user.addFollower(string(activity.Actor))
	p.Lock()
	p.Persist()
	p.Unlock()
	// new follower -- send a post
	go p.postArticle(user)
	util.JsonResponse(w, http.StatusOK, "")

	// Send a follow response to the follower's inbox.
	accept := ActivityContext{Activity: Activity{
		ID:     p.userBaseUrl(user.Name) + "&id=" + uuid.NewString(),
		Actor:  p.userBaseUrl(user.Name),
		Type:   "Accept",
		Object: ActivityContext{Activity: *activity, Context: DefaultContext()},
		To:     []string{activity.Actor},
	},
		Context: SecurityContext(),
	}
	destActorInbox := activity.Actor + "/inbox"
	jsonData, err := json.Marshal(accept)
	if err != nil {
		glog.Errorf("Error marshalling %v: %v", accept, err)
		return
	}
	glog.V(1).Infof("%v", string(jsonData))
	req, err := http.NewRequest("POST", destActorInbox, bytes.NewBuffer(jsonData))
	if err != nil {
		glog.Errorf("Error creating request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/activity+json")
	req.Header.Set("Host", req.URL.Host)
	req.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	p.sign(user, req, jsonData)
	glog.Infof("Sending %+v", req)
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		glog.Errorf("Error sending %v: %v", req, err)
		return
	}
	body, _ := ioutil.ReadAll(resp.Body)
	glog.Infof("Got response %+v - %v", resp, string(body))
	resp.Body.Close()
}

func (p *activitypub) InboxHandler(user *User, w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		util.ErrorResponse(w, http.StatusServiceUnavailable, "GET not supported for inbox")
		return
	}
	activity := &Activity{}
	body, _ := ioutil.ReadAll(r.Body)
	if err := json.Unmarshal(body, activity); err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Error unmarshalling %v: %v", string(body), err))
		return
	}
	glog.Infof("Parsed: %v", activity)
	if strings.ToLower(activity.Type) == "follow" {
		p.FollowActivityHandler(user, activity, w, r)
		return
	} else if strings.ToLower(activity.Type) == "undo" {
		p.UnfollowActivityHandler(user, activity, w, r)
		return
	}
	glog.V(1).Infof("Unsupported activity type %v: %v", activity.Type, activity)
	// Just say OK to things like undo...
	util.ErrorResponse(w, http.StatusOK, "")
}
