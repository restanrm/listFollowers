package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"os/signal"
	"time"

	"github.com/dustin/go-nma"

	"github.com/restanrm/twitter"
)

var nmaKey = ""
var templates *template.Template

type Followers struct {
	Ids                 []int  `json:"ids"`
	Next_cursor         int    `json:"next_cursor"`
	Next_cursor_str     string `json:"next_cursor_str"`
	Previous_cursor     int    `json:"previous_cursor"`
	Previous_cursor_str string `json:"previous_cursor_str"`
}

func handleError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func isFollowerInFollowerList(id int, liste []int) bool {
	for i := 0; i < len(liste); i++ {
		if liste[i] == id {
			return true
		}
	}
	return false
}

type User struct {
	Name string `json:"screen_name"`
}

func (u User) String() string {
	return fmt.Sprintf("[%s]: https://twitter.com/%s", u.Name, u.Name)
}

func diffFollowers(prev, cur []int) ([]int, []int) {
	var followersLost []int
	var followersWin []int
	for _, elt := range prev {
		if !isFollowerInFollowerList(elt, cur) {
			followersLost = append(followersLost, elt)
		}
	}
	for _, elt := range cur {
		if !isFollowerInFollowerList(elt, prev) {
			followersWin = append(followersWin, elt)
		}
	}
	return followersLost, followersWin
}

func getFollowers(twitr twitter.Twitter, screen_name string) Followers {
	bFollowers, err := twitr.FollowersIds(screen_name, -1)
	handleError(err)

	var curFollowers Followers
	err = json.Unmarshal(bFollowers, &curFollowers)
	handleError(err)
	return curFollowers
}

func checkEnvVar(variable string) {
	if variable == "" {
		fmt.Println("Error of env variable, check that TWITTER_USERNAME, TWITTER_KEY, TWITTER_SECRET are defined")
		fmt.Println("NOTIFY_MY_ANDROID_KEY is optional")
		os.Exit(-1)
	}
}

func main() {
	var twitr twitter.Twitter
	templates = template.Must(template.New("nma").Parse(templateNMA))
	templates = template.Must(templates.New("console").Parse(templateNotifyConsole))

	username := os.Getenv("TWITTER_USERNAME")
	checkEnvVar(username)
	key := os.Getenv("TWITTER_KEY")
	checkEnvVar(key)
	secret := os.Getenv("TWITTER_SECRET")
	checkEnvVar(secret)
	nmaKey = os.Getenv("NOTIFY_MY_ANDROID_KEY") // this parameter is optional, no need to check
	n := nma.New(nmaKey)
	err := n.Verify(nmaKey)
	handleError(err)

	twitr = twitter.NewTwitter(key, secret)
	defer twitr.Close()

	done := make(chan bool)

	c := make(chan os.Signal)

	signal.Notify(c, os.Interrupt)

	go func(twitr twitter.Twitter, done <-chan bool) {
		var end bool
		var prevFollowers, curFollowers Followers
		prevFollowers = getFollowers(twitr, username)
		fmt.Printf("Initialisation: %v followers for user %v\n", len(prevFollowers.Ids), username)
		tick := time.Tick(1 * time.Minute)
		for !end {
			select {
			case <-done:
				end = true
			case <-tick:
				curFollowers = getFollowers(twitr, username)
				if len(curFollowers.Ids) == 0 {
					continue
				}
				lost, win := diffFollowers(prevFollowers.Ids, curFollowers.Ids)
				res := parseResult(twitr, result{lose: lost, win: win, source: curFollowers})
				notify(res)
				prevFollowers = curFollowers
			}
		}
	}(twitr, done)

	select {
	case <-c:
		done <- true
	}
}

type result struct {
	source Followers
	win    []int
	lose   []int
}

type Result struct {
	Source                  Followers
	WinMessage, LoseMessage string
	Win                     []User
	Lose                    []User
}

func parseResult(twitr twitter.Twitter, tab result) Result {
	var out Result
	out.Source = tab.source

	fn := func(dest *[]User, source []int) {
		for _, id := range source {
			var user User
			bUser, err := twitr.ShowId(id)
			handleError(err)
			err = json.Unmarshal(bUser, &user)
			handleError(err)
			*dest = append(*dest, user)
		}
	}

	if len(tab.lose) < 15 {
		fn(&out.Lose, tab.lose)
	} else {
		out.LoseMessage += "Too many users lost."
	}
	if len(tab.win) < 15 {
		fn(&out.Win, tab.win)
	} else {
		out.WinMessage += "Too many users won."
	}
	return out
}

type stringWriter struct {
	s string
}

func (s *stringWriter) Write(p []byte) (n int, err error) {
	for _, b := range p {
		s.s += string(b)
		n++
	}
	return
}

func (s stringWriter) String() string { return s.s }

func notify(result Result) {
	if result.Source.Ids == nil {
		return
	}
	err := templates.ExecuteTemplate(os.Stdout, "console", result)
	handleError(err)
	if nmaKey != "" {
		var s stringWriter
		err = templates.ExecuteTemplate(&s, "nma", result)
		handleError(err)
		if s.s == "" {
			return
		}
		n := nma.New(nmaKey)
		e := nma.Notification{
			Application: "ListFollowers",
			Event:       "Liste of followers",
			Description: s.s,
			ContentType: "text/html",
		}
		err = n.Notify(&e)
		handleError(err)
	}
}

var templateNMA = `{{if .LoseMessage}}<p>{{.LoseMessage}}</p>
{{else if .Lose}}<h3>Lose Followers</h3>
{{range .Lose}}<p>{{.String}}</p>
{{end}}{{end}}{{if .WinMessage}}<p>{{.WinMessage}}</p>
{{else if .Win}}<h3>Win Followers:</h3>
{{range .Win}}<p>{{.String}}</p>
{{end}}{{end}}`

var templateNotifyConsole = `{{if .LoseMessage}}{{.LoseMessage}}
{{else if .Lose}}Lose Followers:
{{range .Lose}}{{.String}}
{{end}}{{end}}{{if .WinMessage}}{{.WinMessage}}
{{else if .Win}}Win Followers:
{{range .Win}}{{.String}}
{{end}}{{end}}`
