package model

import (
	"time"
)

type (
	UserPublicInfo struct {
		Email string `json:"_id" bson:"_id"`
		Role  Role   `json:"role" bson:"role"`
	}

	Query struct {
		Tags         []string     `json:"tags,omitempty" bson:"tags,omitempty"`
		Notification Notification `json:"notification" bson:"notification"`
		LastUpdated  *time.Time   `json:"last_updated,omitempty" bson:"last_updated,omitempty"`
	}

	User struct {
		UserPublicInfo `bson:"inline"`
		Name           string `json:"name,omitempty" bson:"name,omitempty"`
		Password       string `json:"password,omitempty" bson:"password"`
		Query          Query  `json:"query" bson:"query"`
		// a member to store last read info
	}

	Role int

	Notification int
)

const (
	RoleBasic Role = iota
	RoleAdmin

	NotificationNull Notification = iota
	NotificationDaily
	// NotificationInstant does not exist because data is scrapped daily
)

// TODO: tidy up
// func (u *User) MarshalJSON() ([]byte, error) {
// 	pubInfoJSON, err := json.Marshal(u.UserPublicInfo)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// creata a struct excluding a particular field
// 	intermediateJSON, err := json.Marshal(u)
// 	if err != nil {
// 		return nil, err
// 	}

// 	toMap := map[string]interface{}{}
// 	json.Unmarshal(intermediateJSON, &toMap)

// 	for _, field := range []string{"UserPublicInfo"} {
// 		delete(toMap, field)
// 	}

// 	pubInfoMap := map[string]interface{}{}
// 	json.Unmarshal(pubInfoJSON, &toMap)
// 	for k, v := range pubInfoMap {
// 		toMap[k] = v
// 	}

// 	outJSON, err := json.Marshal(toMap)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return outJSON, nil
// }
