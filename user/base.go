package user

import (
	"seven5"
)

type Base interface {
	seven5.Session
	FirstName() String
	LastName() String
	Name() String
	Email() String
	UserId() seven5.Id
}

type SimpleBase struct {
	fetch map[string]seven5.Fetcher
}
type BaseManager struct {
	wrapped *seven5.SimpleSessionManager
}

func (self *BaseManager) Find(id string) (Session,error) {
	return self.wrapped.Find(id)
}
func (self *BaseManager) Destroy(id string) error {
	return self.wrapped.Destroy(id)
}
func (self *BaseManager) Generate(id string, f Fetcher, r *http.Request, state string, code string) (Session,error) {
	return nil,nil
}