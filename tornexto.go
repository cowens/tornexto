package main

import (
	"fmt"
	"net/http"
	"net/url"
	"io/ioutil"
	"appengine"
	"appengine/urlfetch"
	"encoding/json"
	"strings"
	"path"
)

func init() {
	http.HandleFunc("/auth", auth)
	http.HandleFunc("/home", home)
	http.HandleFunc("/next", next)
	http.HandleFunc("/nothing", nothing)
}

func nothing(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<html><body>no new items were found<br /><br />There may be more items <a href=\"https://theoldreader.com\">in other folders</a>.</body></html>")
}

func auth(w http.ResponseWriter, r *http.Request) {
	auth_token := r.FormValue("token")
	auth_cookie := http.Cookie{ Name: "auth", Value: auth_token, MaxAge: 0 }
	http.SetCookie(w, &auth_cookie)
	http.Redirect(w, r, "/home", http.StatusFound)
}

func home(w http.ResponseWriter, r *http.Request) {
	auth_cookie, err := r.Cookie("auth")
	if err != nil {
		fmt.Fprintf(w, "Could not find authorization, please go to http://tornexto.appspot.com/auth?token=XXXXXXXXXXXXXXXXX where XXXXXXXXXXXXXXXXX is your token.")
		return
	}
	auth_token := auth_cookie.Value

	fmt.Fprintf(w, "<html><body>token found, now drag one or more of these links to your bookmark toolbar:<br /><ul>")
	c := appengine.NewContext(r)
	client := urlfetch.Client(c)
	fmt.Fprintln(w, `<li><a href="/next">(all folders)</a>`)
	folders, _ := get_folders(c, client, auth_token)
	for _, folder := range folders {
		fmt.Fprintf(w, `<li><a href="/next?folder=` + folder + `">` + folder + "</a>\n")
	}
	fmt.Fprintf(w, "</ul></body></html>")
}

func next(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	auth_cookie, err := r.Cookie("auth")

	if err != nil {
		fmt.Fprintf(w, "Could not find authorization, please go to http://tornexto.appspot.com/auth?token=XXXXXXXXXXXXXXXXX where XXXXXXXXXXXXXXXXX is your token.")
		return
	}

	auth_token := auth_cookie.Value
	folder     := r.FormValue("folder")

	if auth_token == "" {
		fmt.Fprintf(w, "bad token")
		return
	}
	
	client := urlfetch.Client(c)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	next_id, err := get_next_id(client, folder, auth_token)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	if next_id == "" {
		http.Redirect(w, r, "/nothing", http.StatusFound)
		return
	}

	url, err := get_url_for_item(client, next_id, auth_token)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}
	mark_item_as_read(client, next_id, auth_token)

	http.Redirect(w, r, url, http.StatusFound)
}

func mark_item_as_read(client *http.Client, id string, auth_token string) error {
	url := "https://theoldreader.com/reader/api/0/edit-tag"
	args := "a=user/-/state/com.google/read&i=" + id
	req, _ := http.NewRequest("POST", url, strings.NewReader(args))
	req.Header.Set("Authorization", "GoogleLogin auth=" + auth_token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return nil
}

func get_folders(c appengine.Context, client *http.Client, auth_token string) ([]string, error) {
	url := "https://theoldreader.com/reader/api/0/tag/list?output=json"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "GoogleLogin auth=" + auth_token)
	ret := make([]string, 0)

	resp, err := client.Do(req)
	if err != nil {
		return ret, err
	}

	defer resp.Body.Close()
        json_bytes, _ := ioutil.ReadAll(resp.Body)

	type Folder struct {
		ID string
	}

	type Folders struct {
		Tags []Folder
	}

	var tags Folders
	json_err := json.Unmarshal(json_bytes, &tags)
	if json_err != nil {
		return ret, json_err
	}

	ret = make([]string, len(tags.Tags))
	for index, folder := range tags.Tags {
		ret[index] = path.Base(folder.ID)
	}

	return ret, nil
}

func get_url_for_item(client *http.Client, id string, auth_token string) (string, error) {
	url := "https://theoldreader.com/reader/api/0/stream/items/contents?output=json&i=" + id
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "GoogleLogin auth=" + auth_token)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
        json_bytes, _ := ioutil.ReadAll(resp.Body)

	type Item struct {
		Href string
	}

	type Items struct {
		Canonical []Item
	}

	type FolderItem struct {
		Items []Items
	}

	var folder_item FolderItem
	json_err := json.Unmarshal(json_bytes, &folder_item)
	if json_err != nil {
		return "", json_err
	}

	return string(folder_item.Items[0].Canonical[0].Href), nil
}

func get_next_id(client *http.Client, folder string, auth_token string) (string, error) {
	filter := "user/-/label/" + folder
	if folder == "" {
		filter = "user/-/state/com.google/reading-list"
	}
	url    := "https://theoldreader.com/reader/api/0/stream/items/ids?output=json&xt=user/-/state/com.google/read&r=o&s=" + url.QueryEscape(filter)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "GoogleLogin auth=" + auth_token)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
        json_bytes, _ := ioutil.ReadAll(resp.Body)

	type Item struct {
		ID string
	}
	type ItemRefs struct {
		ItemRefs []Item
	}

	var items ItemRefs
	json_err := json.Unmarshal(json_bytes, &items)
	if json_err != nil {
		return "", json_err
	}

	if len(items.ItemRefs) == 0 {
		return "", nil
	}

	next_id := items.ItemRefs[0].ID
	return next_id, nil
}
