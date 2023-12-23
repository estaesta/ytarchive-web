make a website with go lang as it backend and htmx as it frontend
the website has auth when first opening
it will ask username and password. the username and password defined manually on the back end code
on front page, it has a form named "url" and a button "archive"
when submitted, it will run this command
```
ytarchive -v --debug -o "archive/%(channel)s/[%(upload_date)s]_%(title)s(%(id)s)/[%(upload_date)s]_%(title)s(%(id)s)" --add-metadata --write-thumbnail --write-description -merge --cookies cookies.txt -w {{ url }} best
```
it will then show the output process of that command in scrollable fragment/window

on the bottom side, it also show path tree of the "archive/" directory
