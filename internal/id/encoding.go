package id

import "encoding/hex"

type EncodedPageURL string
type EncodedPageID string
type EncodedWidgetID string

func EncodePageURL(pageURL PageURL) EncodedPageURL {
	return EncodedPageURL(hex.EncodeToString([]byte(pageURL)))
}

func DecodePageURL(encodedPageURL PageURL) (PageURL, error) {
	pageURL, err := hex.DecodeString(string(encodedPageURL))
	if err != nil {
		return "", err
	}
	return PageURL(pageURL), nil
}

func EncodePageID(pageID PageID) EncodedPageID {
	return EncodedPageID(hex.EncodeToString([]byte(pageID)))
}

func DecodePageID(encodedPageID PageID) (PageID, error) {
	pageID, err := hex.DecodeString(string(encodedPageID))
	if err != nil {
		return "", err
	}
	return PageID(pageID), nil
}

func EncodeWidgetID(widgetID WidgetID) EncodedWidgetID {
	return EncodedWidgetID(hex.EncodeToString([]byte(widgetID)))
}

func DecodeWidgetID(encodedWidgetID WidgetID) (WidgetID, error) {
	widgetID, err := hex.DecodeString(string(encodedWidgetID))
	if err != nil {
		return "", err
	}
	return WidgetID(widgetID), nil
}
