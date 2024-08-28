package main

import (
	"fmt"

	"github.com/exzotic5485/orb-bot-go/store"
	"github.com/gorcon/rcon"
)

const MaxOrbs = 3

var orbStore = store.NewStore("orbs.json")

func CanClaimOrb(id string) bool {
	return orbStore.Get(id) < MaxOrbs
}

func ClaimOrb(id, username string) error {
	var r, err = rcon.Dial(*RconHost, *RconPassword)

	if err != nil {
		return err
	}

	defer r.Close()

	response, err := r.Execute(fmt.Sprintf("origin gui %s", username))

	if err != nil {
		return err
	}

	if response == "No player was found" {
		return ErrPlayerNotFound
	}

	orbStore.Increment(id, 1)

	return nil
}

func GetUsedOrbs(id string) int {
	return orbStore.Get(id)
}
