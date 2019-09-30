package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
)

type UserInfo struct {
	User User `json:"user"`
}

type ContactMethods struct {
	ID          string      `json:"id"`
	Type        string      `json:"type"`
	Summary     string      `json:"summary"`
	Self        string      `json:"self"`
	HTMLURL     interface{} `json:"html_url"`
	Label       string      `json:"label"`
	Address     string      `json:"address"`
	Blacklisted bool        `json:"blacklisted"`
	CountryCode int         `json:"country_code"`
	Enabled     bool        `json:"enabled,omitempty"`
}
type User struct {
	Name           string           `json:"name"`
	Email          string           `json:"email"`
	Role           string           `json:"role"`
	Description    string           `json:"description"`
	ContactMethods []ContactMethods `json:"contact_methods"`
}

var (
	authtoken = os.Getenv("TOKEN")
	userid    = os.Getenv("USERID")
	phoneid   = os.Getenv("PHONEID")
	smsid     = os.Getenv("SMSID")
)

func getPhoneNumberList() (nextPhone string) {
	client := &http.Client{}
	url := fmt.Sprintf(`https://api.pagerduty.com/users/%v?include%%5B%%5D=contact_methods`, userid)
	fmt.Println("URL:", url)
	reqest, err := http.NewRequest("GET", url, nil)

	//add header
	reqest.Header.Set("Content-Type", "application/json")
	reqest.Header.Set("Accept", "application/vnd.pagerduty+json;version=2")
	reqest.Header.Set("Authorization", "Token token="+authtoken)

	response, err := client.Do(reqest) //commit
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("handle err")
	}
	fmt.Println(string(body))
	// to struct
	var u UserInfo
	json.Unmarshal(body, &u)

	// phoneList
	phoneList := strings.Split(u.User.Description, ",")

	// get current phone
	var currentPhone string
	for k, v := range u.User.ContactMethods {
		if v.Type == "phone_contact_method" {
			currentPhone = u.User.ContactMethods[k].Address
		}
	}

	// get next phone
	for i := 0; i < len(phoneList); i++ {
		fmt.Println(i, phoneList[i], currentPhone)
		if currentPhone == phoneList[i] {
			if i == len(phoneList)-1 {
				i = 0
				nextPhone = phoneList[i]
				break
			}
			nextPhone = phoneList[i+1]
			fmt.Println("currentPhone--->", currentPhone)
			fmt.Println("nextPhone--->", nextPhone)
			break
		}

	}
	return
}

func putRequest(contact, contactid, nextPhone string) {
	url := fmt.Sprintf(`https://api.pagerduty.com/users/%v/contact_methods/%v`, userid, contactid)

	client := &http.Client{}
	var contactMethodgo = ContactMethods{
		ID:          contactid,
		Type:        contact,
		Summary:     "Mobile",
		Self:        url,
		HTMLURL:     "null",
		Label:       "Mobile",
		Address:     nextPhone,
		Blacklisted: false,
		CountryCode: 86,
	}

	payloadtxt, err := json.Marshal(contactMethodgo)
	if err != nil {
		log.Print(err)
	}

	payload := strings.NewReader(string(payloadtxt))

	req, err := http.NewRequest("PUT", url, payload)
	if err != nil {
		log.Print(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.pagerduty+json;version=2")
	req.Header.Set("Authorization", "Token token="+authtoken)

	resp, err := client.Do(req)
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		log.Print(err)
	}

	log.Print(string(body))
}

func updateUserPhone() {
	numTypeMap := make(map[string]string, 2)
	numTypeMap["phone_contact_method"] = phoneid
	numTypeMap["sms_contact_method"] = smsid

	nextPhone := getPhoneNumberList()

	for contact, contactid := range numTypeMap {
		fmt.Println(contact, contactid, nextPhone)
		putRequest(contact, contactid, nextPhone)
	}
}

func main() {
	lambda.Start(updateUserPhone)
}
