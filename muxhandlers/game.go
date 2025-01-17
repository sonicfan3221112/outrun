package muxhandlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/fluofoxxo/outrun/config"
	"github.com/fluofoxxo/outrun/consts"
	"github.com/fluofoxxo/outrun/db"
	"github.com/fluofoxxo/outrun/emess"
	"github.com/fluofoxxo/outrun/helper"
	"github.com/fluofoxxo/outrun/netobj"
	"github.com/fluofoxxo/outrun/requests"
	"github.com/fluofoxxo/outrun/responses"
	"github.com/fluofoxxo/outrun/status"
)

func GetDailyChallengeData(helper *helper.Helper) {
	// no player, agnostic
	baseInfo := helper.BaseInfo(emess.OK, status.OK)
	response := responses.DailyChallengeData(baseInfo)
	err := helper.SendResponse(response)
	if err != nil {
		helper.InternalErr("Error sending response", err)
	}
}

func GetCostList(helper *helper.Helper) {
	// no player, agonstic
	baseInfo := helper.BaseInfo(emess.OK, status.OK)
	response := responses.DefaultCostList(baseInfo)
	err := helper.SendResponse(response)
	if err != nil {
		helper.InternalErr("Error sending response", err)
	}
}

func GetMileageData(helper *helper.Helper) {
	player, err := helper.GetCallingPlayer()
	if err != nil {
		helper.InternalErr("Error getting player", err) // TODO: see if InternalErr is consistent with other usage of this context
		return
	}
	baseInfo := helper.BaseInfo(emess.OK, status.OK)
	response := responses.DefaultMileageData(baseInfo, player)
	err = helper.SendResponse(response)
	if err != nil {
		helper.InternalErr("Error sending response", err)
	}
}

func GetCampaignList(helper *helper.Helper) {
	baseInfo := helper.BaseInfo(emess.OK, status.OK)
	response := responses.DefaultCampaignList(baseInfo)
	err := helper.SendResponse(response)
	if err != nil {
		helper.InternalErr("Error sending response", err)
	}
}

func QuickActStart(helper *helper.Helper) {
	player, err := helper.GetCallingPlayer()
	if err != nil {
		helper.InternalErr("Error getting calling player", err)
		return
	}
	baseInfo := helper.BaseInfo(emess.OK, status.OK)
	response := responses.DefaultQuickActStart(baseInfo, player)
	err = helper.SendResponse(response)
	if err != nil {
		helper.InternalErr("Error sending response", err)
	}
}

func ActStart(helper *helper.Helper) {
	player, err := helper.GetCallingPlayer()
	if err != nil {
		helper.InternalErr("Error getting calling player", err)
		return
	}
	baseInfo := helper.BaseInfo(emess.OK, status.OK)
	response := responses.DefaultActStart(baseInfo, player)
	err = helper.SendResponse(response)
	if err != nil {
		helper.InternalErr("Error sending response", err)
	}
}

func ActRetry(helper *helper.Helper) {
	player, err := helper.GetCallingPlayer()
	if err != nil {
		helper.InternalErr("Error getting calling player", err)
		return
	}
	player.PlayerState.NumRedRings -= 5
	err = db.SavePlayer(player)
	if err != nil {
		helper.InternalErr("Error saving player", err)
		return
	}
	baseInfo := helper.BaseInfo(emess.OK, status.OK)
	response := responses.NewBaseResponse(baseInfo)
	err = helper.SendResponse(response)
	if err != nil {
		helper.InternalErr("Error sending response", err)
	}
}

