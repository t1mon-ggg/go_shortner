package storage

import (
	"github.com/t1mon-ggg/go_shortner/app/models"
)

//Data - application storage interface
type Storage interface {
	Write(models.ClientData) error                  //write to storage
	ReadByCookie(string) (models.ClientData, error) //read from storage by cookie
	ReadByTag(string) (models.ShortData, error)     //read from storage by tag
	TagByURL(string, string) (string, error)        //get tag from storage by url
	Close() error                                   //close storage pointer
	Ping() error                                    //get storage status
	Cleaner(<-chan models.DelWorker, int)           //mark tag as deleted
}
