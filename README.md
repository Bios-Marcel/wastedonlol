# wastedonlol

Back when I started playing League of Legends, there was a site called
[wastedonlol](https://wol.gg/). It was able to sum up the time you have spent
in all the matches that you've ever played. At some point however, Riot Games
removed access to that data from the public API. While the original page still
exists and also still displays an amount of hours, that's only the time spent
in ranked games. Normal games don't get taken into account anymore. Since I
wasn't able to find any script or tool that allows me to retrieve this data, I
simply wrote my own.

It's written in Golang and requires an API key and your summoner name.

Here's how to use it:

1. Generate an API key over at https://developer.riotgames.com/
2. Find out on which of these endpoints your account lies on:
  * ru
  * kr
  * br1
  * oc1
  * jp1
  * na1
  * eun1
  * euw1
  * tr1
  * la1
  * la2
3. Launch the tool:
```shell
go run . --server="euw1" --apikey="YOUR_KEY" --summonername="YOUR_SUMMONER_NAME"
```

Optionally you can set the flag `verbose` to true, in order to get additional
information while you are waiting. This is might be helpful, since it takes
quite a while to retrieve all your information, due to the fact that the LoL
API has quite the strict rate limiting.

> 20 requests per second
>
> 100 requests per 2 minutes
