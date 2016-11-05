package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/nlopes/slack"
)

var (
	conf        *Config
	dataDirPath string
	configPath  string
)

// Config holds all the necessary global parameters to run Rosie.
type Config struct {
	//RoomName is where the virtual pal will hang out.
	RoomName string `json:"default_room"`

	//FriendName is the name of the virtual pal we'll run.
	FriendName string `json:"friend_name"`

	//SlackKey is the API key we'll use to talk to Slack.  Don't make
	//this public!!!
	SlackKey string `json:"slack_key"`

	//SlackTeam is the team we'll connect to.
	SlackTeam string `json:"slack_team"`
}

// LoadConfig parses a config json file.
func LoadConfig(path string) (*Config, error) {
	var c *Config

	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	// decode the supplied JSON room data...
	decoder := json.NewDecoder(fd)
	if err := decoder.Decode(&c); err != nil {
		return nil, fmt.Errorf("Error parsing config file at %v: %v", path, err)
	}

	return c, nil
}

func init() {
	flag.StringVar(&dataDirPath, "data_dir_path", "./data/", "Path to data files")
	flag.StringVar(&configPath, "config_path", "./config.json", "Path to config file")
	flag.Parse()
}

func loop(rtm *slack.RTM, roomID string) {
	for {
		select {
		case msg := <-rtm.IncomingEvents:
			fmt.Printf("%+v\n", msg)
			switch ev := msg.Data.(type) {
			case *slack.ConnectedEvent:
				fmt.Println("Infos:", ev.Info)
				// Replace #general with your Channel ID
				rtm.SendMessage(rtm.NewOutgoingMessage("/me wakes up and looks around", roomID))
			case *slack.MessageEvent:
				fmt.Printf("Message: %v\n", ev)
			case *slack.RTMError:
				fmt.Printf("Error: %s\n", ev.Error())
			}
		}
	}
}

// Looks up a channel ID by its name.  Err is non-nil if it isn't found.
func channelIDByName(api *slack.Client, name string) (string, error) {
	channels, err := api.GetChannels(true)
	if err != nil {
		return "", fmt.Errorf("Can't get channels to enumerate: %v", err)
	}

	for _, c := range channels {
		if c.Name == conf.RoomName {
			return c.ID, nil
		}
	}
	return "", fmt.Errorf("Channel %s not found", name)
}

func main() {
	var err error
	conf, err = LoadConfig(configPath)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	api := slack.New(conf.SlackKey)
	api.SetDebug(false)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	id, err := channelIDByName(api, conf.RoomName)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	loop(rtm, id)
}
