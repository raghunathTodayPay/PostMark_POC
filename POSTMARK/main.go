package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type PostmarkTemplate struct {
	Name     string `json:"Name"`
	Subject  string `json:"Subject"`
	HtmlBody string `json:"HtmlBody"`
	TextBody string `json:"TextBody"`
	Active   bool   `json:"active,omitempty"`
}

type PostmarkResponse struct {
	ErrorCode  int    `json:"ErrorCode"`
	Message    string `json:"Message"`
	TemplateID int64  `json:"TemplateID"`
}

type PostmarkTemplateListResponse struct {
	Templates []PostmarkTemplateDetails `json:"Templates"`
}

type PostmarkTemplateDetails struct {
	TemplateID int64  `json:"TemplateId"`
	Name       string `json:"Name"`
	Subject    string `json:"Subject"`
	Active     bool   `json:"Active"`
}

type EmailRequest struct {
	From       string `json:"From"`
	To         string `json:"To"`
	Subject    string `json:"Subject"`
	HtmlBody   string `json:"HtmlBody"`
	TextBody   string `json:"TextBody"`
	TemplateID int    `json:"TemplateID"`
}

type EmailResponse struct {
	To      string `json:"To"`
	Message string `json:"Message"`
}

type Client struct {
	baseURL    string
	apiToken   string
	httpClient *http.Client
}

func NewClient(apiToken string) *Client {
	return &Client{
		baseURL:    "https://api.postmarkapp.com",
		apiToken:   apiToken,
		httpClient: &http.Client{},
	}
}

func (c *Client) doRequest(method, url string, body interface{}, result interface{}) error {
	fullURL := c.baseURL + url
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, fullURL, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Postmark-Server-Token", c.apiToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		return fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, bodyString)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

func (c *Client) CreateTemplate(template PostmarkTemplate) (int64, error) {
	var postmarkResponse PostmarkResponse
	if err := c.doRequest("POST", "/templates", template, &postmarkResponse); err != nil {
		return 0, err
	}

	if postmarkResponse.ErrorCode != 0 {
		return 0, fmt.Errorf("failed to create template: %s", postmarkResponse.Message)
	}

	return postmarkResponse.TemplateID, nil
}

func (c *Client) UpdateTemplate(ID uint64, template PostmarkTemplate) error {
	url := fmt.Sprintf("/templates/%d", ID)
	var postmarkResponse PostmarkResponse
	if err := c.doRequest("PUT", url, template, &postmarkResponse); err != nil {
		return err
	}

	if postmarkResponse.ErrorCode != 0 {
		return fmt.Errorf("failed to update template: %s", postmarkResponse.Message)
	}

	return nil
}

func (c *Client) DeleteTemplate(ID uint64) error {
	url := fmt.Sprintf("/templates/%d", ID)
	var postmarkResponse PostmarkResponse
	if err := c.doRequest("DELETE", url, nil, &postmarkResponse); err != nil {
		return err
	}

	if postmarkResponse.ErrorCode != 0 {
		return fmt.Errorf("failed to delete template: %s", postmarkResponse.Message)
	}

	return nil
}

func (c *Client) GetTemplates(offset, count int) ([]PostmarkTemplateDetails, error) {
	url := fmt.Sprintf("/templates?offset=%d&count=%d", offset, count)
	var postmarkResponse PostmarkTemplateListResponse
	if err := c.doRequest("GET", url, nil, &postmarkResponse); err != nil {
		return nil, err
	}
	return postmarkResponse.Templates, nil
}

func (c *Client) SendEmail(email EmailRequest) (EmailResponse, error) {
	var emailResponse EmailResponse
	if err := c.doRequest("POST", "/email", email, &emailResponse); err != nil {
		return EmailResponse{}, err
	}
	return emailResponse, nil
}

func main() {
	client := NewClient("SERVER TOKEN")

	offset := 0
	count := 20

	// Get list of templates
	templates, err := client.GetTemplates(offset, count)
	if err != nil {
		log.Fatalf("Error getting templates: %v", err)
	}

	// Print template details
	for _, template := range templates {
		fmt.Printf("Template ID: %d, Name: %s, Subject: %s, Active: %t\n",
			template.TemplateID, template.Name, template.Subject, template.Active)
	}

	// Create a template
	// template := PostmarkTemplate{
	// 	Name:     "Test Template",
	// 	Subject:  "Hello, {{name}}!",
	// 	HtmlBody: "<html><body>Hello, {{name}}!</body></html>",
	// 	TextBody: "Hello, {{name}}!",
	// 	Active:   true,
	// }
	// templateID, err := client.CreateTemplate(template)
	// if err != nil {
	// 	log.Fatalf("Error creating template: %v", err)
	// }
	// fmt.Printf("Created template with ID: %d\n", templateID)

	// // Update a template
	// updatedTemplate := PostmarkTemplate{
	// 	Name:     "Updated Template",
	// 	Subject:  "Updated Hello, {{name}}!",
	// 	HtmlBody: "<html><body>Updated Hello, {{name}}!</body></html>",
	// 	TextBody: "Updated Hello, {{name}}!",
	// 	Active:   true,
	// }
	// err = client.UpdateTemplate(uint64(templateID), updatedTemplate)
	// if err != nil {
	// 	log.Fatalf("Error updating template: %v", err)
	// }
	// fmt.Printf("Updated template with ID: %d\n", templateID)

	// Send an email
	// email := EmailRequest{
	// 	From:     "mote.sai@todaypay.me",
	// 	To:       "raghunath.tiwari+2@todaypay.me",
	// 	TemplateID: 36274083,
	// 	TextBody: "Hello!",
	// }
	// emailResponse, err := client.SendEmail(email)
	// if err != nil {
	// 	log.Fatalf("Error sending email: %v", err)
	// }
	// fmt.Printf("Sent email to: %s\n", emailResponse.To)

	// Delete a template
	// err = client.DeleteTemplate(uint64(templateID))
	// if err != nil {
	// 	log.Fatalf("Error deleting template: %v", err)
	// }
	// fmt.Printf("Deleted template with ID: %d\n", templateID)
}
