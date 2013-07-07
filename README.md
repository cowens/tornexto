Provide a next bookmark for The Old Reader
==========================================

How to use tornexto
-------------------

[The Old Reader](http://theoldreader.com) is an RSS reader similar to Google Reader;
however, it is missing a vital feature that Google Reader had: a URL you could hit 
to get the next item from a folder.  This Go based Google App Engine application uses
the API for The Old Reader to provide that feature.

It is currently in a very rough, but working, state.  To use it, you must first
provide your [The Old Reader API token](https://theoldreader.com/reader/api/0/token)
to it.  This token can be retrieved from the link above if you are already
logged in to The Old Reader (you may have to set a password if you are using
Google to sign in).  Once you have your token, you must visit the [tornexto
auth page](https://tornexto.appspot.com/auth) and enter it into the form there.
Doing this will store a cookie in your web browser containing the token.

Once you have registered your token, you can create a new bookmark with the URL
by dragging one of the links on the [tornexto home page](https://tornexto.appspot.com/home)
to your bookmark toolbar or by adding a bookmark to 

    https://tornext.appspot.com/next?folder=XXXXXX

where XXXXXX is the name of one of your folders.

IMPORTANT PRIVACY NOTE:
-----------------------

Providing this token will allow tornexto to anything it wants with you The Old
Reader account (read items, mark items as read, etc).  Currently it only uses
this permission to ensure you have a valid token, fetch a list of your folders,
fetch a list of unread items in a folder you specify, and mark items as read
when they have been served to you.  If this is unacceptable to you, it is
possible to set up your own version of tornexto running in Google App Engine.
At a later date, detailed instructions will be provided in this document on how
to do that.
