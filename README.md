# youtubesearch
Inspired by the [youtube_search Python package](https://github.com/joetats/youtube_search). This library allows you to search for 
youtube videos without using the API.

# Example
```go
// Use without timeout
results, err := youtubesearch.Search("Rob Pike Go speech", 0)

// Use with timeout of 3 seconds
results, err := youtubesearch.Search("Rob Pike Go speech", time.Duration(3_000_000_000))
