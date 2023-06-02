package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const TwitterUrl = "https://api.twitter.com"
const AuthUrl = TwitterUrl + "/oauth2/token"
const RetweetsUrl = TwitterUrl + "/1.1/statuses/retweets/"

type Login struct {
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
}

type Retweets []struct {
	User User `json:"user"`
}

type User struct {
	ID int `json:"id"`
}

func getEncodingCredentials() string {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file check if exist: %v", err)
	}

	data := os.Getenv("API_KEY") + ":" + os.Getenv("SECRET_KEY")

	return base64.StdEncoding.EncodeToString([]byte(data))
}

func login() Login {
	req, _ := http.NewRequest(http.MethodPost, AuthUrl, bytes.NewReader([]byte("grant_type=client_credentials")))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	req.Header.Add("Authorization", "Basic "+getEncodingCredentials())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error during login request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error during read of login response: %v", err)
	}

	var login Login
	if err := json.Unmarshal(body, &login); err != nil {
		log.Fatalf("Can not unmarshal JSON: %v", err)
	}

	return login
}

func getRetweets(id int) Retweets {
	req, _ := http.NewRequest(http.MethodGet, RetweetsUrl+strconv.Itoa(id)+".json?trim_user=true", nil)
	req.Header.Add("Authorization", "Bearer "+login().AccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error during login request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error during read of login response: %v", err)
	}

	var retweets Retweets
	if err := json.Unmarshal(body, &retweets); err != nil {
		log.Fatalf("Can not unmarshal JSON: %v", err)
	}

	return retweets
}

func readFile() (Retweets, error) {
	f, err := os.Open("retweets.txt")

	if err != nil {
		log.Fatalf("Unable to read file: %v", err)
	}
	defer f.Close()

	var retweets Retweets

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		retweetID, err := strconv.Atoi(line)
		if err != nil {
			return nil, fmt.Errorf("error parsing retweet ID: %v", err)
		}

		retweet := struct {
			User User `json:"user"`
		}{}
		retweet.User.ID = retweetID
		retweets = append(retweets, retweet)
	}

	if scanner.Err() != nil {
		return nil, fmt.Errorf("error when reading file: %v", scanner.Err())
	}

	return retweets, nil
}

func (retweets *Retweets) Store() error {
	f, err := os.Create("retweets.txt")
	if err != nil {
		log.Fatalf("Unable to create file: %v", err)
	}
	defer f.Close()

	existingRetweets, err := readFile()
	if err != nil {
		return err
	}

	mergedRetweets := mergeRetweets(existingRetweets, *retweets)

	for _, retweet := range mergedRetweets {
		_, err := f.WriteString(fmt.Sprintf("%d\n", retweet.User.ID))

		if err != nil {
			log.Fatalf("Unable to write in file: %v", err)
		}
	}

	return nil
}

func mergeRetweets(existing, new Retweets) Retweets {
	merged := existing

	for _, retweet := range new {
		found := false
		for _, existingRetweet := range existing {
			if retweet.User.ID == existingRetweet.User.ID {
				found = true
				break
			}
		}
		if !found {
			merged = append(merged, retweet)
		}
	}

	return merged
}

func (retweets *Retweets) PickWinner() (User, error) {
	n := rand.Intn(len(*retweets))

	return (*retweets)[n].User, nil
}

func main() {
	id := flag.Int("id", 1633078861259759618, "The ID of a Twitter retweet")
	pickWinner := flag.Bool("pick", false, "Pick winner ?")
	flag.Parse()

	retweets := getRetweets(*id)
	retweets.Store()
	fmt.Println("New retweets saved in retweets.txt")

	if *pickWinner {
		winner, err := retweets.PickWinner()
		if err != nil {
			log.Fatalf("Error when pick winner: %v", err)
		}

		fmt.Println("The winner is:", winner.ID)
	}
}
