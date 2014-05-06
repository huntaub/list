package email

import (
	"errors"
	"fmt"
	"github.com/huntaub/list/app/models"
	"github.com/mattbaird/gochimp"
	"net/url"
	"time"
)

// Send Email Verification
func SendVerificationEmail(user *models.User, baseURL string) error {
	// Open New Mandrill API
	api, err := gochimp.NewMandrill(MandrillAPIKey)
	if err != nil {
		return err
	}

	// Create the message
	message := gochimp.Message{
		To: []gochimp.Recipient{
			// Only One Recipient!
			gochimp.Recipient{
				Email: user.Email,
				Name:  user.FullName,
			},
		},
		TrackOpens:       true,
		TrackClicks:      true,
		InlineCss:        true,
		AutoText:         true,
		TrackingDomain:   "track.list.hunterleath.com",
		SigningDomain:    "list.hunterleath.com",
		ReturnPathDomain: "track.list.hunterleath.com",
		Merge:            true,
		GlobalMergeVars: []gochimp.Var{
			gochimp.Var{
				Name:    "NAME",
				Content: user.FullName,
			},
			gochimp.Var{
				Name:    "LDATE",
				Content: time.Now().Format("Monday, January 2"),
			},
			gochimp.Var{
				Name:    "VERIFICATION_URL",
				Content: fmt.Sprintf("%v/verify/%v/%v", baseURL, user.VerificationKey, url.QueryEscape(user.Email)),
			},
		},
	}

	// Send the message
	response, err := api.MessageSendTemplate("leath-s-list-email-verification", nil, message, false)
	for _, v := range response {
		if v.RejectedReason != "" {
			return errors.New(v.RejectedReason)
		}
	}

	return err
}
