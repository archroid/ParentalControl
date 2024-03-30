package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"

	"github.com/bwmarrin/discordgo"
	log "github.com/charmbracelet/log"
	"github.com/joho/godotenv"
	"github.com/kbinani/screenshot"

	"github.com/mitchellh/go-ps"

	"github.com/nfnt/resize"
)

const (
	Intents = discordgo.IntentsDirectMessages |
		discordgo.IntentsGuildBans |
		discordgo.IntentsGuildEmojis |
		discordgo.IntentsGuildIntegrations |
		discordgo.IntentsGuildInvites |
		discordgo.IntentsGuildMembers |
		discordgo.IntentsGuildMessageReactions |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildVoiceStates |
		discordgo.IntentsGuilds |
		discordgo.IntentsGuildVoiceStates
)

var a = app.New()
var w fyne.Window
var datafolder = "D:/Apps/pctl/"
var myDiscordID string
var logFilePath = datafolder + "pctl.log"

func main() {

	// Load the .env file
	err := godotenv.Load(datafolder + ".env")
	if err != nil {
		log.Error("Error loading .env file", err)
	}

	myDiscordID = os.Getenv("MY_DISCORD_ID")

	// save logs into a file
	os.Remove(logFilePath)
	f, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Warn("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(io.MultiWriter(f, os.Stdout))

	// Start the discord bot in a goroutine
	go func() {
		dg, err := discordgo.New("Bot " + os.Getenv("DISCORD_BOT_TOKEN"))
		if err != nil {
			log.Error(err)
		}
		dg.Identify.Intents = discordgo.MakeIntent(Intents)

		dg.AddHandler(messageCreate)
		dg.AddHandler(ready)

		for {
			err = dg.Open()
			if err != nil {
				log.Error("Error opening connection to Discord", err)
				dg.Close()
				time.Sleep(5 * time.Second)
			} else {
				log.Info("Discord bot is now running")
				break
			}
		}

	}()

	// Start javaw running check in a goroutine

	var isLocked = os.Getenv("ISLOCKED")

	for {
		if isLocked == "TRUE" {
			if ifMineIsRunning() {
				w = a.NewWindow("Images")
				f, _ := os.Open(datafolder + "locked.png")
				uncodedimage, _, err := image.Decode(f)
				if err != nil {
					log.Error(err)
				}
				img := canvas.NewImageFromImage(uncodedimage)
				w.SetContent(img)
				w.Resize(fyne.NewSize(1280, 720))
				w.CenterOnScreen()
				w.SetPadded(false)
				w.SetFullScreen(true)
				w.RequestFocus()
				w.ShowAndRun()
				w.Close()
			} else {
				time.Sleep(1 * time.Minute)
			}
		} else {
			break
		}
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	channel, _ := s.UserChannelCreate(myDiscordID)
	s.ChannelMessageSend(
		myDiscordID,
		"Something went wrong while sending the DM!",
	)
	_, _ = s.ChannelMessageSend(channel.ID, "PC ON!\n Active Commands== \n /screenshot \n /exes \n /killjava \n /shutdown \n /log \n /unlock \n")

	s.UpdateListeningStatus("/Running")

	for {
		if ifMineIsRunning() {
			_, _ = s.ChannelMessageSend(channel.ID, "javaw.exe is running!")
			time.Sleep(30 * time.Minute)

		} else {
			time.Sleep(1 * time.Minute)
		}
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "/screenshot" {
		imgData, err := os.ReadFile(getScreenshot())
		if err != nil {
			log.Error("Error reading image file", err)
			return
		}

		file := &discordgo.File{
			Name:   "image.png",
			Reader: bytes.NewReader(imgData),
		}

		_, err = s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
			Files: []*discordgo.File{
				file,
			},
		})
		if err != nil {
			log.Error("Error sending image", err)
		}
	}

	if m.Content == "/exes" {
		exes := exes()
		for i := 0; i < len(exes); i += 2000 {
			end := i + 2000
			if end > len(exes) {
				end = len(exes)
			}
			_, err := s.ChannelMessageSend(m.ChannelID, exes[i:end])
			if err != nil {
				log.Error(err)
			}
		}
	}

	if m.Content == "/killjava" {
		s.ChannelMessageSend(m.ChannelID, killjavaw())

	}

	if m.Content == "/shutdown" {
		s.ChannelMessageSend(m.ChannelID, "shutting down")

		cmd := exec.Command("shutdown", "/s", "/hybrid", "/f", "/t", "0")
		err := cmd.Run()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "error shutting down")

		}
	}

	if m.Content == "/log" {

		file, err := os.ReadFile(logFilePath)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "error reading log")
		} else {
			s.ChannelMessageSend(m.ChannelID, string(file))
		}

	}

	if m.Content == "/unlock" {
		w.Close()
	}

}

func getScreenshot() string {
	bounds := screenshot.GetDisplayBounds(0)

	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		panic(err)
	}

	resizedImg := resize.Resize(1024, 576, img, resize.Lanczos3)
	fileName := fmt.Sprintf(datafolder + "screenshot" + ".png")
	file, _ := os.Create(fileName)
	defer file.Close()
	png.Encode(file, resizedImg)

	return fileName
}

func exes() string {
	processList, err := ps.Processes()
	if err != nil {
		log.Error("Error reading process list", err)
		return err.Error()
	}

	var processes string

	for x := range processList {
		process := processList[x]

		processes = processes + process.Executable() + "\n"

	}
	return processes

}

func ifMineIsRunning() bool {
	processList, err := ps.Processes()
	if err != nil {
		log.Error("Error reading process list", err)
		return false
	}

	for x := range processList {
		process := processList[x]

		if process.Executable() == "javaw.exe" {
			return true
		}

	}

	return false
}

func killjavaw() string {
	err := exec.Command("taskkill", "/im", "javaw.exe", "/f").Run()
	if err != nil {
		return "Error killing javaw.exe"
	} else {
		return "done"
	}
}

/*

Build Command:
env CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=/usr/bin/x86_64-w64-mingw32-gcc go build -ldflags="-H windowsgui"

*/
