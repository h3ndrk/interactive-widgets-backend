package id

import "testing"

func TestPageIDFromPageURLAndRoomID(t *testing.T) {
	testCases := []struct {
		pageURL  PageURL
		roomID   RoomID
		pageID   PageID
		gotError bool
	}{
		// pageURL not empty, pageURL with /-prefix,    pageURL with /-suffix,    roomID is valid UUIDv4 (lowercase)
		{"/", "00000000-0000-4000-8000-aaaaaaaaaaaa", "/00000000-0000-4000-8000-aaaaaaaaaaaa", false},
		{"//", "00000000-0000-4000-8000-aaaaaaaaaaaa", "", true},
		{"/foo/", "00000000-0000-4000-8000-aaaaaaaaaaaa", "/foo/00000000-0000-4000-8000-aaaaaaaaaaaa", false},
		// pageURL not empty, pageURL with /-prefix,    pageURL with /-suffix,    roomID is valid UUIDv4 (uppercase)
		{"/", "00000000-0000-4000-8000-AAAAAAAAAAAA", "/00000000-0000-4000-8000-AAAAAAAAAAAA", false},
		{"//", "00000000-0000-4000-8000-AAAAAAAAAAAA", "", true},
		{"/foo/", "00000000-0000-4000-8000-AAAAAAAAAAAA", "/foo/00000000-0000-4000-8000-AAAAAAAAAAAA", false},
		// pageURL not empty, pageURL without /-prefix, pageURL with /-suffix,    roomID is valid UUIDv4 (lowercase)
		{"foo/", "00000000-0000-4000-8000-aaaaaaaaaaaa", "", true},
		// pageURL not empty, pageURL without /-prefix, pageURL with /-suffix,    roomID is valid UUIDv4 (uppercase)
		{"foo/", "00000000-0000-4000-8000-AAAAAAAAAAAA", "", true},
		// pageURL not empty, pageURL with /-prefix,    pageURL without /-suffix, roomID is valid UUIDv4 (lowercase)
		{"/foo", "00000000-0000-4000-8000-aaaaaaaaaaaa", "/foo/00000000-0000-4000-8000-aaaaaaaaaaaa", false},
		// pageURL not empty, pageURL with /-prefix,    pageURL without /-suffix, roomID is valid UUIDv4 (uppercase)
		{"/foo", "00000000-0000-4000-8000-AAAAAAAAAAAA", "/foo/00000000-0000-4000-8000-AAAAAAAAAAAA", false},
		// pageURL empty,     pageURL without /-prefix, pageURL without /-suffix, roomID is valid UUIDv4 (lowercase)
		{"", "00000000-0000-4000-8000-aaaaaaaaaaaa", "", true},
		// pageURL empty,     pageURL without /-prefix, pageURL without /-suffix, roomID is valid UUIDv4 (uppercase)
		{"", "00000000-0000-4000-8000-AAAAAAAAAAAA", "", true},
		// pageURL not empty, pageURL without /-prefix, pageURL without /-suffix, roomID is valid UUIDv4 (lowercase)
		{"foo", "00000000-0000-4000-8000-aaaaaaaaaaaa", "", true},
		// pageURL not empty, pageURL without /-prefix, pageURL without /-suffix, roomID is valid UUIDv4 (uppercase)
		{"foo", "00000000-0000-4000-8000-AAAAAAAAAAAA", "", true},
		// pageURL not empty, pageURL with /-prefix,    pageURL with /-suffix,    roomID is invalid UUIDv4
		{"/", "00000000-0000-4000X8000-aaaaaaaaaaaa", "", true},
		{"//", "00000000-0000-4000X8000-aaaaaaaaaaaa", "", true},
		{"/foo/", "00000000-0000-4000X8000-aaaaaaaaaaaa", "", true},
		// pageURL not empty, pageURL without /-prefix, pageURL with /-suffix,    roomID is invalid UUIDv4
		{"foo/", "00000000-0000-4000X8000-aaaaaaaaaaaa", "", true},
		// pageURL not empty, pageURL with /-prefix,    pageURL without /-suffix, roomID is invalid UUIDv4
		{"/foo", "00000000-0000-4000X8000-aaaaaaaaaaaa", "", true},
		// pageURL empty,     pageURL without /-prefix, pageURL without /-suffix, roomID is invalid UUIDv4
		{"", "00000000-0000-4000X8000-aaaaaaaaaaaa", "", true},
		// pageURL not empty, pageURL without /-prefix, pageURL without /-suffix, roomID is invalid UUIDv4
		{"foo", "00000000-0000-4000X8000-aaaaaaaaaaaa", "", true},
	}

	for i, testCase := range testCases {
		pageID, err := PageIDFromPageURLAndRoomID(testCase.pageURL, testCase.roomID)
		if pageID != testCase.pageID {
			t.Errorf("Test Case %d: Got page ID \"%s\", expected \"%s\"", i, pageID, testCase.pageID)
		}
		if testCase.gotError && err == nil {
			t.Errorf("Test Case %d: Got no error, expected error", i)
		} else if !testCase.gotError && err != nil {
			t.Errorf("Test Case %d: Got error (%v), expected no error", i, err)
		}
	}
}

