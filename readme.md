# RSS Subscriber API

## To Run

```go run .```

## To Use

* To get a list of current RSS feeds subscribed to:

```GET http://localhost:8000/api/feeds```

* To add, or edit, an existing RSS feed. If you don't provide an ID it will add a new one. If you do provide an ID it will edit the existing feed:

```PUT http://localhost:8000/api/feeds```

_Example Body:_
```
{
    "name": "Reuters Tech",
    "url": "http://feeds.reuters.com/reuters/technologyNews?format=xml",
    "category": "Technology",
    "provider": "Reuters"
}
```

* To get a list of all articles for all feeds:

``` GET http://localhost:8000/api/articles```

* To filter by category, or provider:

```GET http://localhost:8000/api/articles/technology```

```GET http://localhost:8000/api/articles/technology/BBC```

```GET http://localhost:8000/api/articles/{category}/{provider}```

* To sort by date published

```GET http://localhost:8000/api/articles?sort=date```

```GET {URL}?sort=date```

## Todo

1. Add Tests - Requires a lot of mocking.
2. Caching was working, but only for the first feed, so I've removed that. It should work for two feeds
3. Gorilla Mux Routing won't scale as more options get added, with more time I would refactor it to have optional querys.
4. With more time I'd add a Multistage Dockerfile to build and deploy this.
5. I added the gofeed library about halfway through after getting stuck on an XML encoding issue. I'd probably refactor the code a lot more to use the types from that 3rd party library.


## Assumptions

All the requested functionality can be implemented using this API. Not every function needs a separate endpoint.

Some of the listed client app features (showing the single news article as HTML for example) wouldn't be provided through this API - Once they have the RSS feed and links, they would go direct to the article linked from the news provider to download that as each news provider will have different html formatting.
