# Twitter Contest CLI
Retrieves all users who have retweeted a tweet and selects a winner from among them.

## Usage
### Parameters
* **id** (required): The ID of the tweet you want
* **pick** (optionnal): If you want to get a winner

### Example
To run this tool :

```bash
go run main.go -id=<ID of a tweet> -pick
```

You can run the script periodically to populate the file with users who have retweeted over time and then, using the "pick" parameter, select a winner.
