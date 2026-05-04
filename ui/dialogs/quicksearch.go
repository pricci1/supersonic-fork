package dialogs

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/dweymouth/supersonic/backend/mediaprovider"
	myTheme "github.com/dweymouth/supersonic/ui/theme"
	"github.com/dweymouth/supersonic/ui/util"
)

type QuickSearch struct {
	SearchDialog *SearchDialog
	results      []*mediaprovider.SearchResult
	mp           mediaprovider.MediaProvider

	OnPlay          func(t mediaprovider.ContentType, id string, item any, shuffle bool)
	OnAddToQueue    func(t mediaprovider.ContentType, id string, item any, next bool)
	OnAddToPlaylist func(t mediaprovider.ContentType, id string, item any)
	OnSetFavorite   func(trackID string, fav bool)
	OnSetRating     func(trackID string, rating int)
	OnDownload      func(track *mediaprovider.Track)
	OnShare         func(trackID string)
	OnPlaySongRadio func(track *mediaprovider.Track)
	OnShowTrackInfo func(track *mediaprovider.Track)
}

func NewQuickSearch(mp mediaprovider.MediaProvider, im util.ImageFetcher) *QuickSearch {
	q := &QuickSearch{mp: mp}
	q.SearchDialog = NewSearchDialog(im, lang.L("Search Everywhere"), lang.L("Close"), q.onSearched)
	q.SearchDialog.OnShowContextMenu = q.showMenu
	return q
}

func (q *QuickSearch) onSearched(query string) []*mediaprovider.SearchResult {
	if query != "" {
		if res, err := q.mp.SearchAll(query, 50); err != nil {
			q.results = nil
			log.Printf("Error searching: %s", err.Error())
		} else {
			q.results = res
		}
	}
	return q.results
}

func (q *QuickSearch) showMenu(idx int, pos fyne.Position) {
	cType := q.results[idx].Type
	id := q.results[idx].ID
	item := q.results[idx].Item

	canvas := fyne.CurrentApp().Driver().CanvasForObject(q.SearchDialog)

	switch cType {
	case mediaprovider.ContentTypeTrack:
		track := item.(*mediaprovider.Track)
		menu := util.NewTrackContextMenu(false, q.trackNavigationMenuItems(track))
		menu.OnPlay = func(shuffle bool) {
			q.OnPlay(cType, id, item, shuffle)
		}
		menu.OnAddToQueue = func(next bool) {
			q.OnAddToQueue(cType, id, item, next)
		}
		menu.OnAddToPlaylist = func() {
			q.OnAddToPlaylist(cType, id, item)
		}
		menu.OnDownload = func() {
			q.OnDownload(item.(*mediaprovider.Track))
		}
		menu.OnFavorite = func(fav bool) {
			q.OnSetFavorite(id, fav)
		}
		menu.OnSetRating = func(rating int) {
			q.OnSetRating(id, rating)
		}
		menu.OnPlaySongRadio = func() {
			q.OnPlaySongRadio(track)
		}
		menu.OnShowInfo = func() {
			q.OnShowTrackInfo(track)
		}
		menu.OnShare = func() {
			q.OnShare(id)
		}
		menu.ShowAtPosition(pos, canvas)
	default:
		play := fyne.NewMenuItem(lang.L("Play"), func() {
			q.OnPlay(cType, id, item, false)
		})
		play.Icon = theme.MediaPlayIcon()
		shuffle := fyne.NewMenuItem(lang.L("Shuffle"), func() {
			q.OnPlay(cType, id, item, true)
		})
		shuffle.Icon = myTheme.ShuffleIcon
		playNext := fyne.NewMenuItem(lang.L("Play next"), func() {
			q.OnAddToQueue(cType, id, item, true)
		})
		playNext.Icon = myTheme.PlayNextIcon
		add := fyne.NewMenuItem(lang.L("Add to queue"), func() {
			q.OnAddToQueue(cType, id, item, false)
		})
		add.Icon = theme.ContentAddIcon()
		menu := fyne.NewMenu("", play, shuffle, playNext, add)

		if cType != mediaprovider.ContentTypeRadioStation {
			playlist := fyne.NewMenuItem(lang.L("Add to playlist")+"...", func() {
				q.OnAddToPlaylist(cType, id, item)
			})
			playlist.Icon = myTheme.PlaylistIcon
			menu.Items = append(menu.Items, playlist)
		}
		if cType == mediaprovider.ContentTypeAlbum {
			if album, ok := item.(*mediaprovider.Album); ok {
				menu.Items = append(menu.Items, fyne.NewMenuItemSeparator())
				menu.Items = append(menu.Items, q.goToArtistMenuItem(album.ArtistIDs, album.ArtistNames))
			}
		}

		widget.ShowPopUpMenuAtPosition(menu, canvas, pos)
	}
}

func (q *QuickSearch) trackNavigationMenuItems(track *mediaprovider.Track) []*fyne.MenuItem {
	return []*fyne.MenuItem{
		q.goToArtistMenuItem(track.ArtistIDs, track.ArtistNames),
		q.goToAlbumMenuItem(track.AlbumID),
	}
}

func (q *QuickSearch) goToArtistMenuItem(artistIDs, artistNames []string) *fyne.MenuItem {
	item := fyne.NewMenuItem(lang.L("Go to artist"), nil)
	item.Icon = myTheme.ArtistIcon

	var childItems []*fyne.MenuItem
	for i, artistID := range artistIDs {
		if artistID == "" {
			continue
		}
		artistName := artistID
		if i < len(artistNames) && artistNames[i] != "" {
			artistName = artistNames[i]
		}
		childItems = append(childItems, fyne.NewMenuItem(artistName, func() {
			q.navigateTo(mediaprovider.ContentTypeArtist, artistID)
		}))
	}

	if len(childItems) == 0 || q.SearchDialog.OnNavigateTo == nil {
		item.Disabled = true
		return item
	}
	if len(childItems) == 1 {
		item.Action = childItems[0].Action
		return item
	}
	item.ChildMenu = fyne.NewMenu("", childItems...)
	return item
}

func (q *QuickSearch) goToAlbumMenuItem(albumID string) *fyne.MenuItem {
	item := fyne.NewMenuItem(lang.L("Go to album"), func() {
		q.navigateTo(mediaprovider.ContentTypeAlbum, albumID)
	})
	item.Icon = myTheme.AlbumIcon
	item.Disabled = albumID == "" || q.SearchDialog.OnNavigateTo == nil
	return item
}

func (q *QuickSearch) navigateTo(contentType mediaprovider.ContentType, id string) {
	if q.SearchDialog.OnNavigateTo != nil {
		q.SearchDialog.OnNavigateTo(contentType, id)
	}
}

func (q *QuickSearch) SetOnDismiss(onDismiss func()) {
	q.SearchDialog.OnDismiss = onDismiss
}

func (q *QuickSearch) SetOnNavigateTo(onNavigateTo func(mediaprovider.ContentType, string)) {
	q.SearchDialog.OnNavigateTo = onNavigateTo
}

func (q *QuickSearch) MinSize() fyne.Size {
	return q.SearchDialog.MinSize()
}

func (q *QuickSearch) GetSearchEntry() fyne.Focusable {
	return q.SearchDialog.GetSearchEntry()
}
