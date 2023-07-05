package activitypub

import (
	"net/http"
	"regexp"

	sig "github.com/go-fed/httpsig"
	"github.com/golang/glog"
)

func (p *activitypub) sign(u *User, r *http.Request, body []byte) {
	prefs := []sig.Algorithm{sig.RSA_SHA256}
	headers := []string{sig.RequestTarget, "host", "date", "digest"}
	signer, _, err := sig.NewSigner(prefs, sig.DigestSha256, headers, sig.Signature, 60)
	if err != nil {
		glog.Errorf("Error creating signer: %v", err)
		return
	}
	if err := signer.SignRequest(u.PrivateKey, p.userBaseUrl(u.Name)+"#main-key", r, body); err != nil {
		glog.Errorf("Error appending signature: %v", err)
	}
	sigStr := r.Header.Get("Signature")
	re := regexp.MustCompile("algorithm=\"hs2019\"")
	sigStr = re.ReplaceAllString(sigStr, string("algorithm=\""+sig.RSA_SHA256+"\""))
	r.Header.Set("Signature", sigStr)
}
