package models_test

import (
	"errors"
	"html-link-parser/models"
	"testing"
	"time"
)

var mockLinkAbout = models.Link{
	ID:              1,
	Href:            "/about-us",
	Text:            "About Us",
	SourceUrl:       "https://github.com/yanlinneo",
	BaseUrl:         "https://github.com/yanlinneo",
	CreatedDateTime: time.Now(),
}

var mockLinkContact = models.Link{
	ID:              2,
	Href:            "/contact-us",
	Text:            "Contact Us",
	SourceUrl:       "https://github.com/yanlinneo",
	BaseUrl:         "https://github.com/yanlinneo",
	CreatedDateTime: time.Now(),
}

var mockLinks = []models.Link{mockLinkAbout, mockLinkContact}

type MockLink struct {
	ID      int64
	Href    string
	Text    string
	BaseUrl string
}

func RelativePaths(baseUrl string) ([]models.Link, error) {
	links := []models.Link{}

	for _, ml := range mockLinks {
		if ml.BaseUrl == baseUrl {
			links = append(links, ml)
		}
	}

	return links, nil
}

func (link MockLink) UpdateStatus() (int64, error) {
	exist := false

	for _, ml := range mockLinks {
		if ml.ID == link.ID {
			exist = true
			break
		}
	}

	if !exist {
		return 0, nil
	}

	return 1, nil // rows affected
}

func (link MockLink) Add() (int64, error) {
	exist := false

	for _, ml := range mockLinks {
		if ml.Href == link.Href && ml.BaseUrl == link.BaseUrl {
			exist = true
			break
		}
	}

	if exist {
		return 0, errors.New("Link already exists.")
	}

	return 3, nil
}

func TestRelativePaths_BaseUrlNotInDatabase(t *testing.T) {
	baseUrl := "https://google.com.sg"
	links, err := RelativePaths(baseUrl)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(links) > 0 {
		t.Errorf("Want Length = 0 , got Length = %d", len(links))
	}
}

func TestRelativePaths_BaseUrlInDatabase(t *testing.T) {
	baseUrl := "https://github.com/yanlinneo"
	links, err := RelativePaths(baseUrl)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(links) != 2 {
		t.Errorf("want Length == %v , got Length = 0", len(links))
	}

	if links[0].Href != "/about-us" || links[1].Href != "/contact-us" {
		t.Errorf("want links[0].Href = /about-us, links[1].Href = /contact-us , got links[0].Href = %s, links[1].Href =  %s", links[0].Href, links[1].Href)
	}
}

func TestUpdateStatus(t *testing.T) {
	mockLinkTest := MockLink{ID: 2}
	rowsAffected, err := mockLinkTest.UpdateStatus()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if rowsAffected != 1 {
		t.Errorf("Expected Rows Affected == 1, got Rows Affected == %d", rowsAffected)
	}
}

func TestUpdateStatus_IDNotInDatabase(t *testing.T) {
	mockLinkTest := MockLink{ID: 12}
	rowsAffected, err := mockLinkTest.UpdateStatus()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if rowsAffected != 0 {
		t.Errorf("Expected Rows Affected == 0, got Rows Affected == %d", rowsAffected)
	}
}

func TestAdd(t *testing.T) {
	mockLinkHome := MockLink{Href: "/home", Text: "Home", BaseUrl: "https://github.com/yanlinneo"}
	id, err := mockLinkHome.Add()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if id != 3 {
		t.Errorf("Expected ID == 3, got ID == %d", id)
	}
}

func TestAdd_DuplicateEntry(t *testing.T) {
	mockLinkHome := MockLink{Href: "/about-us", Text: "About", BaseUrl: "https://github.com/yanlinneo"}
	id, err := mockLinkHome.Add()

	if err == nil {
		t.Fatalf("Expected duplicate entry error, got no error (err = nil)")
	}

	if id > 0 {
		t.Fatalf("Expected id = 0, got %d", id)
	}
}
