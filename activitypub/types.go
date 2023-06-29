package activitypub

import "fmt"

type WebFingerLink struct {
	Rel      string `json:"rel"`
	Type     string `json:"type"`
	Href     string `json:"href"`
	Template string `json:"template"`
}

type WebFingerNode struct {
	Subject string          `json:"subject"`
	Aliases []string        `json:"aliases"`
	Links   []WebFingerLink `json:"links"`
}

type Actor struct {
	ID            string `json:"id"`
	Type          string `json:"type"`
	Inbox         string `json:"inbox"`
	Outbox        string `json:"outbox"`
	Approves      bool   `json:"manuallyApprovesFollwers"`
	PreferredName string `json:"preferredUsername"`
	Summary       string `json:"summary"`
	Discoverable  bool   `json:"discoverable"`
	Name          string `json:"name"`
}

type Context struct {
	Context []string `json:"@context"`
}

type ActorResponse struct {
	Actor
	Context
}

func (p *activitypub) actorForUser(name string) (a *Actor) {
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
	}
	return
}

func DefaultContext() Context {
	return Context{Context: []string{"https://www.w3.org/ns/activitystreams", "https://w3id.org/security/v1"}}
}

func WrapActor(a *Actor) ActorResponse {
	return ActorResponse{Actor: *a, Context: DefaultContext()}
}
