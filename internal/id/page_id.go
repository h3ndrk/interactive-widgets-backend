package id

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

type PageID string
type PageURL string
type RoomID string

func PageIDFromPageURLAndRoomID(pageURL PageURL, roomID RoomID) (PageID, error) {
	if len(pageURL) == 0 {
		return "", errors.New("Malformed page URL: Page URL is empty")
	}
	if pageURL[0:1] != "/" {
		return "", errors.Errorf("Malformed page URL \"%s\": First character not \"/\"", pageURL)
	}
	parts := strings.Split(string(pageURL), "/")
	for i, part := range parts {
		if i != 0 && i < len(parts)-1 && len(part) == 0 {
			return "", errors.Errorf("Malformed page URL \"%s\": Multiple slashes after each other", pageURL)
		}
	}
	if pageURL[len(pageURL)-1:] == "/" {
		pageURL = pageURL[:len(pageURL)-1]
	}
	match, err := regexp.MatchString("(?i)^[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12}$", string(roomID))
	if err != nil {
		return "", errors.Wrapf(err, "Malformed room ID \"%s\": RegExp error", roomID)
	}
	if !match {
		return "", errors.Errorf("Malformed room ID \"%s\": Not a UUIDv4", roomID)
	}
	return PageID(string(pageURL) + "/" + string(roomID)), nil
}

func PageURLAndRoomIDFromPageID(pageID PageID) (PageURL, RoomID, error) {
	if len(pageID) == 0 {
		return "", "", errors.New("Malformed page ID: Page ID is empty")
	}
	r, err := regexp.Compile("(?i)^(\\/.+)*\\/([0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})$")
	if err != nil {
		return "", "", errors.Wrapf(err, "Malformed page ID \"%s\": RegExp error", pageID)
	}
	matches := r.FindStringSubmatch(string(pageID))
	if len(matches) != 3 {
		return "", "", errors.Errorf("Malformed page ID \"%s\": Not a page ID", pageID)
	}
	if matches[1] == "" {
		matches[1] = "/"
	}
	return PageURL(matches[1]), RoomID(matches[2]), nil
}
