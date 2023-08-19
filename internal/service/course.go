package service

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type CourseService interface {
	Invite(email string, rate int)
}

type courseService struct {
}

func NewCourseService() CourseService {
	return &courseService{}
}
func (c *courseService) Invite(email string, rate int) {
	apiURL := "https://skillspace.ru/api/open/v1/course/student-invite"

	// Parameters
	courseID := "14074"
	var groupID string
	switch rate {
	case 1:
		groupID = "22749"
	case 2:
		groupID = "24378"
	case 3:
		groupID = "24379"
	}
	token := "a70c4e05-26f2-3b73-8235-39833dd49747"
	name := "Student Name"
	comment := "Your Comment"

	// Create, send request and process response
	req, err := createRequest(apiURL, token, email, name, comment, courseID, groupID)
	if err != nil {
		panic(err)
	}

	resp, err := sendRequest(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	response := processResponse(resp)

	// Output the response
	fmt.Println("Response:", response)
}

func createRequest(apiURL, token, email, name, comment, courseID, groupID string) (*http.Request, error) {
	// Build query parameters
	params := url.Values{}
	params.Add("token", token)
	params.Add("email", email)
	params.Add("name", name)
	params.Add("comment", comment)
	params.Add(fmt.Sprintf("courses[%s]", courseID), groupID)

	// Create the request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBufferString(params.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return req, nil
}

func sendRequest(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	return client.Do(req)
}

func processResponse(resp *http.Response) string {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return string(body)
}
