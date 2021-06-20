package model

type Token struct {
	Uid string `json:"uid"`
	Roles string `json:"roles"`
	Access Rules `json:"access"`
	Deny Rules `json:"deny"`
	Info Userinfo `json:"info"`
}

type Rules struct {
	Read string `json:"read"`
	Write string `json:"write"`
	Delete string `json:"delete"`
	Admin string `json:"admin"`
}

type Userinfo struct {
	Name string `json:"name"`
	ClientType string `json:"client_type"`
}