func QuickPostGameResults(helper *helper.Helper) {
	recv := helper.GetGameRequest()
	var request requests.QuickPostGameResultsRequest
	err := json.Unmarshal(recv, &request)
	if err != nil {
		helper.Err("Error unmarshalling", err)
		return
	}
	player, err := helper.GetCallingPlayer()
	if err != nil {
		helper.InternalErr("Error getting calling player", err)
		return
	}

	mainC, err := player.GetMainChara()
	if err != nil {
		helper.InternalErr("Error getting main character", err)
		return
	}
	subC, err := player.GetSubChara()
	if err != nil {
		helper.InternalErr("Error getting sub character", err)
		return
	}
	mainCIndex := player.IndexOfChara(mainC.ID) // TODO: check if -1
	subCIndex := player.IndexOfChara(subC.ID)   // TODO: check if -1
	playCharacters := []netobj.Character{
		mainC,
		subC,
	}
	if request.Closed == 0 { // If the game wasn't exited out of
		player.PlayerState.NumRings += request.Rings
		player.PlayerState.NumRedRings += request.RedRings
		player.PlayerState.Animals += request.Animals
		playerTimedHighScore := player.PlayerState.TimedHighScore
		if request.Score > playerTimedHighScore {
			player.PlayerState.TimedHighScore = request.Score
		}
		//player.PlayerState.TotalDistance += request.Distance  // We don't do this in timed mode!
		// increase character(s)'s experience
		expIncrease := request.Rings + request.FailureRings // all rings collected
		abilityIndex := 1
		for abilityIndex == 1 { // unused ability is at index 1
			abilityIndex = rand.Intn(len(mainC.AbilityLevel))
		}
		// check that increases exist
		_, ok := consts.UpgradeIncreases[mainC.ID]
		if !ok {
			helper.InternalErr("Error getting upgrade increase", fmt.Errorf("no key '%s' in consts.UpgradeIncreases", mainC.ID))
			return
		}
		_, ok = consts.UpgradeIncreases[subC.ID]
		if !ok {
			helper.InternalErr("Error getting upgrade increase", fmt.Errorf("no key '%s' in consts.UpgradeIncreases", subC.ID))
			return
		}
		if mainC.Level < 100 {
			mainC.Exp += expIncrease
			for mainC.Exp >= mainC.Cost {
				// more exp than cost = level up
				mainC.Level++                                   // increase level
				mainC.AbilityLevel[abilityIndex]++              // increase ability level
				mainC.Exp -= mainC.Cost                         // remove cost from exp
				mainC.Cost += consts.UpgradeIncreases[mainC.ID] // increase cost
			}
		}
		if subC.Level < 100 {
			subC.Exp += expIncrease
			for subC.Exp >= subC.Cost {
				// more exp than cost = level up
				subC.Level++                                  // increase level
				subC.AbilityLevel[abilityIndex]++             // increase ability level
				subC.Exp -= subC.Cost                         // remove cost from exp
				subC.Cost += consts.UpgradeIncreases[subC.ID] // increase cost
			}
		}

		playCharacters = []netobj.Character{ // TODO: check if this redefinition is needed
			mainC,
			subC,
		}
		//err = db.SavePlayer(player)
	}

	/*
		if err != nil {
			helper.InternalErr("Error saving player", err)
			return
		}
	*/

	baseInfo := helper.BaseInfo(emess.OK, status.OK)
	response := responses.DefaultQuickPostGameResults(baseInfo, player, playCharacters)
	// apply the save after the response so that we don't break the leveling
	player.CharacterState[mainCIndex] = mainC
	player.CharacterState[subCIndex] = subC
	err = db.SavePlayer(player)
	if err != nil {
		helper.InternalErr("Error saving player", err)
		return
	}

	err = helper.SendResponse(response)
	if err != nil {
		helper.InternalErr("Error sending response", err)
	}
}

