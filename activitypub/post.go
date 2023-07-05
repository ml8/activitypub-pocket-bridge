package activitypub

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/google/uuid"
)

func (p *activitypub) Note(user *User, post *Post) *Activity {
	a := &Activity{
		ID:           p.userBaseUrl(user.Name) + "&post=" + uuid.NewString(),
		Type:         "Note",
		To:           []string{ToAll},
		Published:    time.Now().Format(http.TimeFormat),
		AttributedTo: p.userBaseUrl(user.Name),
		Content:      post.Content(),
	}
	return a
}

func (p *activitypub) PeriodicPoster() {
	for {
		p.Lock()
		for _, user := range p.Users {
			go p.postArticle(user)
		}
		p.Unlock()
		glog.Infof("Sleeping for %v...", p.Interval)
		time.Sleep(p.Interval)
	}
}

func (p *activitypub) postArticle(u *User) {
	if !p.Pocket.IsLoggedIn(u.Name) {
		glog.Warningf("User %v not logged in", u.Name)
		return
	}
	u.Lock()
	if len(u.Followers) == 0 {
		glog.Infof("User %v has no followers", u.Name)
		u.Unlock()
		return
	}
	u.Unlock()
	art, err := p.Pocket.RandArticleForUser(u.Name)
	if err != nil {
		glog.Errorf("Error retrieving article for %v: %v", u.Name, err)
		return
	}
	glog.Infof("Posting %v", art)
	note := ActivityContext{Activity: *p.Note(u, &Post{&art}), Context: DefaultContext()}
	activity := ActivityContext{
		Activity: Activity{
			ID:     note.ID,
			Type:   "Create",
			Actor:  p.userBaseUrl(u.Name),
			To:     []string{ToAll},
			Object: note,
		},
		Context: SecurityContext(),
	}
	body, err := json.Marshal(activity)
	if err != nil {
		glog.Errorf("Error marshaling %v: %v", activity, err)
		return
	}
	glog.V(1).Infof("%v", string(body))
	postFunc := func(follower string) error {
		destActorInbox := follower + "/inbox"
		req, err := http.NewRequest("POST", destActorInbox, bytes.NewBuffer(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/activity+json")
		req.Header.Set("Host", req.URL.Host)
		req.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
		p.sign(u, req, body)
		glog.V(1).Infof("Sending %+v", req)
		client := &http.Client{}

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		body, _ := ioutil.ReadAll(resp.Body)
		glog.V(1).Infof("Got headers %+v", resp)
		glog.V(1).Infof("Got body %v", string(body))
		resp.Body.Close()
		return nil
	}
	err = u.forEachFollower(postFunc)
	if err != nil {
		glog.Errorf("Error posting article for %v: %v", u.Name, err)
	}
}
