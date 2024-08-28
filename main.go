package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"slices"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

var (
	GuildID      = flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
	BotToken     = flag.String("token", "", "Bot access token")
	RconHost     = flag.String("rconhost", "", "The ip and port of the minecraft rcon server")
	RconPassword = flag.String("rconpassword", "", "The password defined in your 'server.properties' file")
	BypassRoleID = flag.String("bypassrole", "", "The discord role ID to bypass the orb limit")
)

var bot *discordgo.Session

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "orb",
			Description: "Change your origin on the SMP. You must be logged in for this to work.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "username",
					Description: "Your minecraft username.",
					Required:    true,
				},
			},
		},
		{
			Name:        "remaining",
			Description: "Check how many orbs you have available to use.",
		},
	}

	commandHandlers = map[string]func(bot *discordgo.Session, i *discordgo.InteractionCreate){
		"orb": func(bot *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			username := options[0].StringValue()

			bypassLimit := slices.Contains(i.Member.Roles, *BypassRoleID)

			if !bypassLimit && !CanClaimOrb(i.Member.User.ID) {
				bot.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("✖ You have reached your limit of orb claims. (Max: `%d`)", MaxOrbs),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})

				return
			}

			err := ClaimOrb(i.Member.User.ID, username)

			if err != nil {
				if errors.Is(err, ErrPlayerNotFound) {
					bot.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: fmt.Sprintf("✖ Could not find player: `%v` in the server. Make sure you are in the server before redeeming an orb.", username),
						},
					})

					return
				}

				log.Printf("Error while redeeming orb for id '%s', user '%s': %v", i.Member.User.ID, username, err)

				bot.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "⚠ An error has occured while claiming the orb. Please try again or contact support.",
					},
				})
				
				return
			}

			var (
				usedOrbs string
				maxOrbs  string
			)

			if bypassLimit {
				usedOrbs = "∞"
				maxOrbs = "∞"
			} else {
				usedOrbs = strconv.Itoa(GetUsedOrbs(i.Member.User.ID))
				maxOrbs = strconv.Itoa(MaxOrbs)
			}

			bot.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("✔ Succesfully redeemed Orb on player: `%s`. (`%s/%s`)", username, usedOrbs, maxOrbs),
				},
			})
		},

		"remaining": func(bot *discordgo.Session, i *discordgo.InteractionCreate) {
			usedOrbs := GetUsedOrbs(i.Member.User.ID)

			bot.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("🟢 You have `%d` orbs remaining. Used `%d/%d`", MaxOrbs-usedOrbs, usedOrbs, MaxOrbs),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		},
	}
)

func init() { flag.Parse() }

func init() {
	var err error
	bot, err = discordgo.New("Bot " + *BotToken)

	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	bot.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	bot.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)

		registerCommands()
	})
}

func main() {
	err := bot.Open()

	if err != nil {
		log.Fatal(err)
	}

	defer bot.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop
}

func registerCommands() {
	log.Println("Registering commands...")

	var count int

	for _, v := range commands {
		_, err := bot.ApplicationCommandCreate(bot.State.User.ID, *GuildID, v)

		if err != nil {
			log.Fatalf("Failed to register command '%v': %v", v.Name, err)
		}

		count++
	}

	log.Printf("Registered %d commands.", count)
}
