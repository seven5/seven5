
package hellorest

import "seven5"

func saveFavoriteTeam(obj interface{}) error {
	err := seven5.SaveEntry("FavoriteTeam",obj)
	return err
}

func init()  {
	var fname = "/Volumes/External/seven5-dev/seven5/samples/hellorest/src/hellorest/favorite_team.json.json"
	err := seven5.LoadVocab(fname,&FavoriteTeam{},saveFavoriteTeam)  
	if err!=nil {
		panic("error loading vocab data "+fname+" at applicatin startup!")
	}
	return nil
}
