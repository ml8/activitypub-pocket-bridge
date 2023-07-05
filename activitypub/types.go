package activitypub

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"sync"

	"github.com/golang/glog"
	"github.com/ml8/ap-bot/pocket"
	"golang.org/x/exp/slices"
)

const (
	ToAll = "https://www.w3.org/ns/activitystreams#Public"
)

// ActivityStreams/Pub spec types

type WebFingerLink struct {
	Rel      string `json:"rel,omitempty"`
	Type     string `json:"type,omitempty"`
	Href     string `json:"href,omitempty"`
	Template string `json:"template,omitempty"`
}

type WebFingerNode struct {
	Subject string          `json:"subject,omitempty"`
	Aliases []string        `json:"aliases,omitempty"`
	Links   []WebFingerLink `json:"links,omitempty"`
}

type Actor struct {
	ID            string    `json:"id,omitempty"`
	Type          string    `json:"type,omitempty"`
	Inbox         string    `json:"inbox,omitempty"`
	Outbox        string    `json:"outbox,omitempty"`
	Approves      bool      `json:"manuallyApprovesFollwers,omitempty"`
	PreferredName string    `json:"preferredUsername,omitempty"`
	Summary       string    `json:"summary,omitempty"`
	Discoverable  bool      `json:"discoverable,omitempty"`
	Name          string    `json:"name,omitempty"`
	Followers     string    `json:"followers,omitempty"`
	PublicKey     PublicKey `json:"publicKey,omitempty"`
}

type Activity struct {
	ID           string      `json:"id,omitempty"`
	Actor        string      `json:"actor,omitempty"`
	Type         string      `json:"type,omitempty"`
	Object       interface{} `json:"object,omitempty"`
	To           []string    `json:"to,omitempty"`
	Cc           []string    `json:"cc,omitempty"`
	Content      string      `json:"content,omitempty"`
	Published    string      `json:"published,omitempty"`
	AttributedTo string      `json:"attributedTo,omitempty"`
}

type Collection struct {
	ID         string        `json:"id,omitempty"`
	Type       string        `json:"type,omitempty"`
	TotalItems int           `json:"totalItems,omitempty"`
	Items      []interface{} `json:"items,omitempty"`
}

// Internal types

type PublicKey struct {
	KeyId string `json:"id"`
	Owner string `json:"owner"`
	Key   string `json:"publicKeyPem"`
}

type User struct {
	sync.Mutex
	Name       string          `json:"name,omitempty"`
	Followers  []string        `json:"followers,omitempty"` // followers: should be user@service.social
	PrivateKey *rsa.PrivateKey `json:"privatekey,omitempty"`
}

func (u *User) encodePublicKey() (s string) {
	blk := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&u.PrivateKey.PublicKey),
	}
	return string(pem.EncodeToMemory(blk))
}

func (u *User) addFollower(follower string) {
	u.Lock()
	defer u.Unlock()
	if idx := slices.Index(u.Followers, follower); idx < 0 {
		u.Followers = append(u.Followers, follower)
		glog.Infof("Added follower of %v: %v", u.Name, follower)
	} else {
		glog.Warningf("Follower of %v already exists: %v", u.Name, follower)
	}
}

func (u *User) delFollower(follower string) {
	u.Lock()
	defer u.Unlock()
	if idx := slices.Index(u.Followers, follower); idx >= 0 {
		u.Followers = slices.Delete(u.Followers, idx, idx+1)
		glog.Infof("Removed follower of %v: %v", u.Name, follower)
		glog.Infof("Followers: %v", u.Followers)
	} else {
		glog.Warningf("Unknown follower of %v: %v", u.Name, follower)
	}
}

func (u *User) forEachFollower(f func(f string) error) error {
	u.Lock()
	followers := slices.Clone(u.Followers)
	u.Unlock()
	var err error
	for _, follower := range followers {
		t := f(follower)
		if t != nil {
			glog.Errorf("Error for follower %v: %v", follower, t)
			err = t
		}
	}
	return err
}

func NewCollection(id string, ordered bool) *Collection {
	typ := "Collection"
	if ordered {
		typ = "OrderedCollection"
	}
	return &Collection{
		ID:         id,
		Type:       typ,
		TotalItems: 0,
	}
}

func (c *Collection) AddItem(item interface{}) {
	c.TotalItems += 1
	c.Items = append(c.Items, item)
}

type Context struct {
	Context interface{} `json:"@context,omitempty"`
}

type ActorContext struct {
	Context
	Actor
}

type ActivityContext struct {
	Context
	Activity
}

type CollectionContext struct {
	Context
	Collection
}

type Post struct {
	*pocket.Article
}

func (p *Post) Content() string {
	return fmt.Sprintf("<b><a href=\"%v\">%v</a></b><br/>%v", p.Url, p.Title, p.Excerpt)
}

func DefaultContext() Context {
	return Context{Context: []string{"https://www.w3.org/ns/activitystreams"}}
}

func SecurityContext() Context {
	return Context{Context: []string{"https://www.w3.org/ns/activitystreams", "https://w3id.org/security/v1"}}
}
