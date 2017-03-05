package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
    "os/exec"

	"github.com/bwmarrin/discordgo"
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()
}

var token string
var buffer = make([][]byte, 0)

func main() {
	if token == "" {
		fmt.Println("No token provided. Please run: playbot -t <bot token>")
		return
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	dg.AddHandler(ready)
	dg.AddHandler(messageCreate)
	dg.AddHandler(guildCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}

	fmt.Println("playbot is now running.  Press CTRL-C to exit.")

	<-make(chan struct{})
	return
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
}

var ch chan int = make(chan int,1)

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
    defer func(){<- ch}()

	if strings.HasPrefix(m.Content, "!play") {
        ch <- 1
		c, err := s.State.Channel(m.ChannelID)
		if err != nil {
			return
		}

		g, err := s.State.Guild(c.GuildID)
		if err != nil {
			return
		}

		for _, vs := range g.VoiceStates {
			if vs.UserID == m.Author.ID {
                download(m.Content)
                _ = loadSound()
				playSound(s, g.ID, vs.ChannelID)
				return
			}
		}
	}

	if strings.HasPrefix(m.Content, "!stop") {
		c, err := s.State.Channel(m.ChannelID)
		if err != nil {
			return
		}

		g, err := s.State.Guild(c.GuildID)
		if err != nil {
			return
		}

		for _, vs := range g.VoiceStates {
			if vs.UserID == m.Author.ID {
                vc, _ := s.ChannelVoiceJoin(g.ID, vs.ChannelID, false, true)
                _ = vc.Disconnect()
                _ = os.Remove("out.dca")
				return
			}
		}
    }
}

func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}

	for _, channel := range event.Guild.Channels {
		if channel.ID == event.Guild.ID {
			_, _ = s.ChannelMessageSend(channel.ID, "playbot is ready!")
			return
		}
	}
}

func loadSound() error {
	file, err := os.Open("out.dca")

	if err != nil {
		fmt.Println("Error opening dca file :", err)
		return err
	}

	var opuslen int16

	for {
		err = binary.Read(file, binary.LittleEndian, &opuslen)

		if err == io.EOF || err == io.ErrUnexpectedEOF {
			file.Close()
			return nil
		}

		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			return err
		}

		InBuf := make([]byte, opuslen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)

		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			return err
		}

		buffer = append(buffer, InBuf)
	}
}

func playSound(s *discordgo.Session, guildID, channelID string) (err error) {
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}

    time.Sleep(100 * time.Millisecond)

    _ = vc.Speaking(true)

    for _, buff := range buffer {
        vc.OpusSend <- buff
    }

    _ = vc.Speaking(false)

    _ = vc.Disconnect()

	return nil
}

func download(message string) {
    url := strings.Split(message," ")[1]
    exec.Command("./url2dca.sh",url).Run()
}
