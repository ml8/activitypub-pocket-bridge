// URL handlers for pocket authentication and pocket API.
package pocket

import (
	"net/http"

	"github.com/golang/glog"
)

const (
	PocketUrl              = "https://getpocket.com"
	PocketAuthRequestUrl   = "/v3/oauth/request"
	PocketAuthAuthorizeUrl = "/v3/oauth/authorize"
	RegisterUrlTemplate    = "/register/{account}"
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
	Username    string
	AccessToken string
	AuthCode    string
}

type pocket struct {
	Tokens map[string]*Userdata // map user->token
	AppKey string
	AppUrl string
}

func Init(url, key string, bootstrap *BootstrapData) Pocket {
	glog.Infof("Application at %v", url)
	p := &pocket{Tokens: make(map[string]*Userdata), AppUrl: url, AppKey: key}
	for u, c := range bootstrap.Users {
		p.Tokens[u] = &Userdata{
			Username:    u,
			AccessToken: c,
			AuthCode:    "",
		}
	}
	return p
}

func (p *pocket) IsLoggedIn(user string) bool {
	u, ok := p.Tokens[user]
	return ok && u.AccessToken != ""
}
