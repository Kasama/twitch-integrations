package modules

import (
	"time"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	"github.com/joeyak/go-twitch-eventsub/v2"
	"github.com/nicklaw5/helix/v2"
)

const rewardIDPunish = ""
const punishDuration = 5 * time.Minute
const punishThreshold = -5 * time.Minute

type PunishableRedeemInfo struct {
	target     string
	prettyName string
	time       time.Time
}

func NewPunishableRedeemInfo(target string, prettyName string) *PunishableRedeemInfo {
	return &PunishableRedeemInfo{
		target:     target,
		prettyName: prettyName,
		time:       time.Now(),
	}
}

type PunishRewardsModule struct {
	helixClient   *helix.Client
	broadcasterID string
	redeems       []PunishableRedeemInfo
}

func NewPunishRewardsModule(broadcasterID string) *PunishRewardsModule {
	return &PunishRewardsModule{
		broadcasterID: broadcasterID,
		redeems:       make([]PunishableRedeemInfo, 0),
	}
}

// Register implements events.EventHandler.
func (m *PunishRewardsModule) Register() {
	events.Register(m.handleHelixClient)
	events.Register(m.handleNewRedeem)
	events.Register(m.handlePunishReward)
}

func (m *PunishRewardsModule) handleHelixClient(client *helix.Client) error {
	m.helixClient = client
	return nil
}

func (m *PunishRewardsModule) handleNewRedeem(info *PunishableRedeemInfo) error {
	m.redeems = append(m.redeems, *info)
	return nil
}

func (m *PunishRewardsModule) handlePunishReward(reward *twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) error {
	if reward.Reward.ID != rewardIDPunish {
		return nil
	}
	if len(m.redeems) == 0 {
		return nil
	}

	threshold := time.Now().Add(punishThreshold)

	for _, redeem := range m.redeems {
		if redeem.time.Before(threshold) {
			continue
		}
		_, err := m.helixClient.BanUser(&helix.BanUserParams{
			BroadcasterID: m.broadcasterID,
			ModeratorId:   m.broadcasterID,
			Body: helix.BanUserRequestBody{
				Duration: int(punishDuration.Truncate(time.Second).Seconds()),
				Reason:   "Punição por usar rewards contra o Kasama",
				UserId:   redeem.target,
			},
		})

		if err != nil {
			logger.Debugf("Failed to ban user '%s': %v", redeem.target, err)
		}
	}
	m.redeems = make([]PunishableRedeemInfo, 0)

	return nil
}

var _ events.EventHandler = &PunishRewardsModule{}
