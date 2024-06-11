package modules

import (
	"math/rand"
	"time"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/http/views"
	"github.com/joeyak/go-twitch-eventsub/v2"
)

var phrases []string = []string{
	"https://cdn.discordapp.com/attachments/1177749122590179348/1240789868431278130/chopan25.png?ex=6660e39d&is=665f921d&hm=4573996efd63e327003ce21b23d857c949de40d9b2e7125271ef0af7d484d8d2&",
	"https://media.discordapp.net/attachments/1177749122590179348/1233512796717842473/chopan24.png?ex=6660c850&is=665f76d0&hm=d8734ae261c08fae5ab9f35b5a75dc95680930557e4ddd9ffb7789acec0e6dfd&=&format=webp&quality=lossless&width=360&height=57",
	"https://cdn.discordapp.com/attachments/1177749122590179348/1233532893901361183/image.png?ex=6660db07&is=665f8987&hm=e0d761b8af23ed432dbd5d176d30d721a78f8290e482df255286d3c0bc933155&",
	"https://cdn.discordapp.com/attachments/1177749122590179348/1232193984164659241/image.png?ex=66609952&is=665f47d2&hm=c08feac1148a0dc19b098cab8f81abfc0eb985ef00d4912b08e6eca882debe96&",
	"https://cdn.discordapp.com/attachments/1177749122590179348/1231464404596424826/bom-dia-engasgo.png?ex=666094d9&is=665f4359&hm=5c4132761bfd584dcdc8dbf0ad41c83a2b84d18999eb6fc8c092e67600177f7e&",
	"https://cdn.discordapp.com/attachments/1177749122590179348/1231463178060169247/pronto-nasci-pobre-mas-n-otario.png?ex=666093b5&is=665f4235&hm=3f9dc197fbd7ad719861f6737511bf35a5b3c24bc0816d339390d2afe89a5764&",
	"https://cdn.discordapp.com/attachments/1177749122590179348/1229895797257142362/chopan22.png?ex=6660ceb8&is=665f7d38&hm=2c0dbc6e6bc7727f0417fb4f8a762f8b1186fb47130a7a54e6b9ff711e77d9cc&",
	"https://media.discordapp.net/attachments/1110225341575860245/1162487161224970310/image.png?ex=6660ca89&is=665f7909&hm=4125ec7e1aa0692c556393435e72f0bc2e2db7fa0997948ef0fcdb832b7e8311&=&format=webp&quality=lossless&width=348&height=68",
	"https://cdn.discordapp.com/attachments/1177749122590179348/1225984725433323620/chopan21.png?ex=666114c0&is=665fc340&hm=0ee0891dd110d5eeaa98594059fd69dce70b3af3a77cfff7f3da1f6de291b4cf&",
}

const chopanRewardID = "b76fe0e6-48c4-40cf-91e6-990cad1f7217"

type ChopanModule struct {
	phrase int
}

func NewChopanModule() *ChopanModule {
	now := time.Now()
	seed := now.Year()*1000 + now.YearDay()
	rand.New(rand.NewSource(int64(seed)))

	phrase := rand.Intn(len(phrases))

	return &ChopanModule{
		phrase: phrase,
	}
}

// Register implements events.EventHandler.
func (m *ChopanModule) Register() {
	events.Register(m.handleReward)
}

func (m *ChopanModule) handleReward(reward *twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) error {
	if reward.Reward.ID != chopanRewardID {
		return nil
	}

	events.Dispatch(NewWebEvent("chopan_phrase", views.RenderToString(views.ChopanPhrase(phrases[m.phrase]))))

	return nil
}

var _ events.EventHandler = &ChopanModule{}
