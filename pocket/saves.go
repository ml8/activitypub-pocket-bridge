package pocket

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"reflect"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/ml8/ap-bot/util"
)

const (
	GetUrl             = "/v3/get"
	ArticleUrlTemplate = "/article/{account}"
	Limit              = 50
)

type Article struct {
	Title   string `json:"title"`
	Excerpt string `json:"excerpt"`
	Url     string `json:"url"`
}

type GetRequest struct {
	ConsumerKey string `json:"consumer_key"`
	AccessToken string `json:"access_token"`
	Count       int    `json:"count"`
	DetailType  string `json:"detailType"`
	Sort        string `json:"sort"`
}

type GetItem struct {
	ItemId        string `json:"item_id"`
	ResolvedUrl   string `json:"resolved_url"`
	ResolvedTitle string `json:"resolved_title"`
	Excerpt       string `json:"excerpt"`
}

type GetResponse struct {
	Status int                `json:"status"`
	Items  map[string]GetItem `json:"list"`
}

func (p *pocket) getQuery(u *Userdata) (q []byte, err error) {
	// Return a json query
	req := GetRequest{
		ConsumerKey: p.AppKey,
		AccessToken: u.AccessToken,
		Count:       Limit,
		DetailType:  "simple",
		Sort:        "newest",
	}
	q, err = json.Marshal(req)
	return
}

func (p *pocket) ArticleHandler(w http.ResponseWriter, r *http.Request) {
	user := mux.Vars(r)["account"]
	a, err := p.RandArticleForUser(user)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Error retrieving article for user %v: %v", user, err))
		return
	}
	util.JsonResponse(w, http.StatusOK, a)
}

func (p *pocket) RandArticleForUser(user string) (a Article, err error) {
	p.Lock()
	u, ok := p.Tokens[user]
	p.Unlock()
	if !ok {
		err = errors.New(fmt.Sprintf("No user %v", user))
		return
	}
	if u.AccessToken == "" {
		err = errors.New(fmt.Sprintf("User %v not authenticated", user))
		return
	}

	// Get user articles
	var q []byte
	q, err = p.getQuery(&u)
	if err != nil {
		glog.Errorf("Error marshalling json for %v: %v", u, err)
		return
	}
	req, err := http.NewRequest("POST", PocketUrl+GetUrl, bytes.NewBuffer(q))
	if err != nil {
		glog.Errorf("Error creating query for %v: %v", user, err)
		return
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		glog.Errorf("Error posting login request: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	get := &GetResponse{}
	if err = json.Unmarshal(body, get); err != nil {
		glog.Errorf("Error unmarshalling %v: %v", string(body), err)
		return
	}

	// Choose random article
	i := rand.Intn(len(get.Items))
	keys := reflect.ValueOf(get.Items).MapKeys()
	art := get.Items[keys[i].String()]
	glog.Infof("Got %v articles; chose %v: %v", len(get.Items), i, art)
	a = Article{
		Title:   art.ResolvedTitle,
		Excerpt: art.Excerpt,
		Url:     art.ResolvedUrl,
	}
	return
}
