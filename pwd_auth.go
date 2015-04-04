package seven5

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	AUTH_OP_LOGIN         = "login"
	AUTH_OP_LOGOUT        = "logout"
	AUTH_OP_PWD_RESET     = "pwdreset"
	AUTH_OP_PWD_RESET_REQ = "pwdresetreq"
)

//PasswordAuthParameters is passed from client to server to request login, login
//or to use (consume) a reset request.  XXX Ugh, this has to be manually copied
//over to the client side library. XXX
type PasswordAuthParameters struct {
	Username         string
	Password         string
	ResetRequestUdid string
	UserUdid         string
	Op               string
}

//Valdating session manager is one that can also check the validity of a
//username and password.  This can be easily wrapped around a SimpleSessionManager.
//The ValidateCredentials method should return "" (first return) for a failed
//login attempt and use the error for something more serious, like the database
//cannot be reached.  If the first returned value from ValidateCredentials is
//not "" it should be a unique id, both of the first two returned values will be
//sent to the (nested) session manager's Assign.
type ValidatingSessionManager interface {
	SessionManager
	ValidateCredentials(username, password string) (string, interface{}, error)
	SendUserDetails(i interface{}, w http.ResponseWriter) error
	GenerateResetRequest(string) (string, error)
	UseResetRequest(string, string, string) (bool, error)
}

//SimplePasswordHandler is a utility for handling login-logout and authentication
//checks.  It expects to be given a SessionManager that it will work in combination
//with.
type SimplePasswordHandler struct {
	vsm ValidatingSessionManager
	cm  CookieMapper
}

//
// NewSimplePasswordHandler returns a password handler utility object that is
// associated with the SessionManager provided.  Normally the caller will want
// to bind /me" to MeHandler, /auth to AuthHandler.
//
func NewSimplePasswordHandler(vsm ValidatingSessionManager, cm CookieMapper) *SimplePasswordHandler {
	return &SimplePasswordHandler{
		vsm: vsm,
		cm:  cm,
	}
}

//
// Check verifies that the username and password provided are the ones we expect
// via a calle the ValidatingSessionManager. It returns nil,nil in the case of a
// failed check on the password provided.
//
func (self *SimplePasswordHandler) Check(username, pwd string) (Session, error) {
	uniq, userData, err := self.vsm.ValidateCredentials(username, pwd)
	if err != nil {
		return nil, err
	}
	if uniq == "" {
		return nil, nil
	}
	return self.vsm.Assign(uniq, userData, time.Time{})
}

//
// Me returns the currently logged in user to the client.
//
func (self *SimplePasswordHandler) MeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Cache-Control", "no-cache, must-revalidate") //HTTP 1.1
	w.Header().Add("Pragma", "no-cache")                         //HTTP 1.0

	val, err := self.cm.Value(r)
	if err != nil && err != NO_SUCH_COOKIE {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err != nil { //no cookie
		http.Error(w, "no cookie", http.StatusUnauthorized)
		return
	}
	sr, err := self.vsm.Find(strings.TrimSpace(val))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if sr == nil {
		http.Error(w, "no session", http.StatusUnauthorized)
		return
	}

	if sr.Session != nil {
		if err := self.vsm.SendUserDetails(sr.Session.UserData(), w); err != nil {
			log.Printf("failed to send user data: %v", err)
		}
		return
	}
	i, err := self.vsm.Generate(sr.UniqueId)
	if err != nil {
		http.Error(w, fmt.Sprintf("unable to recover session: %v", err), http.StatusInternalServerError)
		return
	}
	recovered, err := self.vsm.Assign(sr.UniqueId, i, time.Time{})
	if err := self.vsm.SendUserDetails(recovered.UserData(), w); err != nil {
		log.Printf("failed to send user data: %v", err)
	}
	return

}

func (self *SimplePasswordHandler) AuthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Cache-Control", "no-cache, must-revalidate") //HTTP 1.1
	w.Header().Add("Pragma", "no-cache")                         //HTTP 1.0

	//READ INPUT FROM CLIENT
	buf := make([]byte, 512)
	n, err := io.ReadFull(r.Body, buf)
	if err != nil && err != io.ErrUnexpectedEOF {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	r.Body.Close()

	//DECODE INPUT FROM CLIENT
	b := bytes.NewBuffer(buf[:n])
	dec := json.NewDecoder(b)
	var auth PasswordAuthParameters
	err = dec.Decode(&auth)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	val, err := self.cm.Value(r)
	if err != nil && err != NO_SUCH_COOKIE {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//
	//PW RESET REQ? (Can be done without being logged in)
	//
	if auth.Op == AUTH_OP_PWD_RESET_REQ {
		resetUdid, err := self.vsm.GenerateResetRequest(auth.Username)
		if err != nil {
			WriteError(w, err)
			log.Printf("[AUTH] error returned from GenerateResetRequest %v", err)
			return
		}
		log.Printf("[AUTH] generated password reset req for user %s: %s",
			auth.UserUdid, resetUdid)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}")) //need to prevent the client-side dying
		return
	}

	//
	//PW RESET? (Can be done without being logged in)
	//
	if auth.Op == AUTH_OP_PWD_RESET {
		//UseResetRequest(userId string, requestId string, newpwd string) (bool, error) {

		ok, err := self.vsm.UseResetRequest(auth.UserUdid, auth.ResetRequestUdid, auth.Password)
		if err != nil {
			WriteError(w, err)
			log.Printf("[AUTH] error returned from UseResetRequest %v", err)
			return
		}
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			log.Printf("[AUTH] UseResetRequest refused to update password: %s", auth.ResetRequestUdid)
			return
		}
		log.Printf("[AUTH] reset password for user %s with token %s",
			auth.UserUdid, auth.ResetRequestUdid)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}")) //need to prevent the client-side dying
		return
	}

	//
	//LOGOUT?
	//
	if auth.Op == AUTH_OP_LOGOUT {
		if err == NO_SUCH_COOKIE {
			http.Error(w, "not logged in", http.StatusBadRequest)
		} else {
			self.cm.RemoveCookie(w)
			self.vsm.Destroy(val)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{}")) //prevent client side dying
		}
		return
	}

	//
	// MUST BE LOGIN
	//
	session, err := self.Check(auth.Username, auth.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if session == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	log.Printf("[AUTH] user %s is authenticated", auth.Username)
	self.cm.AssociateCookie(w, session)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}")) //need to prevent the client-side dying
}
