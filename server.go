package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v72"
	portalsession "github.com/stripe/stripe-go/v72/billingportal/session"
	"github.com/stripe/stripe-go/v72/checkout/session"
	"github.com/stripe/stripe-go/v72/webhook"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("godotenv.Load: %v", err)
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	stripe.SetAppInfo(&stripe.AppInfo{
		Name:    "stripe-samples/checkout-single-subscription",
		Version: "0.0.1",
		URL:     "https://github.com/stripe-samples/checkout-single-subscription",
	})

	http.Handle("/", http.FileServer(http.Dir(os.Getenv("STATIC_DIR"))))

	http.HandleFunc("/config", handleConfig)
	http.HandleFunc("/create-checkout-session", handleCreateCheckoutSession)
	http.HandleFunc("/checkout-session", handleCheckoutSession)
	http.HandleFunc("/customer-portal", handleCustomerPortal)
	http.HandleFunc("/webhook", handleWebhook)
	http.HandleFunc("/html/success.html", handleSuccessPage)

	addr := "0.0.0.0:4242"
	log.Printf("Listening on %s ...", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, struct {
		PublishableKey string `json:"publishableKey"`
		BasicPrice     string `json:"basicPrice"`
		ProPrice       string `json:"proPrice"`
	}{
		PublishableKey: os.Getenv("STRIPE_PUBLISHABLE_KEY"),
		BasicPrice:     os.Getenv("BASIC_PRICE_ID"),
		ProPrice:       os.Getenv("PRO_PRICE_ID"),
	}, nil)
}

func handleCreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm()
	priceId := r.PostFormValue("priceId")
	priceId = "price_1NrbSmJEPR9stpONKB1z5MnV"
	fmt.Println("priceId: ", priceId)
	params := &stripe.CheckoutSessionParams{
		SuccessURL: stripe.String(os.Getenv("DOMAIN") + "/html/success.html?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String(os.Getenv("DOMAIN") + "/canceled.html"),
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceId),
				Quantity: stripe.Int64(1),
			},
		},
	}

	s, err := session.New(params)
	if err != nil {
		writeJSON(w, nil, err)
		return
	}
	http.Redirect(w, r, s.URL, http.StatusSeeOther)
}

func handleCheckoutSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	sessionID := r.URL.Query().Get("sessionId")
	s, err := session.Get(sessionID, nil)
	writeJSON(w, s, err)
}

func handleCustomerPortal(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm()
	sessionID := r.PostFormValue("sessionId")[0:]

	s, err := session.Get(sessionID, nil)
	if err != nil {
		writeJSON(w, nil, err)
		return
	}

	returnURL := os.Getenv("DOMAIN")

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(s.Customer.ID),
		ReturnURL: stripe.String(returnURL),
	}
	ps, _ := portalsession.New(params)

	http.Redirect(w, r, ps.URL, http.StatusSeeOther)
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("ioutil.ReadAll: %v", err)
		return
	}

	event, err := webhook.ConstructEvent(b, r.Header.Get("Stripe-Signature"), os.Getenv("STRIPE_WEBHOOK_SECRET"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("webhook.ConstructEvent: %v", err)
		return
	}

	if event.Type != "checkout.session.completed" {
		return
	}

	sessionObj := event.Data.Object
	fmt.Println("sessionObj: ", sessionObj)
	customerName := sessionObj["customer_details"].(map[string]interface{})["name"].(string)
	amount := sessionObj["amount_total"].(float64) / 100
	successStatus := sessionObj["payment_status"].(string)

	fmt.Println("Customer Name: ", customerName)
	fmt.Println("Amount: $", amount)
	fmt.Println("Success Status: ", successStatus)

	w.WriteHeader(http.StatusOK)
}

func writeJSON(w http.ResponseWriter, v interface{}, err error) {
	var respVal interface{}
	if err != nil {
		msg := err.Error()
		var serr *stripe.Error
		if errors.As(err, &serr) {
			msg = serr.Msg
		}
		w.WriteHeader(http.StatusBadRequest)
		var e errResp
		e.Error.Message = msg
		respVal = e
	} else {
		respVal = v
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(respVal); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("json.NewEncoder.Encode: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := io.Copy(w, &buf); err != nil {
		log.Printf("io.Copy: %v", err)
		return
	}
}

func handleSuccessPage(w http.ResponseWriter, r *http.Request) {
	// Serve the success page
	http.ServeFile(w, r, "client/success.html")
}

type errResp struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}
