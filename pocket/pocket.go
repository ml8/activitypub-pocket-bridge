// URL handlers for pocket authentication and pocket API.
package pocket

import (
	"net/http"
	"sync"

	"github.com/golang/glog"
	"github.com/ml8/ap-bot/util"
)

const (
	PocketUrl              = "https://getpocket.com"
	PocketAuthRequestUrl   = "/v3/oauth/request"
	PocketAuthAuthorizeUrl = "/v3/oauth/authorize"
	RegisterUrlRoot        = "/register"
	RegisterUrlTemplate    = RegisterUrlRoot + "/{account}"
	CallbackUrlRoot        = "/callback"
	CallbackUrlTemplate    = CallbackUrlRoot + "/{account}"
)

type Pocket interface {
	RegisterHandler(w http.ResponseWriter, r *http.Request)
	RegisterCallback(w http.ResponseWriter, r *http.Request)
	ArticleHandler(w http.ResponseWriter, r *http.Request)
	RandArticleForUser(user string) (Article, error)
	IsLoggedIn(user string) bool
}

type BootstrapData struct {
	Users map[string]string
}

type Userdata struct {
	Username    string `json:"username,omitempty"`
	AccessToken string `json:"accesstoken,omitempty"`
	AuthCode    string `json:"authcode,omitempty"`
}

type ResourceMap struct {
	AppUrl string
	Host   string
}

type pocket struct {
	sync.Mutex
	Tokens         map[string]Userdata // protected by mutex
	AppKey         string
	Resources      ResourceMap
	StateInterface util.Persister
}

func Init(key string, resources ResourceMap, bootstrap *BootstrapData, statefile string) Pocket {
	glog.Infof("Application at %v", resources.AppUrl)
	p := &pocket{Tokens: make(map[string]Userdata), Resources: resources, AppKey: key}
	for u, c := range bootstrap.Users {
		p.Tokens[u] = Userdata{
			Username:    u,
			AccessToken: c,
			AuthCode:    "",
		}
	}
	p.StateInterface = util.NewPersister(statefile)
	p.Recover()
	if len(bootstrap.Users) != 0 {
		p.Persist()
	}
	return p
}

func (p *pocket) Recover() {
	// Requires mutex
	p.StateInterface.Read(&p.Tokens)
	glog.Infof("Recovered %+v", p.Tokens)
}

func (p *pocket) Persist() {
	// Requires mutex
	p.StateInterface.Write(p.Tokens)
}

func (p *pocket) IsLoggedIn(user string) bool {
	p.Lock()
	defer p.Unlock()
	u, ok := p.Tokens[user]
	return ok && u.AccessToken != ""
}
