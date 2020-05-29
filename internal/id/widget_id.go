package id

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type WidgetID string
type WidgetIndex int

func WidgetIDFromPageURLAndRoomIDAndWidgetIndex(pageURL PageURL, roomID RoomID, widgetIndex WidgetIndex) (WidgetID, error) {
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
	if widgetIndex < 0 {
		return "", errors.Errorf("Malformed widget index \"%d\": Widget index is negative", widgetIndex)
	}
	return WidgetID(string(pageURL) + "/" + string(roomID) + "/" + strconv.Itoa(int(widgetIndex))), nil
}

func PageURLAndRoomIDAndWidgetIndexFromWidgetID(widgetID WidgetID) (PageURL, RoomID, WidgetIndex, error) {
	if len(widgetID) == 0 {
		return "", "", 0, errors.New("Malformed widget ID: Page ID is empty")
	}
	r, err := regexp.Compile("(?i)^(\\/.+)*\\/([0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})\\/(\\d+)$")
	if err != nil {
		return "", "", 0, errors.Wrapf(err, "Malformed widget ID \"%s\": RegExp error", widgetID)
	}
	matches := r.FindStringSubmatch(string(widgetID))
	if len(matches) != 4 {
		return "", "", 0, errors.Errorf("Malformed widget ID \"%s\": Not a widget ID", widgetID)
	}
	if matches[1] == "" {
		matches[1] = "/"
	}
	widgetIndex, err := strconv.Atoi(matches[3])
	if err != nil {
		return "", "", 0, errors.Errorf("Malformed widget ID \"%s\": Widget index not numeric", widgetID)
	}
	return PageURL(matches[1]), RoomID(matches[2]), WidgetIndex(widgetIndex), nil
}
