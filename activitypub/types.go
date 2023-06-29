package activitypub

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
	ID       string `json:"id"`
	Type     string `json:"type"`
	Inbox    string `json:"inbox"`
	Outbox   string `json:"outbox"`
	Approves bool   `json:"manuallyApprovesFollwers"`
}

type Context struct {
	Context []string `json:"@context"`
}

type ActorResponse struct {
	Actor
	Context
}

func DefaultContext() Context {
	return Context{Context: []string{"https://www.w3.org/ns/activitystreams", "https://w3id.org/security/v1"}}
}

func WrapActor(a *Actor) ActorResponse {
	return ActorResponse{Actor: *a, Context: DefaultContext()}
}