func PostGameResults(helper *helper.Helper) {
	recv := helper.GetGameRequest()
	var request requests.PostGameResultsRequest
	err := json.Unmarshal(recv, &request)
	if err != nil {
		helper.Err("Error unmarshalling", err)
		return
	}
	player, err := helper.GetCallingPlayer()
	if err != nil {
		helper.InternalErr("Error getting calling player", err)
		return
	}

	mainC, err := player.GetMainChara()
	if err != nil {
		helper.InternalErr("Error getting main character", err)
		return
	}
	subC, err := player.GetSubChara()
	if err != nil {
		helper.InternalErr("Error getting sub character", err)
		return
	}
	playCharacters := []netobj.Character{
		mainC,
		subC,
	}
	if request.Closed == 0 { // If the game wasn't exited out of
		player.PlayerState.NumRings += request.Rings
		player.PlayerState.NumRedRings += request.RedRings
		player.PlayerState.Animals += request.Animals
		playerHighScore := player.PlayerState.HighScore
		if request.Score > playerHighScore {
			player.PlayerState.HighScore = request.Score
		}
		player.PlayerState.TotalDistance += request.Distance
		// increase character(s)'s experience
		expIncrease := request.Rings + request.FailureRings // all rings collected
		abilityIndex := 1
		for abilityIndex == 1 { // unused ability is at index 1
			abilityIndex = rand.Intn(len(mainC.AbilityLevel))
		}
		// check that increases exist
		_, ok := consts.UpgradeIncreases[mainC.ID]
		if !ok {
			helper.InternalErr("Error getting upgrade increase", fmt.Errorf("no key '%s' in consts.UpgradeIncreases", mainC.ID))
			return
		}
		_, ok = consts.UpgradeIncreases[subC.ID]
		if !ok {
			helper.InternalErr("Error getting upgrade increase", fmt.Errorf("no key '%s' in consts.UpgradeIncreases", subC.ID))
			return
		}
		if mainC.Level < 100 {
			mainC.Exp += expIncrease
			for mainC.Exp >= mainC.Cost {
				// more exp than cost = level up
				mainC.Level++                                   // increase level
				mainC.AbilityLevel[abilityIndex]++              // increase ability level
				mainC.Exp -= mainC.Cost                         // remove cost from exp
				mainC.Cost += consts.UpgradeIncreases[mainC.ID] // increase cost
			}
		}
		if subC.Level < 100 {
			subC.Exp += expIncrease
			for subC.Exp >= subC.Cost {
				// more exp than cost = level up
				subC.Level++                                  // increase level
				subC.AbilityLevel[abilityIndex]++             // increase ability level
				subC.Exp -= subC.Cost                         // remove cost from exp
				subC.Cost += consts.UpgradeIncreases[subC.ID] // increase cost
			}
		}

		playCharacters = []netobj.Character{ // TODO: check if this redefinition is needed
			mainC,
			subC,
		}
		//err = db.SavePlayer(player)
	}

	if config.CFile.Debug {
		helper.Out("Pre-function")
		helper.Out(strconv.Itoa(int(player.MileageMapState.Chapter)))
		helper.Out(strconv.Itoa(int(player.MileageMapState.Episode)))
		helper.Out(strconv.Itoa(int(player.MileageMapState.StageTotalScore)))
		helper.Out(strconv.Itoa(int(player.MileageMapState.Point)))
		helper.Out(strconv.Itoa(int(request.Score)))
	}
	player.MileageMapState.StageTotalScore += request.Score
	if player.MileageMapState.Point >= 5 {
		player.MileageMapState.Episode++
		player.MileageMapState.Point = 1
		player.MileageMapState.StageTotalScore = 0
	}
	neededPoint := func() int64 {
		ints, ok := consts.PointScores[player.MileageMapState.Episode]
		if !ok {
			return -2
		}
		for i, v := range ints {
			if v > player.MileageMapState.StageTotalScore {
				// This point is higher than where we've hit
				return int64(i)
			}
		}
		return int64(len(ints) - 1)
	}()
	if config.CFile.Debug {
		helper.Out("neededPoint: " + strconv.Itoa(int(neededPoint)))
	}
	if neededPoint == -2 {
		// Error (point scores not implemented yet!)
		// reset the episode and chapter and point
		if config.CFile.Debug {
			helper.Out("NEEDEDPOINT == -2")
		}
		player.MileageMapState.Episode = 1
		player.MileageMapState.Chapter = 1
		player.MileageMapState.Point = 1
		player.MileageMapState.StageTotalScore = 0
		neededPoint = 0
	}
	player.MileageMapState.Point = neededPoint

	if config.CFile.Debug {
		helper.Out("AFTER")
		helper.Out(strconv.Itoa(int(player.MileageMapState.Chapter)))
		helper.Out(strconv.Itoa(int(player.MileageMapState.Episode)))
		helper.Out(strconv.Itoa(int(player.MileageMapState.StageTotalScore)))
		helper.Out(strconv.Itoa(int(player.MileageMapState.Point)))
		helper.Out(strconv.Itoa(int(request.Score)))
	}

	mainCIndex := player.IndexOfChara(mainC.ID) // TODO: check if -1
	subCIndex := player.IndexOfChara(subC.ID)   // TODO: check if -1

	baseInfo := helper.BaseInfo(emess.OK, status.OK)
	response := responses.DefaultPostGameResults(baseInfo, player, playCharacters)
	// apply the save after the response so that we don't break the leveling
	player.CharacterState[mainCIndex] = mainC
	player.CharacterState[subCIndex] = subC
	err = db.SavePlayer(player)
	if err != nil {
		helper.InternalErr("Error saving player", err)
		return
	}

	err = helper.SendResponse(response)
	if err != nil {
		helper.InternalErr("Error sending response", err)
	}
}

func GetFreeItemList(helper *helper.Helper) {
	// Probably agnostic...
	baseInfo := helper.BaseInfo(emess.OK, status.OK)
	response := responses.DefaultFreeItemList(baseInfo)
	err := helper.SendResponse(response)
	if err != nil {
		helper.InternalErr("Error sending response", err)
	}
}

func GetMileageReward(helper *helper.Helper) {
	player, err := helper.GetCallingPlayer()
	if err != nil {
		helper.InternalErr("Error getting calling player", err)
		return
	}
	baseInfo := helper.BaseInfo(emess.OK, status.OK)
	response := responses.DefaultMileageReward(baseInfo, player)
	err = helper.SendResponse(response)
	if err != nil {
		helper.InternalErr("Error sending response", err)
	}
}
