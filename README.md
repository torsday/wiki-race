# wRace

---

# Description

- Find the shortest path between two Wikipedia articles.

---
# Architecture

- We use a graph to represents wikipedia pages (one page to a node), and then traverse the graph using:
    1. Breadth First Search, in conjunction with 
    2. concurrent, bi-directional graph traversal from both ends (start and finish).

---
# API

```txt
GET http://localhost:8083/wiki-race/goLang
``` 

With `start` and `destination` params, like so:

<http://localhost:8083/wiki-race/goLang?start=St.%20Olaf%20College&destination=pantheon%20(religion)>

```txt
GET /wiki-race/goLang?start=St.%20Olaf%20College&destination=pantheon%20(religion) HTTP/1.1
Host: localhost:8083
Connection: close
User-Agent: Paw/3.1.10 (Macintosh; OS X/10.15.4) GCDHTTPRequest
```

---

# Algorithm Complexity

Uni-directional traversal would have been `[O(b^d)]`, with bidirectional approach, we decrease that to `[O(b^(d/2) + b^(d/2)]`

---
# Setting up & Running

## Building

```sh
make build
```

## Running

```sh
make run
```

You can now use a tool like Postman to query the API, as described above.

## Testing

```sh
make test
```

This runs the Unit Tests.

## Stopping

```sh
make stop
```

# Notes

## Limitations

- To make the paths more fun, I was liberal in what I disqualified from the path (e.g. `Wayback_Machine` links, which are quite prevalent).

## Personal Process

- 3.5 hours learning go and coding. I've only ever done one little project in GoLang, and that was a year ago; I've been looking for an excuse to play with it since.
- 1.5 hours researching options, and fleshing out designs. Most of which didn't get built in the interest of my limited time (and focussing on learning Go); although I would like to tweak it to use the wikimedia api, as well as experiment with graph representation within Redis and/or kafka for horizontal scaling. 
  - I looked into using the Wikimedia api (different than the Wikipedia one), and found that it returns many links not actually on the pages. Possibly because it looks at the history of a page, rather than its current state. Because of that, I chose to use a link scraper instead.
- 1.0 hours researching REST framework options, and learning/building-out gorilla/mux.
- 1.0 hours learning and implementing unit testing in go; with more time I would refactor to increase code coverage dependency injection.
