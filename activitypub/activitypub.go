package activitypub

import (
	"net/http"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/ml8/ap-bot/pocket"
	"github.com/ml8/ap-bot/util"
)

const (
	WebFingerUrl               = "/.well-known/webfinger"
	ActorUrlPrefix             = "/user"
	ActorUrlTemplate           = ActorUrlPrefix + "/{account}"      // /user/{account}
	ActorCollectionUrlTemplate = ActorUrlTemplate + "/{collection}" // /user/{account}/{collection}
)

type ActivityPub interface {
	WebFingerHandler(w http.ResponseWriter, r *http.Request)
	ActorHandler(w http.ResponseWriter, r *http.Request)
	CollectionHandler(w http.ResponseWriter, r *http.Request)
	Start()
}

type ResourceMap struct {
	BaseUrl string
	Host    string
}

type CollectionHandler func(u *User, w http.ResponseWriter, r *http.Request)

type activitypub struct {
	sync.Mutex
	Pocket         pocket.Pocket
	Resources      ResourceMap
	Users          map[string]*User // protected by mutex
	Handlers       map[string]CollectionHandler
	StateInterface util.Persister
	Interval       time.Duration
}

func Init(p pocket.Pocket, resources ResourceMap, statefile string, postInterval time.Duration) ActivityPub {
	pub := &activitypub{
		Pocket:         p,
		Resources:      resources,
		Users:          make(map[string]*User),
		Handlers:       make(map[string]CollectionHandler),
		StateInterface: util.NewPersister(statefile),
		Interval:       postInterval,
	}
	pub.Handlers["followers"] = pub.FollowersHandler
	pub.Handlers["inbox"] = pub.InboxHandler
	pub.Recover()
	return pub
}

func (p *activitypub) Start() {
	go p.PeriodicPoster()
}

func (p *activitypub) userBaseUrl(name string) string {
	return p.Resources.BaseUrl + ActorUrlPrefix + "/" + name
}

func (p *activitypub) userFeatureUrl(feature, name string) string {
	return p.userBaseUrl(name) + "/" + feature
}

func (p *activitypub) Recover() {
	// Requires mutex
	p.StateInterface.Read(&p.Users)
	glog.Infof("Recovered %+v", p.Users)
}

func (p *activitypub) Persist() {
	// Requires mutex
	p.StateInterface.Write(p.Users)
}
