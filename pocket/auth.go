package pocket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/ml8/ap-bot/util"
)

type PreAuthRequest struct {
	ConsumerKey string `json:"consumer_key"`
	RedirectUri string `json:"redirect_uri"`
}

type PreAuthResponse struct {
	Code string `json:"code"`
}

type AuthRequest struct {
	ConsumerKey string `json:"consumer_key"`
	Code        string `json:"code"`
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	Username    string `json:"username"`
}

const (
	successSrc = `
<html>
	<head></head>
	<body>
	  Now, you may follow %v@%v from your mastodon (etc) account.
	</body>
</html>
`
)

func (p *pocket) redirectUrl(token, uri string) string {
	u, _ := url.Parse(PocketUrl + "/auth/authorize")
	q := u.Query()
	q.Set("request_token", token)
	q.Set("redirect_uri", uri)
	u.RawQuery = q.Encode()
	return u.String()
}

func (p *pocket) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	acct := mux.Vars(r)["account"]
	glog.Infof("Registering %v", acct)
	if acct == "" {
		util.ErrorResponse(w, http.StatusPreconditionFailed, "No user found")
		return
	}

	// Create + send auth request
	authReq := PreAuthRequest{
		ConsumerKey: p.AppKey,
		RedirectUri: p.Resources.AppUrl + CallbackUrlRoot + "/" + acct,
	}
	jsonData, err := json.Marshal(authReq)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Error marshalling request: %v", err))
		return
	}
	req, err := http.NewRequest("POST", PocketUrl+PocketAuthRequestUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	authResp := &PreAuthResponse{}
	if err = json.Unmarshal(body, authResp); err != nil {
		msg := fmt.Sprintf("Error unmarshalling %v: %v", string(body), err)
		util.ErrorResponse(w, http.StatusInternalServerError, msg)
		return
	}
	glog.Infof("Got code %v for user %v", authResp.Code, acct)

	p.Lock()
	p.Tokens[acct] = Userdata{
		Username:    acct,
		AccessToken: "",
		AuthCode:    authResp.Code,
	}
	p.Unlock()

	// Redirect user to pocket auth
	redirect := p.redirectUrl(authResp.Code, authReq.RedirectUri)
	http.Redirect(w, r, redirect, 301)
}

func (p *pocket) RegisterCallback(w http.ResponseWriter, r *http.Request) {
	acct := mux.Vars(r)["account"]
	glog.Infof("Authorizing %v", acct)
	if acct == "" {
		util.ErrorResponse(w, http.StatusPreconditionFailed, "No user found")
		return
	}

	p.Lock()
	user, ok := p.Tokens[acct]
	p.Unlock()
	if !ok {
		util.ErrorResponse(w, http.StatusPreconditionFailed, fmt.Sprintf("Unknown user %v", acct))
		return
	}

	// User is authenticated; get the token for them.
	authReq := AuthRequest{
		ConsumerKey: p.AppKey,
		Code:        user.AuthCode,
	}

	jsonData, err := json.Marshal(authReq)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Error marshalling request: %v", err))
		return
	}
	req, err := http.NewRequest("POST", PocketUrl+PocketAuthAuthorizeUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Error sending %v: %v", req, err))
		return
	}
	defer resp.Body.Close()

	glog.Infof("Got %v", resp)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	authResp := &AuthResponse{}
	if err = json.Unmarshal(body, authResp); err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Error unmarshalling %v: %v", string(body), err))
		return
	}

	// sweet.
	p.Lock()
	user, _ = p.Tokens[acct]
	user.AccessToken = authResp.AccessToken
	p.Tokens[acct] = user
	p.Persist()
	p.Unlock()
	glog.Infof("Got token %v for user %v", authResp.AccessToken, acct)
	w.Write([]byte(fmt.Sprintf(successSrc, acct, p.Resources.Host)))
}
