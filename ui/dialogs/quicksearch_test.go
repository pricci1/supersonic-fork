package dialogs

import (
	"testing"

	"github.com/dweymouth/supersonic/backend/mediaprovider"
)

func TestQuickSearchTrackNavigationMenuItems(t *testing.T) {
	var navigatedType mediaprovider.ContentType
	var navigatedID string
	q := &QuickSearch{SearchDialog: &SearchDialog{}}
	q.SetOnNavigateTo(func(contentType mediaprovider.ContentType, id string) {
		navigatedType = contentType
		navigatedID = id
	})

	items := q.trackNavigationMenuItems(&mediaprovider.Track{
		ArtistIDs:   []string{"artist-id"},
		ArtistNames: []string{"Artist"},
		AlbumID:     "album-id",
	})
	if len(items) != 2 {
		t.Fatalf("expected 2 navigation menu items, got %d", len(items))
	}

	items[0].Action()
	if navigatedType != mediaprovider.ContentTypeArtist || navigatedID != "artist-id" {
		t.Fatalf("artist menu navigated to %v %q", navigatedType, navigatedID)
	}

	items[1].Action()
	if navigatedType != mediaprovider.ContentTypeAlbum || navigatedID != "album-id" {
		t.Fatalf("album menu navigated to %v %q", navigatedType, navigatedID)
	}
}

func TestQuickSearchGoToArtistMenuItemMultipleArtists(t *testing.T) {
	var navigatedType mediaprovider.ContentType
	var navigatedID string
	q := &QuickSearch{SearchDialog: &SearchDialog{}}
	q.SetOnNavigateTo(func(contentType mediaprovider.ContentType, id string) {
		navigatedType = contentType
		navigatedID = id
	})

	item := q.goToArtistMenuItem([]string{"artist-1", "artist-2"}, []string{"Artist 1", "Artist 2"})
	if item.Disabled {
		t.Fatal("expected artist menu item to be enabled")
	}
	if item.ChildMenu == nil || len(item.ChildMenu.Items) != 2 {
		t.Fatalf("expected child menu with 2 artists, got %#v", item.ChildMenu)
	}

	item.ChildMenu.Items[1].Action()
	if navigatedType != mediaprovider.ContentTypeArtist || navigatedID != "artist-2" {
		t.Fatalf("artist child menu navigated to %v %q", navigatedType, navigatedID)
	}
}

func TestQuickSearchNavigationMenuItemsDisabledWithoutIDs(t *testing.T) {
	q := &QuickSearch{SearchDialog: &SearchDialog{}}
	q.SetOnNavigateTo(func(mediaprovider.ContentType, string) {})

	if item := q.goToArtistMenuItem(nil, nil); !item.Disabled {
		t.Fatal("expected artist navigation menu item to be disabled without artist IDs")
	}
	if item := q.goToAlbumMenuItem(""); !item.Disabled {
		t.Fatal("expected album navigation menu item to be disabled without an album ID")
	}
}
