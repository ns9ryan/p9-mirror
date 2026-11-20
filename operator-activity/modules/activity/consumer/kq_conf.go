package consumer

import (
	"fmt"

	"oa.98ent.com/p9/operator-activity/modules/activity/plugin"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/service"
	"oa.98ent.com/p9/common/xkafka/topic"
)

// GetPluginKqConf 各活动插件 Kafka 消费配置；Brokers 由 StartKq 注入。
func GetPluginKqConf(groupPrefix string) map[string][]kq.KqConf {
	return map[string][]kq.KqConf{
		// 充值活动
		plugin.NameDeposit: {{
			ServiceConf: service.ServiceConf{Name: "activity-deposit", Mode: service.DevMode},
			Topic:       topic.GameUserDepositOrderApproved,
			Group:       fmt.Sprintf("%s.deposit", groupPrefix),
			Offset:      "last",
			Consumers:   8,
			Processors:  8,
			ForceCommit: false,
		}},
		// 投注活动
		plugin.NameBetting: {{
			ServiceConf: service.ServiceConf{Name: "activity-betting", Mode: service.DevMode},
			Topic:       topic.GameUserGameOrderSettled,
			Group:       fmt.Sprintf("%s.betting", groupPrefix),
			Offset:      "last",
			Consumers:   8,
			Processors:  8,
			ForceCommit: false,
		}},
		// 推荐活动
		plugin.NameReferral: {{
			ServiceConf: service.ServiceConf{Name: "activity-referral-register", Mode: service.DevMode},
			Topic:       topic.GameUserRegister,
			Group:       fmt.Sprintf("%s.referral", groupPrefix),
			Offset:      "last",
			Consumers:   8,
			Processors:  8,
			ForceCommit: false,
		}, {
			ServiceConf: service.ServiceConf{Name: "activity-referral-deposit", Mode: service.DevMode},
			Topic:       topic.GameUserDepositOrderApproved,
			Group:       fmt.Sprintf("%s.referral", groupPrefix),
			Offset:      "last",
			Consumers:   8,
			Processors:  8,
			ForceCommit: false,
		}, {
			ServiceConf: service.ServiceConf{Name: "activity-referral-settled", Mode: service.DevMode},
			Topic:       topic.GameUserGameOrderSettled,
			Group:       fmt.Sprintf("%s.referral", groupPrefix),
			Offset:      "last",
			Consumers:   8,
			Processors:  8,
			ForceCommit: false,
		}},
		// 幸运轮盘
		plugin.NameLuckyWheel: {{
			ServiceConf: service.ServiceConf{Name: "activity-luckywheel-settled", Mode: service.DevMode},
			Topic:       topic.GameUserGameOrderSettled,
			Group:       fmt.Sprintf("%s.luckywheel", groupPrefix),
			Offset:      "last",
			Consumers:   8,
			Processors:  8,
			ForceCommit: false,
		}, {
			ServiceConf: service.ServiceConf{Name: "activity-luckywheel-deposit", Mode: service.DevMode},
			Topic:       topic.GameUserDepositOrderApproved,
			Group:       fmt.Sprintf("%s.luckywheel", groupPrefix),
			Offset:      "last",
			Consumers:   8,
			Processors:  8,
			ForceCommit: false,
		}},
	}
}
