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
	"time"
)

func init() {
	// http.HandleFunc("/", down);
	http.HandleFunc("/auth", auth)
	http.HandleFunc("/", home)
	http.HandleFunc("/home", home)
	http.HandleFunc("/next", next)
	http.HandleFunc("/nothing", nothing)
	http.HandleFunc("/logout", logout)
}

func down(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, `http://status.theoldreader.com`, http.StatusFound)
}

func nothing(w http.ResponseWriter, r *http.Request) {
	folder := r.FormValue("folder")
	if folder == "" {
		folder = "your folders"
	}
	client, auth_token := get_client(w, r)
	if client == nil {
		http.Redirect(w, r, `/auth`, http.StatusFound)
		return
	}
	name := get_name(client, auth_token)
	fmt.Fprint(w, `
		<html>
			<body>
				You are logged in as ` + name + `.  If this is not you, then you can
				<a href="/logout">logout</a>.
				<br />
				<br />
				No new items were found in ` + folder + `
				<br />
				<br />
				There may be more items <a href="https://theoldreader.com">in
				other folders</a>.  You could also try creating <a href="/home">
				a bookmarklet for a different folder</a>.
			</body>
		</html>
	`)
}

func auth(w http.ResponseWriter, r *http.Request) {
	auth_token := r.FormValue("token")
	error_text := r.FormValue("err")

	if auth_token != "" {
		set_cookie(w, auth_token)
		http.Redirect(w, r, "/home", http.StatusFound)
	}

	fmt.Fprint(w, `
		<html>
			<body>` + error_text + `
				<form method="GET" action="/auth">
					<label for="token">Enter token here:</label>
					<input id="token" type="password" name="token"></input>
					<input type="submit"></input>
				</form>
				If you do not know your token and you are logged in to
				<a href=https://theoldreader.com>The Old Reader</a>, 
				this <a href="https://theoldreader.com/reader/api/0/token">
				link</a> may list your token.  When you have your token come
				back to this page and enter it here.

			</body>
		</html>`)
	return
}

func get_client(w http.ResponseWriter, r *http.Request) (*http.Client, string) {
	auth_cookie, _ := r.Cookie("auth")

	if auth_cookie == nil {
		http.Redirect(w, r, `/auth`, http.StatusFound)
		return nil, ""
	}
	
	auth_token := auth_cookie.Value
	c          := appengine.NewContext(r)
	client     := &http.Client{
		Transport: &urlfetch.Transport{
			Context: c,
			Deadline: 15 * time.Second,
		},
	}

	if !verify_token(client, auth_token) {
		http.Redirect(w, r, `/auth?err=bad+token`, http.StatusFound)
		return nil, ""
	}

	set_cookie(w, auth_token);
	
	return client, auth_token
}

func logout(w http.ResponseWriter, r *http.Request) {
	set_cookie(w, "")
	http.Redirect(w, r, "/auth", http.StatusFound)
}

func home(w http.ResponseWriter, r *http.Request) {
	client, auth_token := get_client(w, r)
	if client == nil { return }

	name := get_name(client, auth_token)

	fmt.Fprintf(w, `
<html>
	<body>
		You are logged in as ` + name + `.  If this is not you, then you can
		<a href="/logout">logout</a>.
		<br /><br />
		Drag one or more of these links to your bookmark toolbar:<br />
		<ul>
	`)
	fmt.Fprintln(w, `<li><a href="/next">(all folders)</a>`)

	c          := appengine.NewContext(r)
	folders, _ := get_folders(c, client, auth_token)
	for _, folder := range folders {
		fmt.Fprintf(w, `<li><a href="/next?folder=` + folder + `">` + folder + "</a>\n")
	}
	fmt.Fprintf(w, "</ul></body></html>")
}

func next(w http.ResponseWriter, r *http.Request) {
	client, auth_token := get_client(w, r)
	if client == nil { return }

	folder := r.FormValue("folder")
	next_id, err := get_next_id(client, folder, auth_token)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	if next_id == "" {
		http.Redirect(w, r, "/nothing?folder=" + folder, http.StatusFound)
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

func get_name(client *http.Client, auth_token string) string {
	url := "https://theoldreader.com/reader/api/0/user-info?output=json"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "GoogleLogin auth=" + auth_token)

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}

	defer resp.Body.Close()
        json_bytes, _ := ioutil.ReadAll(resp.Body)


	var user_obj map[string]interface{}
	json_err := json.Unmarshal(json_bytes, &user_obj)
	if json_err != nil {
		return ""
	}

	return user_obj["userName"].(string)
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

func set_cookie(w http.ResponseWriter, auth_token string) {
	auth_cookie := http.Cookie{
		Name: "auth",
		//expires in roughly a year
		Expires: time.Now().Add(time.Hour * time.Duration(24*365)),
		Value: auth_token,
		MaxAge: 0,
	}
	http.SetCookie(w, &auth_cookie)
}

func verify_token(client *http.Client, auth_token string) bool {
	if auth_token == "" {
		return false
	}
	url := "https://theoldreader.com/reader/api/0/token"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "GoogleLogin auth=" + auth_token)
	
	resp, err := client.Do(req)
	if err != nil {
		return false
	}

        token_bytes, _ := ioutil.ReadAll(resp.Body)
	fetched_token  := string(token_bytes)

	return fetched_token == auth_token
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
