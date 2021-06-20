package model

import (
	"time"
)

// Configuration - текущая конфигурация для каждого запроса исходя из переданного TokenConfig
type PayIn struct {
	RedirectPostcreate string `json:"redirect_postcreate"`
	Product string `json:"product"`
	UserUID string `json:"user_uid"`
	UserName string `json:"user_name"`
	Configuration Custom `json:"configuration"`
}

type PayOut struct {
	RedirectUrl string `json:"redirect_url"`
	Code int `json:"code"`
	Body []byte `json:"body"`
}

type ConfirmationIn struct {
	Configuration Custom `json:"configuration"`
}

type ConfirmationOut struct {
}

type Payment struct {
	Amount map[string]string `json:"amount"`
	Confirmation map[string]string `json:"confirmation"`
	Description string `json:"description"`
}

type AnswerGateRound struct {
	ID string `json:"id"`
	Status string `json:"status"`
	Paid bool `json:"paid"`
	CreatedAt   time.Time `json:"created_at"`
	Amount map[string]string `json:"amount"`
	Confirmation map[string]string
}

type AnswerConfirmation struct {
	Type   string `json:"type"`
	Event  string `json:"event"`
	Object struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Paid   bool   `json:"paid"`
		Amount struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"amount"`
		AuthorizationDetails struct {
			Rrn      string `json:"rrn"`
			AuthCode string `json:"auth_code"`
		} `json:"authorization_details"`
		CreatedAt   time.Time `json:"created_at"`
		Description string    `json:"description"`
		ExpiresAt   time.Time `json:"expires_at"`
		Metadata    struct {
		} `json:"metadata"`
		PaymentMethod struct {
			Type  string `json:"type"`
			ID    string `json:"id"`
			Saved bool   `json:"saved"`
			Card  struct {
				First6      string `json:"first6"`
				Last4       string `json:"last4"`
				ExpiryMonth string `json:"expiry_month"`
				ExpiryYear  string `json:"expiry_year"`
				CardType    string `json:"card_type"`
			} `json:"card"`
			Title string `json:"title"`
		} `json:"payment_method"`
		Test bool `json:"test"`
	} `json:"object"`
}
