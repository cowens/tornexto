# Provide a next bookmark for The Old Reader

[The Old Reader](http://theoldreader.com) is an RSS reader similar to Google Reader;
however, it is missing a vital feature that Google Reader had: a URL you could hit 
to get the next item from a folder.  This Go based Google App Engine application uses
the API for The Old Reader to provide that feature.

It is currently in a very rough, but working, state.  To use it, visit

    https://tornexto.appspot.com/auth?token=XXXXXX

where XXXXXX is the token describe in the [API](https://github.com/krasnoukhov/theoldreader-api/blob/master/README.md#getting-a-token).  The application will store the token as a cookie in your browser.  **IMPORTANT PRIVACY NOTE:** This token will allow the application to do basically whatever it wants with your feeds.  You can review the code here, but there is no guarantee that the version running at tornexto.appspot.com is the same version you see here.  If this is a problem for you, then you can take this code and modify the app name in the app.yml file and the name of the Go source code and run your own Google App Engine app.

Once you have registered your token, you can create a new bookmark with the URL

    https://tornext.appspot.com/next?folder=XXXXXX

where XXXXXX is the name of one of your folders.  In the near future I plan on making it easier to create these bookmarks by listing the folders and feeds you have available, but my first priority was to get something working at all.
