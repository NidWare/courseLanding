package service

import (
	"bytes"
	"courseLanding/internal/config"
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
	var groupID string
	switch rate {
	case 1:
		groupID = "32904"
	case 2:
		groupID = "32905"
	case 3:
		groupID = "32906"
	}

	req, err := createRequest(config.EduURL, config.Token, email, "Student Name", "Comment", "21298", groupID)
	if err != nil {
		panic(err)
	}

	resp, err := sendRequest(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	response := processResponse(resp)
	fmt.Println("Response for email:", email, response)
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
