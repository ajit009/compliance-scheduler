package confluence

import (
	"aws-compliance-scheduler/pkg/config"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	goconfluence "github.com/virtomize/confluence-go-api"
)

type Content struct {
	Id     string `json:"id"`
	Type   string `json:"type"`
	Status string `json:"status"`
	Title  string `json:"title"`
	Body   struct {
		Storage struct {
			Value          string `json:"value"`
			Representation string `json:"representation"`
		} `json:"storage"`
	} `json:"body"`
	Version struct {
		Number int `json:"number"`
	} `json:"version"`
}

func (w *Wiki) contentEndpoint(contentID string) (*url.URL, error) {
	return url.ParseRequestURI(w.endPoint.String() + "/content/" + contentID)
}

func (w *Wiki) DeleteContent(contentID string) error {
	contentEndPoint, err := w.contentEndpoint(contentID)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("DELETE", contentEndPoint.String(), nil)
	if err != nil {
		return err
	}

	_, err = w.sendRequest(req)
	if err != nil {
		return err
	}
	return nil
}

func (w *Wiki) GetContent(contentID string, expand []string) (*Content, error) {
	contentEndPoint, err := w.contentEndpoint(contentID)
	if err != nil {
		return nil, err
	}
	data := url.Values{}
	data.Set("expand", strings.Join(expand, ","))
	contentEndPoint.RawQuery = data.Encode()

	req, err := http.NewRequest("GET", contentEndPoint.String(), nil)
	if err != nil {
		return nil, err
	}

	res, err := w.sendRequest(req)
	if err != nil {
		return nil, err
	}

	var content Content
	err = json.Unmarshal(res, &content)
	if err != nil {
		return nil, err
	}

	return &content, nil
}

func (w *Wiki) UpdateContent(content *Content) (*Content, error) {
	jsonbody, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}

	contentEndPoint, err := w.contentEndpoint(content.Id)
	req, err := http.NewRequest("PUT", contentEndPoint.String(), strings.NewReader(string(jsonbody)))
	req.Header.Add("Content-Type", "application/json")

	res, err := w.sendRequest(req)
	if err != nil {
		return nil, err
	}

	var newContent Content
	err = json.Unmarshal(res, &newContent)
	if err != nil {
		return nil, err
	}

	return &newContent, nil
}

/*
Updates the confluence page for the reports.
This method would need the confluence user and pass in environment
*/
func UpdateConfluencePage(jj map[string]map[string]string, accountNumber string, resourceType string) bool {
	//fmt.Print(jj)
	fmt.Println("Using " + os.Getenv("CONFLUENCE_USER") + " to update confluence")
	api, err := goconfluence.NewAPI("https://ajlocal.net/wiki/rest/api", string(os.Getenv("CONFLUENCE_USER")), string(os.Getenv("CONFLUENCE_USER_KEY")))

	if err != nil {
		log.Fatal(err)
	}
	markupStringHead := `
	<h5>Following <strong>` + strconv.Itoa(len(jj)) + `</strong> resources have some of the mandatory tags as missing/non-compliant</h5>

	<table data-layout="full-width"><colgroup><col style="width: 287.0px;" /><col style="width: 286.0px;" /><col style="width: 286.0px;" /><col style="width: 505.0px;" /></colgroup>
	<tbody>
	<tr>
		<th><p><strong>Instance id</strong></p></th>
		<th><p><strong>Environment</strong></p></th>
		<th><p><strong>Team</strong></p></th>
		<th><p><strong>Comments</strong></p></th>
	</tr>
	`
	markupStringBody := ``
	for id, tagSet := range jj {
		markupStringBody = markupStringBody + `<tr><td><p>` + id + `</p></td><td><p>` + tagSet["environment"] + `</p></td><td><p>` + tagSet["team"] + `</p></td><td><p>` + tagSet["comments"] + `</p></td></tr>`
	}

	markupStringFoot := `<tr>
	<td>
	<p /></td>
	<td>
	<p /></td>
	<td>
	<p /></td>
	<td>
	<p /></td></tr></tbody></table>`
	resourceId, pageTitle, _ := config.MapAccountToConfluenceResource(accountNumber, resourceType)
	// fmt.Print("Printing markup")
	// fmt.Println(markupStringHead + markupStringBody + markupStringFoot)

	var latestVersionNumber int = 1

	version, err := api.GetContentVersion(resourceId)
	if err != nil {
		fmt.Println("object " + err.Error())
	} else {
		latestVersionNumber = version.Result[0].Number + 1
	}
	data := &goconfluence.Content{
		ID:    resourceId,
		Type:  "page",    // can also be blogpost
		Title: pageTitle, // page title
		Ancestors: []goconfluence.Ancestor{
			goconfluence.Ancestor{
				ID: string(config.GetRootIdForResource(resourceType)),
			},
		},
		Body: goconfluence.Body{
			Storage: goconfluence.Storage{
				Value:          markupStringHead + markupStringBody + markupStringFoot,
				Representation: "storage",
			},
		},
		Version: &goconfluence.Version{
			Number: latestVersionNumber,
		},
		Space: goconfluence.Space{
			Key: "YOUR_CONFLUENCE_SPACE_ID", // Space
		},
	}

	// fmt.Print(data)
	c, err := api.UpdateContent(data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Update completed with doc status %s", c.Status)
	return true
}
