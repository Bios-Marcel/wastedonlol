package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type Summoner struct {
	AccountID string `json:"accountId"`
}

type Matches struct {
	Matches []*Match `json:"matches"`
}

type MatchDuration struct {
	Duration uint64 `json:"gameDuration"`
}

type Match struct {
	GameID int `json:"gameId"`
}

var (
	ctx                = context.Background()
	secondRateLimit    = rate.NewLimiter(rate.Every(1*time.Second), 19)
	twoMinuteRateLimit = rate.NewLimiter(rate.Every(2*time.Minute), 99)

	server       string
	apiKey       string
	summonerName string
	verbose      bool
)

func init() {
	flag.StringVar(&server, "server", "euw1", "specifies which server region should be queried")
	flag.StringVar(&summonerName, "summonername", "", "specifies the summoner which to query for")
	flag.StringVar(&apiKey, "apikey", "", "specifies the api key")
	flag.BoolVar(&verbose, "verbose", false, "specifies whether additional information should be given out.")

	flag.Parse()
}

func main() {
	log.Printf("Querying summoner '%s' in server region '%s'\n", summonerName, server)

	if apiKey == "" {
		log.Fatalln("Please supply an API key by setting the parameter 'apikey'.")
	}

	if summonerName == "" {
		log.Fatalln("Please specify a summoner name by setting the parameter 'summonername'.")
	}

	summoner, err := getSummoner()
	if err != nil {
		log.Fatalf("Error retrieving summoner information: %s", err)
	}

	if verbose {
		log.Printf("Successfully received Account ID: %s\n", summoner.AccountID)
	}

	matches, err := getMatches(summoner)
	if err != nil {
		log.Fatalf("Error retrieving match history: %s\n", err)
	}

	if verbose {
		log.Printf("Retrieved %d matches\n", len(matches))
	}

	duration, sumError := sumDurationAsHours(matches)
	if sumError != nil {
		log.Fatalf("Error summing up game durations: %s", sumError)
	}

	log.Printf("You have played %d matches, which cost you %d hours of your life.", len(matches), duration)
}

func ratelimitCheck() {
	secondRateLimit.Wait(ctx)
	twoMinuteRateLimit.Wait(ctx)
}

func sumDurationAsHours(matches []*Match) (uint64, error) {
	var duration uint64

	for index, match := range matches {
		matchDuration, err := getMatchDuration(match.GameID)
		if err != nil {
			return 0, err
		}

		if verbose && index%19 == 0 {
			log.Printf("So far %d matches have been retrieved, accounting for %d hours of "+
				"playtime.\n%d matches left to query", index+1, duration/60/60, len(matches)-index+1)
		}

		duration += matchDuration
	}

	return duration / 60 / 60, nil
}

func getMatchDuration(matchID int) (uint64, error) {
	ratelimitCheck()
	response, readError := http.Get(fmt.Sprintf("https://%s.api.riotgames.com/lol/match/v4/matches/%d?api_key=%s", server, matchID, apiKey))
	if readError != nil {
		return 0, readError
	}

	rawResponse, readError := ioutil.ReadAll(response.Body)
	if readError != nil {
		return 0, readError
	}

	matchDuration := &MatchDuration{}
	unmarshalError := json.Unmarshal(rawResponse, matchDuration)
	if unmarshalError != nil {
		return 0, unmarshalError
	}

	return matchDuration.Duration, nil
}

func getMatches(summoner *Summoner) ([]*Match, error) {
	var matches []*Match
	nextIndex := 0

	for {
		ratelimitCheck()
		response, readError := http.Get(fmt.Sprintf("https://%s.api.riotgames.com/lol/match/v4/matchlists/by-account/%s?beginIndex=%d&api_key=%s", server, summoner.AccountID, nextIndex, apiKey))
		if readError != nil {
			return nil, readError
		}

		rawResponse, readError := ioutil.ReadAll(response.Body)
		if readError != nil {
			return nil, readError
		}

		responseMatches := &Matches{}
		unmarshalError := json.Unmarshal(rawResponse, responseMatches)
		if unmarshalError != nil {
			return nil, readError
		}

		matches = append(matches, responseMatches.Matches...)
		if len(responseMatches.Matches) != 100 {
			break
		} else {
			nextIndex += 100
		}
	}

	return matches, nil
}

func getSummoner() (*Summoner, error) {
	ratelimitCheck()
	response, readError := http.Get(fmt.Sprintf("https://%s.api.riotgames.com/lol/summoner/v4/summoners/by-name/%s?api_key=%s", server, summonerName, apiKey))
	if readError != nil {
		return nil, readError
	}

	rawResponse, readError := ioutil.ReadAll(response.Body)
	if readError != nil {
		return nil, readError
	}

	summoner := &Summoner{}
	unmarshalError := json.Unmarshal(rawResponse, summoner)
	if unmarshalError != nil {
		return nil, unmarshalError
	}

	return summoner, nil
}