func TestPageURLAndRoomIDFromPageID(t *testing.T) {
	testCases := []struct {
		pageID   PageID
		pageURL  PageURL
		roomID   RoomID
		gotError bool
	}{
		// empty, without /-prefix, without multiple /, without /-suffix, invalid UUIDv4
		{"", "", "", true},
		// not empty, without /-prefix, without multiple /, without /-suffix, invalid UUIDv4
		{"00000000-0000-4000X8000-aaaaaaaaaaaa", "", "", true},
		// not empty, with /-prefix, without multiple /, without /-suffix, invalid UUIDv4
		{"/00000000-0000-4000X8000-aaaaaaaaaaaa", "", "", true},
		// not empty, without /-prefix, with multiple /, without /-suffix, invalid UUIDv4
		{"00000000-/0000-4000X8000-aaaaaaaaaaaa", "", "", true},
		// not empty, with /-prefix, with multiple /, without /-suffix, invalid UUIDv4
		{"//00000000-0000-4000X8000-aaaaaaaaaaaa", "", "", true},
		{"/foo/00000000-0000-4000X8000-aaaaaaaaaaaa", "", "", true},
		// not empty, without /-prefix, without multiple /, with /-suffix, invalid UUIDv4
		{"00000000-0000-4000X8000-aaaaaaaaaaaa/", "", "", true},
		// not empty, with /-prefix, without multiple /, with /-suffix, invalid UUIDv4
		{"/00000000-0000-4000X8000-aaaaaaaaaaaa/", "", "", true},
		// not empty, without /-prefix, with multiple /, with /-suffix, invalid UUIDv4
		{"foo/foo/00000000-0000-4000X8000-aaaaaaaaaaaa/", "", "", true},
		// not empty, with /-prefix, with multiple /, with /-suffix, invalid UUIDv4
		{"//00000000-0000-4000X8000-aaaaaaaaaaaa/", "", "", true},
		{"/foo/00000000-0000-4000X8000-aaaaaaaaaaaa/", "", "", true},
		// not empty, without /-prefix, without multiple /, without /-suffix, valid UUIDv4 (lowercase)
		{"abcdef/00000000-0000-4000-8000-aaaaaaaaaaaa", "", "", true},
		// not empty, without /-prefix, without multiple /, without /-suffix, valid UUIDv4 (uppercase)
		{"abcdef/00000000-0000-4000-8000-AAAAAAAAAAAA", "", "", true},
		// not empty, with /-prefix, without multiple /, without /-suffix, valid UUIDv4 (lowercase)
		{"/00000000-0000-4000-8000-aaaaaaaaaaaa", "/", "00000000-0000-4000-8000-aaaaaaaaaaaa", false},
		// not empty, with /-prefix, without multiple /, without /-suffix, valid UUIDv4 (uppercase)
		{"/00000000-0000-4000-8000-AAAAAAAAAAAA", "/", "00000000-0000-4000-8000-AAAAAAAAAAAA", false},
		// not empty, with /-prefix, with multiple /, without /-suffix, valid UUIDv4 (lowercase)
		{"//00000000-0000-4000-8000-aaaaaaaaaaaa", "", "", true},
		{"/foo/00000000-0000-4000-8000-aaaaaaaaaaaa", "/foo", "00000000-0000-4000-8000-aaaaaaaaaaaa", false},
		// not empty, with /-prefix, with multiple /, without /-suffix, valid UUIDv4 (uppercase)
		{"//00000000-0000-4000-8000-AAAAAAAAAAAA", "", "", true},
		{"/foo/00000000-0000-4000-8000-AAAAAAAAAAAA", "/foo", "00000000-0000-4000-8000-AAAAAAAAAAAA", false},
		// not empty, without /-prefix, without multiple /, with /-suffix, valid UUIDv4 (lowercase)
		{"00000000-0000-4000-8000-aaaaaaaaaaaa/", "", "", true},
		// not empty, without /-prefix, without multiple /, with /-suffix, valid UUIDv4 (uppercase)
		{"00000000-0000-4000-8000-AAAAAAAAAAAA/", "", "", true},
		// not empty, with /-prefix, without multiple /, with /-suffix, valid UUIDv4 (lowercase)
		{"/00000000-0000-4000-8000-aaaaaaaaaaaa/", "", "", true},
		// not empty, with /-prefix, without multiple /, with /-suffix, valid UUIDv4 (uppercase)
		{"/00000000-0000-4000-8000-AAAAAAAAAAAA/", "", "", true},
		// not empty, without /-prefix, with multiple /, with /-suffix, valid UUIDv4 (lowercase)
		{"foo/bar/00000000-0000-4000-8000-aaaaaaaaaaaa", "", "", true},
		// not empty, without /-prefix, with multiple /, with /-suffix, valid UUIDv4 (uppercase)
		{"foo/bar/00000000-0000-4000-8000-AAAAAAAAAAAA", "", "", true},
		// not empty, with /-prefix, with multiple /, with /-suffix, valid UUIDv4 (lowercase)
		{"/foo/00000000-0000-4000-8000-aaaaaaaaaaaa/", "", "", true},
		// not empty, with /-prefix, with multiple /, with /-suffix, valid UUIDv4 (uppercase)
		{"/foo/00000000-0000-4000-8000-AAAAAAAAAAAA/", "", "", true},
	}

	for i, testCase := range testCases {
		pageURL, roomID, err := PageURLAndRoomIDFromPageID(testCase.pageID)
		if pageURL != testCase.pageURL {
			t.Errorf("Test Case %d: Got page URL \"%s\", expected \"%s\"", i, pageURL, testCase.pageURL)
		}
		if roomID != testCase.roomID {
			t.Errorf("Test Case %d: Got room ID \"%s\", expected \"%s\"", i, roomID, testCase.roomID)
		}
		if testCase.gotError && err == nil {
			t.Errorf("Test Case %d: Got no error, expected error", i)
		} else if !testCase.gotError && err != nil {
			t.Errorf("Test Case %d: Got error (%v), expected no error", i, err)
		}
	}
}
