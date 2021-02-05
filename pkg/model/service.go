package model

import (
	"github.com/buildboxapp/service-yandex-money/pkg/jwt"
	"time"
)

type PayIn struct {
	RedirectPostcreate string `json:"redirect_postcreate"`
	Product string `json:"product"`
	Token jwt.Token
}

type PayOut struct {
	RedirectUrl string `json:"redirect_url"`
	Code int `json:"code"`
	Body []byte `json:"body"`
}

type ConfirmationIn struct {
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
